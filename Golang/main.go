package main

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	SDK "interoperability-sdk-golang/interoperability_bridge_golang"
)

type TimedResult struct {
	PageNum  int
	Response SDK.FilterResponse
	Error    error
	Duration int64
}

type GoSDKit struct {
	isLibLoaded bool
}

func NewGoSDKit() *GoSDKit {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Platform & Architecture Check
	isSupportedOs := os == "windows" || os == "darwin" || os == "linux"
	isSupportedArch := arch == "amd64" || arch == "arm64"

	loaded := false
	if isSupportedOs && isSupportedArch {
		loaded = true
	} else {
		fmt.Printf("Unsupported platform: %s (%s). Native features disabled.\n", os, arch)
	}

	return &GoSDKit{isLibLoaded: loaded}
}

func (sdk *GoSDKit) IsReady() bool {
	return sdk.isLibLoaded
}

func (sdk *GoSDKit) FetchPages(startPage, endPage int) []TimedResult {
	count := endPage - startPage + 1
	results := make([]TimedResult, count)
	var wg sync.WaitGroup

	strPtr := func(s string) *string { return &s }

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			pageNum := startPage + idx

			if !sdk.isLibLoaded {
				results[idx] = TimedResult{PageNum: pageNum, Error: fmt.Errorf("library not loaded")}
				return
			}

			// Random delay logic (50ms to 250ms)
			time.Sleep(time.Duration(rand.Intn(201)+50) * time.Millisecond)
			startTime := time.Now()

			// Timeout logic (5 seconds)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			type bridgeRes struct {
				res SDK.FilterResponse
				err error
			}
			done := make(chan bridgeRes, 1)

			go func() {
				params := SDK.FilterParams{Page: strPtr(fmt.Sprintf("%d", pageNum))}
				res, err := SDK.FetchInteroperability(params)
				done <- bridgeRes{res, err}
			}()

			var finalRes bridgeRes
			select {
			case <-ctx.Done():
				finalRes = bridgeRes{err: fmt.Errorf("timeout exceeded")}
			case r := <-done:
				finalRes = r
			}

			results[idx] = TimedResult{
				PageNum:  pageNum,
				Response: finalRes.res,
				Error:    finalRes.err,
				Duration: time.Since(startTime).Milliseconds(),
			}
		}(i)
	}

	wg.Wait()
	return results
}

func main() {
	sdk := NewGoSDKit()
	totalStart := time.Now()

	fmt.Println("--- Bhilani Interop SDK (Golang Concurrency) ---")

	if !sdk.IsReady() {
		fmt.Println("Abort: Native library not loaded for this platform.")
		return
	}

	results := sdk.FetchPages(1, 5)

	for _, tr := range results {
		if tr.Error != nil {
			fmt.Printf("Page %d: Failed (%v) [%dms]\n", tr.PageNum, tr.Error, tr.Duration)
			continue
		}

		res := tr.Response
		totalPages := 0
		if res.Pagination != nil {
			totalPages = int(res.Pagination.TotalPages)
		}

		if len(res.Data) == 0 || tr.PageNum > totalPages {
			fmt.Printf("Page %d: Success (No Data) [%dms]\n", tr.PageNum, tr.Duration)
		} else {
			fmt.Printf("Page %d: Success [%dms]\n", tr.PageNum, tr.Duration)
			for _, item := range res.Data {
				fmt.Printf("  - Title: %s\n", item.Title)
			}
		}
	}

	fmt.Printf("\nTotal session duration: %dms\n", time.Since(totalStart).Milliseconds())
}