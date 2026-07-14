package main

import (
	"encoding/json"
	"fmt"
	SDK "interoperability-sdk-golang/interoperability_bridge_golang"
	"reflect"
	"strings"
)

// clean is the ultimate dynamic helper to handle single Enums, Slices, and Pointers
func clean(v interface{}) interface{} {
	rv := reflect.ValueOf(v)

	// 1. Handle Pointers (dereference)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	// 2. Handle Slices (recursive clean for each item)
	if rv.Kind() == reflect.Slice {
		var list []interface{}
		for i := 0; i < rv.Len(); i++ {
			list = append(list, clean(rv.Index(i).Interface()))
		}
		return list
	}

	// 3. Handle single Enums (strip prefixes automatically)
	s := fmt.Sprintf("%v", rv.Interface())
	s = strings.TrimPrefix(s, "Language")
	s = strings.TrimPrefix(s, "Integration")
	s = strings.TrimPrefix(s, "Crates")
	s = strings.TrimPrefix(s, "Developmentkit")
	
	// Handle nil/null cases from pointer dereferencing
	if s == "<nil>" {
		return nil
	}
	return s
}

func main() {
	fmt.Println("--- Bhilani Interop SDK (Golang) ---")

	// 1. Setup Parameters
	params := SDK.FilterParams{}

	// 2. Call the Rust Bridge
	result, err := SDK.FetchInteroperability(params)
	if err != nil {
		fmt.Printf("Error from Rust: %v\n", err)
		return
	}

	fmt.Println(result.Message)

	// 3. Map the data to include clean string names for Enums
	var response []map[string]interface{}
	for _, item := range result.Data {
		response = append(response, map[string]interface{}{
			"id":             item.Id,
			"title":          item.Title,
			"language":       clean(item.Language),
			"integration":    clean(item.Integration),
			"crates":         clean(item.Crates),
			"developmentkit": clean(item.Developmentkit),
			"opensources":    item.Opensources,
		})
	}

	// 4. Final Pretty-Print
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Printf("JSON Error: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}