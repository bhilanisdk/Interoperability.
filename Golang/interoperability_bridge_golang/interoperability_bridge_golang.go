package interoperability_bridge_golang

/*
#cgo LDFLAGS: -L${SRCDIR}/.. -linteroperability_wrapper_uniffi_golang
#include <interoperability_bridge_golang.h>
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"runtime/cgo"
	"unsafe"
)

// This is needed, because as of go 1.24
// type RustBuffer C.RustBuffer cannot have methods,
// RustBuffer is treated as non-local type
type GoRustBuffer struct {
	inner C.RustBuffer
}

type RustBufferI interface {
	AsReader() *bytes.Reader
	Free()
	ToGoBytes() []byte
	Data() unsafe.Pointer
	Len() uint64
	Capacity() uint64
}

// C.RustBuffer fields exposed as an interface so they can be accessed in different Go packages.
// See https://github.com/golang/go/issues/13467
type ExternalCRustBuffer interface {
	Data() unsafe.Pointer
	Len() uint64
	Capacity() uint64
}

func RustBufferFromC(b C.RustBuffer) ExternalCRustBuffer {
	return GoRustBuffer{
		inner: b,
	}
}

func CFromRustBuffer(b ExternalCRustBuffer) C.RustBuffer {
	return C.RustBuffer{
		capacity: C.uint64_t(b.Capacity()),
		len:      C.uint64_t(b.Len()),
		data:     (*C.uchar)(b.Data()),
	}
}

func RustBufferFromExternal(b ExternalCRustBuffer) GoRustBuffer {
	return GoRustBuffer{
		inner: C.RustBuffer{
			capacity: C.uint64_t(b.Capacity()),
			len:      C.uint64_t(b.Len()),
			data:     (*C.uchar)(b.Data()),
		},
	}
}

func (cb GoRustBuffer) Capacity() uint64 {
	return uint64(cb.inner.capacity)
}

func (cb GoRustBuffer) Len() uint64 {
	return uint64(cb.inner.len)
}

func (cb GoRustBuffer) Data() unsafe.Pointer {
	return unsafe.Pointer(cb.inner.data)
}

func (cb GoRustBuffer) AsReader() *bytes.Reader {
	b := unsafe.Slice((*byte)(cb.inner.data), C.uint64_t(cb.inner.len))
	return bytes.NewReader(b)
}

func (cb GoRustBuffer) Free() {
	rustCall(func(status *C.RustCallStatus) bool {
		C.ffi_interoperability_bridge_golang_rustbuffer_free(cb.inner, status)
		return false
	})
}

func (cb GoRustBuffer) ToGoBytes() []byte {
	return C.GoBytes(unsafe.Pointer(cb.inner.data), C.int(cb.inner.len))
}

func stringToRustBuffer(str string) C.RustBuffer {
	return bytesToRustBuffer([]byte(str))
}

func bytesToRustBuffer(b []byte) C.RustBuffer {
	if len(b) == 0 {
		return C.RustBuffer{}
	}
	// We can pass the pointer along here, as it is pinned
	// for the duration of this call
	foreign := C.ForeignBytes{
		len:  C.int(len(b)),
		data: (*C.uchar)(unsafe.Pointer(&b[0])),
	}

	return rustCall(func(status *C.RustCallStatus) C.RustBuffer {
		return C.ffi_interoperability_bridge_golang_rustbuffer_from_bytes(foreign, status)
	})
}

type BufLifter[GoType any] interface {
	Lift(value RustBufferI) GoType
}

type BufLowerer[GoType any] interface {
	Lower(value GoType) C.RustBuffer
}

type BufReader[GoType any] interface {
	Read(reader io.Reader) GoType
}

type BufWriter[GoType any] interface {
	Write(writer io.Writer, value GoType)
}

func LowerIntoRustBuffer[GoType any](bufWriter BufWriter[GoType], value GoType) C.RustBuffer {
	// This might be not the most efficient way but it does not require knowing allocation size
	// beforehand
	var buffer bytes.Buffer
	bufWriter.Write(&buffer, value)

	bytes, err := io.ReadAll(&buffer)
	if err != nil {
		panic(fmt.Errorf("reading written data: %w", err))
	}
	return bytesToRustBuffer(bytes)
}

func LiftFromRustBuffer[GoType any](bufReader BufReader[GoType], rbuf RustBufferI) GoType {
	defer rbuf.Free()
	reader := rbuf.AsReader()
	item := bufReader.Read(reader)
	if reader.Len() > 0 {
		// TODO: Remove this
		leftover, _ := io.ReadAll(reader)
		panic(fmt.Errorf("Junk remaining in buffer after lifting: %s", string(leftover)))
	}
	return item
}

func rustCallWithError[E any, U any](converter BufReader[*E], callback func(*C.RustCallStatus) U) (U, *E) {
	var status C.RustCallStatus
	returnValue := callback(&status)
	err := checkCallStatus(converter, status)
	return returnValue, err
}

func checkCallStatus[E any](converter BufReader[*E], status C.RustCallStatus) *E {
	switch status.code {
	case 0:
		return nil
	case 1:
		return LiftFromRustBuffer(converter, GoRustBuffer{inner: status.errorBuf})
	case 2:
		// when the rust code sees a panic, it tries to construct a rustBuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(GoRustBuffer{inner: status.errorBuf})))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		panic(fmt.Errorf("unknown status code: %d", status.code))
	}
}

func checkCallStatusUnknown(status C.RustCallStatus) error {
	switch status.code {
	case 0:
		return nil
	case 1:
		panic(fmt.Errorf("function not returning an error returned an error"))
	case 2:
		// when the rust code sees a panic, it tries to construct a C.RustBuffer
		// with the message.  but if that code panics, then it just sends back
		// an empty buffer.
		if status.errorBuf.len > 0 {
			panic(fmt.Errorf("%s", FfiConverterStringINSTANCE.Lift(GoRustBuffer{
				inner: status.errorBuf,
			})))
		} else {
			panic(fmt.Errorf("Rust panicked while handling Rust panic"))
		}
	default:
		return fmt.Errorf("unknown status code: %d", status.code)
	}
}

func rustCall[U any](callback func(*C.RustCallStatus) U) U {
	returnValue, err := rustCallWithError[error](nil, callback)
	if err != nil {
		panic(err)
	}
	return returnValue
}

type NativeError interface {
	AsError() error
}

func writeInt8(writer io.Writer, value int8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint8(writer io.Writer, value uint8) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt16(writer io.Writer, value int16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint16(writer io.Writer, value uint16) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt32(writer io.Writer, value int32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint32(writer io.Writer, value uint32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeInt64(writer io.Writer, value int64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeUint64(writer io.Writer, value uint64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat32(writer io.Writer, value float32) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func writeFloat64(writer io.Writer, value float64) {
	if err := binary.Write(writer, binary.BigEndian, value); err != nil {
		panic(err)
	}
}

func readInt8(reader io.Reader) int8 {
	var result int8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint8(reader io.Reader) uint8 {
	var result uint8
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt16(reader io.Reader) int16 {
	var result int16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint16(reader io.Reader) uint16 {
	var result uint16
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt32(reader io.Reader) int32 {
	var result int32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint32(reader io.Reader) uint32 {
	var result uint32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readInt64(reader io.Reader) int64 {
	var result int64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readUint64(reader io.Reader) uint64 {
	var result uint64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat32(reader io.Reader) float32 {
	var result float32
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func readFloat64(reader io.Reader) float64 {
	var result float64
	if err := binary.Read(reader, binary.BigEndian, &result); err != nil {
		panic(err)
	}
	return result
}

func init() {

	uniffiCheckChecksums()
}

func uniffiCheckChecksums() {
	// Get the bindings contract version from our ComponentInterface
	bindingsContractVersion := 29
	// Get the scaffolding contract version by calling the into the dylib
	scaffoldingContractVersion := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint32_t {
		return C.ffi_interoperability_bridge_golang_uniffi_contract_version()
	})
	if bindingsContractVersion != int(scaffoldingContractVersion) {
		// If this happens try cleaning and rebuilding your project
		panic("interoperability_bridge_golang: UniFFI contract version mismatch")
	}
	{
		checksum := rustCall(func(_uniffiStatus *C.RustCallStatus) C.uint16_t {
			return C.uniffi_interoperability_bridge_golang_checksum_func_fetch_interoperability()
		})
		if checksum != 59692 {
			// If this happens try cleaning and rebuilding your project
			panic("interoperability_bridge_golang: uniffi_interoperability_bridge_golang_checksum_func_fetch_interoperability: UniFFI API checksum mismatch")
		}
	}
}

type FfiConverterUint32 struct{}

var FfiConverterUint32INSTANCE = FfiConverterUint32{}

func (FfiConverterUint32) Lower(value uint32) C.uint32_t {
	return C.uint32_t(value)
}

func (FfiConverterUint32) Write(writer io.Writer, value uint32) {
	writeUint32(writer, value)
}

func (FfiConverterUint32) Lift(value C.uint32_t) uint32 {
	return uint32(value)
}

func (FfiConverterUint32) Read(reader io.Reader) uint32 {
	return readUint32(reader)
}

type FfiDestroyerUint32 struct{}

func (FfiDestroyerUint32) Destroy(_ uint32) {}

type FfiConverterInt32 struct{}

var FfiConverterInt32INSTANCE = FfiConverterInt32{}

func (FfiConverterInt32) Lower(value int32) C.int32_t {
	return C.int32_t(value)
}

func (FfiConverterInt32) Write(writer io.Writer, value int32) {
	writeInt32(writer, value)
}

func (FfiConverterInt32) Lift(value C.int32_t) int32 {
	return int32(value)
}

func (FfiConverterInt32) Read(reader io.Reader) int32 {
	return readInt32(reader)
}

type FfiDestroyerInt32 struct{}

func (FfiDestroyerInt32) Destroy(_ int32) {}

type FfiConverterString struct{}

var FfiConverterStringINSTANCE = FfiConverterString{}

func (FfiConverterString) Lift(rb RustBufferI) string {
	defer rb.Free()
	reader := rb.AsReader()
	b, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Errorf("reading reader: %w", err))
	}
	return string(b)
}

func (FfiConverterString) Read(reader io.Reader) string {
	length := readInt32(reader)
	buffer := make([]byte, length)
	read_length, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		panic(err)
	}
	if read_length != int(length) {
		panic(fmt.Errorf("bad read length when reading string, expected %d, read %d", length, read_length))
	}
	return string(buffer)
}

func (FfiConverterString) Lower(value string) C.RustBuffer {
	return stringToRustBuffer(value)
}

func (c FfiConverterString) LowerExternal(value string) ExternalCRustBuffer {
	return RustBufferFromC(stringToRustBuffer(value))
}

func (FfiConverterString) Write(writer io.Writer, value string) {
	if len(value) > math.MaxInt32 {
		panic("String is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	write_length, err := io.WriteString(writer, value)
	if err != nil {
		panic(err)
	}
	if write_length != len(value) {
		panic(fmt.Errorf("bad write length when writing string, expected %d, written %d", len(value), write_length))
	}
}

type FfiDestroyerString struct{}

func (FfiDestroyerString) Destroy(_ string) {}

type FilterParams struct {
	Integration    *Integration
	Developmentkit *Developmentkit
	Language       *Language
	Crates         *Crates
	Page           *string
	Ids            *string
}

func (r *FilterParams) Destroy() {
	FfiDestroyerOptionalIntegration{}.Destroy(r.Integration)
	FfiDestroyerOptionalDevelopmentkit{}.Destroy(r.Developmentkit)
	FfiDestroyerOptionalLanguage{}.Destroy(r.Language)
	FfiDestroyerOptionalCrates{}.Destroy(r.Crates)
	FfiDestroyerOptionalString{}.Destroy(r.Page)
	FfiDestroyerOptionalString{}.Destroy(r.Ids)
}

type FfiConverterFilterParams struct{}

var FfiConverterFilterParamsINSTANCE = FfiConverterFilterParams{}

func (c FfiConverterFilterParams) Lift(rb RustBufferI) FilterParams {
	return LiftFromRustBuffer[FilterParams](c, rb)
}

func (c FfiConverterFilterParams) Read(reader io.Reader) FilterParams {
	return FilterParams{
		FfiConverterOptionalIntegrationINSTANCE.Read(reader),
		FfiConverterOptionalDevelopmentkitINSTANCE.Read(reader),
		FfiConverterOptionalLanguageINSTANCE.Read(reader),
		FfiConverterOptionalCratesINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterFilterParams) Lower(value FilterParams) C.RustBuffer {
	return LowerIntoRustBuffer[FilterParams](c, value)
}

func (c FfiConverterFilterParams) LowerExternal(value FilterParams) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[FilterParams](c, value))
}

func (c FfiConverterFilterParams) Write(writer io.Writer, value FilterParams) {
	FfiConverterOptionalIntegrationINSTANCE.Write(writer, value.Integration)
	FfiConverterOptionalDevelopmentkitINSTANCE.Write(writer, value.Developmentkit)
	FfiConverterOptionalLanguageINSTANCE.Write(writer, value.Language)
	FfiConverterOptionalCratesINSTANCE.Write(writer, value.Crates)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Page)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Ids)
}

type FfiDestroyerFilterParams struct{}

func (_ FfiDestroyerFilterParams) Destroy(value FilterParams) {
	value.Destroy()
}

type FilterResponse struct {
	Message    string
	Data       []Interoperability
	Pagination *PaginationMetadata
}

func (r *FilterResponse) Destroy() {
	FfiDestroyerString{}.Destroy(r.Message)
	FfiDestroyerSequenceInteroperability{}.Destroy(r.Data)
	FfiDestroyerOptionalPaginationMetadata{}.Destroy(r.Pagination)
}

type FfiConverterFilterResponse struct{}

var FfiConverterFilterResponseINSTANCE = FfiConverterFilterResponse{}

func (c FfiConverterFilterResponse) Lift(rb RustBufferI) FilterResponse {
	return LiftFromRustBuffer[FilterResponse](c, rb)
}

func (c FfiConverterFilterResponse) Read(reader io.Reader) FilterResponse {
	return FilterResponse{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterSequenceInteroperabilityINSTANCE.Read(reader),
		FfiConverterOptionalPaginationMetadataINSTANCE.Read(reader),
	}
}

func (c FfiConverterFilterResponse) Lower(value FilterResponse) C.RustBuffer {
	return LowerIntoRustBuffer[FilterResponse](c, value)
}

func (c FfiConverterFilterResponse) LowerExternal(value FilterResponse) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[FilterResponse](c, value))
}

func (c FfiConverterFilterResponse) Write(writer io.Writer, value FilterResponse) {
	FfiConverterStringINSTANCE.Write(writer, value.Message)
	FfiConverterSequenceInteroperabilityINSTANCE.Write(writer, value.Data)
	FfiConverterOptionalPaginationMetadataINSTANCE.Write(writer, value.Pagination)
}

type FfiDestroyerFilterResponse struct{}

func (_ FfiDestroyerFilterResponse) Destroy(value FilterResponse) {
	value.Destroy()
}

type Interoperability struct {
	Id             int32
	Title          string
	Crates         []Crates
	Language       Language
	Integration    Integration
	Developmentkit *[]Developmentkit
	Opensources    *[]Opensources
	Media          *[]string
}

func (r *Interoperability) Destroy() {
	FfiDestroyerInt32{}.Destroy(r.Id)
	FfiDestroyerString{}.Destroy(r.Title)
	FfiDestroyerSequenceCrates{}.Destroy(r.Crates)
	FfiDestroyerLanguage{}.Destroy(r.Language)
	FfiDestroyerIntegration{}.Destroy(r.Integration)
	FfiDestroyerOptionalSequenceDevelopmentkit{}.Destroy(r.Developmentkit)
	FfiDestroyerOptionalSequenceOpensources{}.Destroy(r.Opensources)
	FfiDestroyerOptionalSequenceString{}.Destroy(r.Media)
}

type FfiConverterInteroperability struct{}

var FfiConverterInteroperabilityINSTANCE = FfiConverterInteroperability{}

func (c FfiConverterInteroperability) Lift(rb RustBufferI) Interoperability {
	return LiftFromRustBuffer[Interoperability](c, rb)
}

func (c FfiConverterInteroperability) Read(reader io.Reader) Interoperability {
	return Interoperability{
		FfiConverterInt32INSTANCE.Read(reader),
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterSequenceCratesINSTANCE.Read(reader),
		FfiConverterLanguageINSTANCE.Read(reader),
		FfiConverterIntegrationINSTANCE.Read(reader),
		FfiConverterOptionalSequenceDevelopmentkitINSTANCE.Read(reader),
		FfiConverterOptionalSequenceOpensourcesINSTANCE.Read(reader),
		FfiConverterOptionalSequenceStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterInteroperability) Lower(value Interoperability) C.RustBuffer {
	return LowerIntoRustBuffer[Interoperability](c, value)
}

func (c FfiConverterInteroperability) LowerExternal(value Interoperability) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[Interoperability](c, value))
}

func (c FfiConverterInteroperability) Write(writer io.Writer, value Interoperability) {
	FfiConverterInt32INSTANCE.Write(writer, value.Id)
	FfiConverterStringINSTANCE.Write(writer, value.Title)
	FfiConverterSequenceCratesINSTANCE.Write(writer, value.Crates)
	FfiConverterLanguageINSTANCE.Write(writer, value.Language)
	FfiConverterIntegrationINSTANCE.Write(writer, value.Integration)
	FfiConverterOptionalSequenceDevelopmentkitINSTANCE.Write(writer, value.Developmentkit)
	FfiConverterOptionalSequenceOpensourcesINSTANCE.Write(writer, value.Opensources)
	FfiConverterOptionalSequenceStringINSTANCE.Write(writer, value.Media)
}

type FfiDestroyerInteroperability struct{}

func (_ FfiDestroyerInteroperability) Destroy(value Interoperability) {
	value.Destroy()
}

type Opensources struct {
	Name string
	Link *string
	Tags *[]string
}

func (r *Opensources) Destroy() {
	FfiDestroyerString{}.Destroy(r.Name)
	FfiDestroyerOptionalString{}.Destroy(r.Link)
	FfiDestroyerOptionalSequenceString{}.Destroy(r.Tags)
}

type FfiConverterOpensources struct{}

var FfiConverterOpensourcesINSTANCE = FfiConverterOpensources{}

func (c FfiConverterOpensources) Lift(rb RustBufferI) Opensources {
	return LiftFromRustBuffer[Opensources](c, rb)
}

func (c FfiConverterOpensources) Read(reader io.Reader) Opensources {
	return Opensources{
		FfiConverterStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalSequenceStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterOpensources) Lower(value Opensources) C.RustBuffer {
	return LowerIntoRustBuffer[Opensources](c, value)
}

func (c FfiConverterOpensources) LowerExternal(value Opensources) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[Opensources](c, value))
}

func (c FfiConverterOpensources) Write(writer io.Writer, value Opensources) {
	FfiConverterStringINSTANCE.Write(writer, value.Name)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.Link)
	FfiConverterOptionalSequenceStringINSTANCE.Write(writer, value.Tags)
}

type FfiDestroyerOpensources struct{}

func (_ FfiDestroyerOpensources) Destroy(value Opensources) {
	value.Destroy()
}

type PaginationMetadata struct {
	CurrentPage  uint32
	ItemsPerPage uint32
	TotalPages   uint32
	TotalItems   uint32
	NextPageUrl  *string
	PrevPageUrl  *string
	FirstPageUrl *string
	LastPageUrl  *string
}

func (r *PaginationMetadata) Destroy() {
	FfiDestroyerUint32{}.Destroy(r.CurrentPage)
	FfiDestroyerUint32{}.Destroy(r.ItemsPerPage)
	FfiDestroyerUint32{}.Destroy(r.TotalPages)
	FfiDestroyerUint32{}.Destroy(r.TotalItems)
	FfiDestroyerOptionalString{}.Destroy(r.NextPageUrl)
	FfiDestroyerOptionalString{}.Destroy(r.PrevPageUrl)
	FfiDestroyerOptionalString{}.Destroy(r.FirstPageUrl)
	FfiDestroyerOptionalString{}.Destroy(r.LastPageUrl)
}

type FfiConverterPaginationMetadata struct{}

var FfiConverterPaginationMetadataINSTANCE = FfiConverterPaginationMetadata{}

func (c FfiConverterPaginationMetadata) Lift(rb RustBufferI) PaginationMetadata {
	return LiftFromRustBuffer[PaginationMetadata](c, rb)
}

func (c FfiConverterPaginationMetadata) Read(reader io.Reader) PaginationMetadata {
	return PaginationMetadata{
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterUint32INSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
		FfiConverterOptionalStringINSTANCE.Read(reader),
	}
}

func (c FfiConverterPaginationMetadata) Lower(value PaginationMetadata) C.RustBuffer {
	return LowerIntoRustBuffer[PaginationMetadata](c, value)
}

func (c FfiConverterPaginationMetadata) LowerExternal(value PaginationMetadata) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[PaginationMetadata](c, value))
}

func (c FfiConverterPaginationMetadata) Write(writer io.Writer, value PaginationMetadata) {
	FfiConverterUint32INSTANCE.Write(writer, value.CurrentPage)
	FfiConverterUint32INSTANCE.Write(writer, value.ItemsPerPage)
	FfiConverterUint32INSTANCE.Write(writer, value.TotalPages)
	FfiConverterUint32INSTANCE.Write(writer, value.TotalItems)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.NextPageUrl)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.PrevPageUrl)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.FirstPageUrl)
	FfiConverterOptionalStringINSTANCE.Write(writer, value.LastPageUrl)
}

type FfiDestroyerPaginationMetadata struct{}

func (_ FfiDestroyerPaginationMetadata) Destroy(value PaginationMetadata) {
	value.Destroy()
}

type BridgeError struct {
	err error
}

// Convience method to turn *BridgeError into error
// Avoiding treating nil pointer as non nil error interface
func (err *BridgeError) AsError() error {
	if err == nil {
		return nil
	} else {
		return err
	}
}

func (err BridgeError) Error() string {
	return fmt.Sprintf("BridgeError: %s", err.err.Error())
}

func (err BridgeError) Unwrap() error {
	return err.err
}

// Err* are used for checking error type with `errors.Is`
var ErrBridgeErrorNetworkError = fmt.Errorf("BridgeErrorNetworkError")
var ErrBridgeErrorParsingError = fmt.Errorf("BridgeErrorParsingError")
var ErrBridgeErrorApiError = fmt.Errorf("BridgeErrorApiError")

// Variant structs
type BridgeErrorNetworkError struct {
	Field0 string
}

func NewBridgeErrorNetworkError(
	var0 string,
) *BridgeError {
	return &BridgeError{err: &BridgeErrorNetworkError{
		Field0: var0}}
}

func (e BridgeErrorNetworkError) destroy() {
	FfiDestroyerString{}.Destroy(e.Field0)
}

func (err BridgeErrorNetworkError) Error() string {
	return fmt.Sprint("NetworkError",
		": ",

		"Field0=",
		err.Field0,
	)
}

func (self BridgeErrorNetworkError) Is(target error) bool {
	return target == ErrBridgeErrorNetworkError
}

type BridgeErrorParsingError struct {
	Field0 string
}

func NewBridgeErrorParsingError(
	var0 string,
) *BridgeError {
	return &BridgeError{err: &BridgeErrorParsingError{
		Field0: var0}}
}

func (e BridgeErrorParsingError) destroy() {
	FfiDestroyerString{}.Destroy(e.Field0)
}

func (err BridgeErrorParsingError) Error() string {
	return fmt.Sprint("ParsingError",
		": ",

		"Field0=",
		err.Field0,
	)
}

func (self BridgeErrorParsingError) Is(target error) bool {
	return target == ErrBridgeErrorParsingError
}

type BridgeErrorApiError struct {
	Field0 string
}

func NewBridgeErrorApiError(
	var0 string,
) *BridgeError {
	return &BridgeError{err: &BridgeErrorApiError{
		Field0: var0}}
}

func (e BridgeErrorApiError) destroy() {
	FfiDestroyerString{}.Destroy(e.Field0)
}

func (err BridgeErrorApiError) Error() string {
	return fmt.Sprint("ApiError",
		": ",

		"Field0=",
		err.Field0,
	)
}

func (self BridgeErrorApiError) Is(target error) bool {
	return target == ErrBridgeErrorApiError
}

type FfiConverterBridgeError struct{}

var FfiConverterBridgeErrorINSTANCE = FfiConverterBridgeError{}

func (c FfiConverterBridgeError) Lift(eb RustBufferI) *BridgeError {
	return LiftFromRustBuffer[*BridgeError](c, eb)
}

func (c FfiConverterBridgeError) Lower(value *BridgeError) C.RustBuffer {
	return LowerIntoRustBuffer[*BridgeError](c, value)
}

func (c FfiConverterBridgeError) LowerExternal(value *BridgeError) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[*BridgeError](c, value))
}

func (c FfiConverterBridgeError) Read(reader io.Reader) *BridgeError {
	errorID := readUint32(reader)

	switch errorID {
	case 1:
		return &BridgeError{&BridgeErrorNetworkError{
			Field0: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 2:
		return &BridgeError{&BridgeErrorParsingError{
			Field0: FfiConverterStringINSTANCE.Read(reader),
		}}
	case 3:
		return &BridgeError{&BridgeErrorApiError{
			Field0: FfiConverterStringINSTANCE.Read(reader),
		}}
	default:
		panic(fmt.Sprintf("Unknown error code %d in FfiConverterBridgeError.Read()", errorID))
	}
}

func (c FfiConverterBridgeError) Write(writer io.Writer, value *BridgeError) {
	switch variantValue := value.err.(type) {
	case *BridgeErrorNetworkError:
		writeInt32(writer, 1)
		FfiConverterStringINSTANCE.Write(writer, variantValue.Field0)
	case *BridgeErrorParsingError:
		writeInt32(writer, 2)
		FfiConverterStringINSTANCE.Write(writer, variantValue.Field0)
	case *BridgeErrorApiError:
		writeInt32(writer, 3)
		FfiConverterStringINSTANCE.Write(writer, variantValue.Field0)
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiConverterBridgeError.Write", value))
	}
}

type FfiDestroyerBridgeError struct{}

func (_ FfiDestroyerBridgeError) Destroy(value *BridgeError) {
	switch variantValue := value.err.(type) {
	case BridgeErrorNetworkError:
		variantValue.destroy()
	case BridgeErrorParsingError:
		variantValue.destroy()
	case BridgeErrorApiError:
		variantValue.destroy()
	default:
		_ = variantValue
		panic(fmt.Sprintf("invalid error value `%v` in FfiDestroyerBridgeError.Destroy", value))
	}
}

type Crates uint

const (
	CratesUniffi  Crates = 1
	CratesWasm    Crates = 2
	CratesNeon    Crates = 3
	CratesRustler Crates = 4
	CratesRobusta Crates = 5
	CratesFrb     Crates = 6
	CratesMagnus  Crates = 7
	CratesPyo3    Crates = 8
	CratesRnet    Crates = 9
)

type FfiConverterCrates struct{}

var FfiConverterCratesINSTANCE = FfiConverterCrates{}

func (c FfiConverterCrates) Lift(rb RustBufferI) Crates {
	return LiftFromRustBuffer[Crates](c, rb)
}

func (c FfiConverterCrates) Lower(value Crates) C.RustBuffer {
	return LowerIntoRustBuffer[Crates](c, value)
}

func (c FfiConverterCrates) LowerExternal(value Crates) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[Crates](c, value))
}
func (FfiConverterCrates) Read(reader io.Reader) Crates {
	id := readInt32(reader)
	return Crates(id)
}

func (FfiConverterCrates) Write(writer io.Writer, value Crates) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerCrates struct{}

func (_ FfiDestroyerCrates) Destroy(value Crates) {
}

type Developmentkit uint

const (
	DevelopmentkitCli Developmentkit = 1
	DevelopmentkitApp Developmentkit = 2
)

type FfiConverterDevelopmentkit struct{}

var FfiConverterDevelopmentkitINSTANCE = FfiConverterDevelopmentkit{}

func (c FfiConverterDevelopmentkit) Lift(rb RustBufferI) Developmentkit {
	return LiftFromRustBuffer[Developmentkit](c, rb)
}

func (c FfiConverterDevelopmentkit) Lower(value Developmentkit) C.RustBuffer {
	return LowerIntoRustBuffer[Developmentkit](c, value)
}

func (c FfiConverterDevelopmentkit) LowerExternal(value Developmentkit) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[Developmentkit](c, value))
}
func (FfiConverterDevelopmentkit) Read(reader io.Reader) Developmentkit {
	id := readInt32(reader)
	return Developmentkit(id)
}

func (FfiConverterDevelopmentkit) Write(writer io.Writer, value Developmentkit) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerDevelopmentkit struct{}

func (_ FfiDestroyerDevelopmentkit) Destroy(value Developmentkit) {
}

type Integration uint

const (
	IntegrationPending   Integration = 1
	IntegrationCompleted Integration = 2
)

type FfiConverterIntegration struct{}

var FfiConverterIntegrationINSTANCE = FfiConverterIntegration{}

func (c FfiConverterIntegration) Lift(rb RustBufferI) Integration {
	return LiftFromRustBuffer[Integration](c, rb)
}

func (c FfiConverterIntegration) Lower(value Integration) C.RustBuffer {
	return LowerIntoRustBuffer[Integration](c, value)
}

func (c FfiConverterIntegration) LowerExternal(value Integration) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[Integration](c, value))
}
func (FfiConverterIntegration) Read(reader io.Reader) Integration {
	id := readInt32(reader)
	return Integration(id)
}

func (FfiConverterIntegration) Write(writer io.Writer, value Integration) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerIntegration struct{}

func (_ FfiDestroyerIntegration) Destroy(value Integration) {
}

type Language uint

const (
	LanguageJvm    Language = 1
	LanguageSwift  Language = 2
	LanguageNode   Language = 3
	LanguageBeam   Language = 4
	LanguageDart   Language = 5
	LanguageGolang Language = 6
	LanguageRuby   Language = 7
	LanguagePython Language = 8
	LanguageNet    Language = 9
)

type FfiConverterLanguage struct{}

var FfiConverterLanguageINSTANCE = FfiConverterLanguage{}

func (c FfiConverterLanguage) Lift(rb RustBufferI) Language {
	return LiftFromRustBuffer[Language](c, rb)
}

func (c FfiConverterLanguage) Lower(value Language) C.RustBuffer {
	return LowerIntoRustBuffer[Language](c, value)
}

func (c FfiConverterLanguage) LowerExternal(value Language) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[Language](c, value))
}
func (FfiConverterLanguage) Read(reader io.Reader) Language {
	id := readInt32(reader)
	return Language(id)
}

func (FfiConverterLanguage) Write(writer io.Writer, value Language) {
	writeInt32(writer, int32(value))
}

type FfiDestroyerLanguage struct{}

func (_ FfiDestroyerLanguage) Destroy(value Language) {
}

type FfiConverterOptionalString struct{}

var FfiConverterOptionalStringINSTANCE = FfiConverterOptionalString{}

func (c FfiConverterOptionalString) Lift(rb RustBufferI) *string {
	return LiftFromRustBuffer[*string](c, rb)
}

func (_ FfiConverterOptionalString) Read(reader io.Reader) *string {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterStringINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalString) Lower(value *string) C.RustBuffer {
	return LowerIntoRustBuffer[*string](c, value)
}

func (c FfiConverterOptionalString) LowerExternal(value *string) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[*string](c, value))
}

func (_ FfiConverterOptionalString) Write(writer io.Writer, value *string) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterStringINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalString struct{}

func (_ FfiDestroyerOptionalString) Destroy(value *string) {
	if value != nil {
		FfiDestroyerString{}.Destroy(*value)
	}
}

type FfiConverterOptionalPaginationMetadata struct{}

var FfiConverterOptionalPaginationMetadataINSTANCE = FfiConverterOptionalPaginationMetadata{}

func (c FfiConverterOptionalPaginationMetadata) Lift(rb RustBufferI) *PaginationMetadata {
	return LiftFromRustBuffer[*PaginationMetadata](c, rb)
}

func (_ FfiConverterOptionalPaginationMetadata) Read(reader io.Reader) *PaginationMetadata {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterPaginationMetadataINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalPaginationMetadata) Lower(value *PaginationMetadata) C.RustBuffer {
	return LowerIntoRustBuffer[*PaginationMetadata](c, value)
}

func (c FfiConverterOptionalPaginationMetadata) LowerExternal(value *PaginationMetadata) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[*PaginationMetadata](c, value))
}

func (_ FfiConverterOptionalPaginationMetadata) Write(writer io.Writer, value *PaginationMetadata) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterPaginationMetadataINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalPaginationMetadata struct{}

func (_ FfiDestroyerOptionalPaginationMetadata) Destroy(value *PaginationMetadata) {
	if value != nil {
		FfiDestroyerPaginationMetadata{}.Destroy(*value)
	}
}

type FfiConverterOptionalCrates struct{}

var FfiConverterOptionalCratesINSTANCE = FfiConverterOptionalCrates{}

func (c FfiConverterOptionalCrates) Lift(rb RustBufferI) *Crates {
	return LiftFromRustBuffer[*Crates](c, rb)
}

func (_ FfiConverterOptionalCrates) Read(reader io.Reader) *Crates {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterCratesINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalCrates) Lower(value *Crates) C.RustBuffer {
	return LowerIntoRustBuffer[*Crates](c, value)
}

func (c FfiConverterOptionalCrates) LowerExternal(value *Crates) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[*Crates](c, value))
}

func (_ FfiConverterOptionalCrates) Write(writer io.Writer, value *Crates) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterCratesINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalCrates struct{}

func (_ FfiDestroyerOptionalCrates) Destroy(value *Crates) {
	if value != nil {
		FfiDestroyerCrates{}.Destroy(*value)
	}
}

type FfiConverterOptionalDevelopmentkit struct{}

var FfiConverterOptionalDevelopmentkitINSTANCE = FfiConverterOptionalDevelopmentkit{}

func (c FfiConverterOptionalDevelopmentkit) Lift(rb RustBufferI) *Developmentkit {
	return LiftFromRustBuffer[*Developmentkit](c, rb)
}

func (_ FfiConverterOptionalDevelopmentkit) Read(reader io.Reader) *Developmentkit {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterDevelopmentkitINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalDevelopmentkit) Lower(value *Developmentkit) C.RustBuffer {
	return LowerIntoRustBuffer[*Developmentkit](c, value)
}

func (c FfiConverterOptionalDevelopmentkit) LowerExternal(value *Developmentkit) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[*Developmentkit](c, value))
}

func (_ FfiConverterOptionalDevelopmentkit) Write(writer io.Writer, value *Developmentkit) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterDevelopmentkitINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalDevelopmentkit struct{}

func (_ FfiDestroyerOptionalDevelopmentkit) Destroy(value *Developmentkit) {
	if value != nil {
		FfiDestroyerDevelopmentkit{}.Destroy(*value)
	}
}

type FfiConverterOptionalIntegration struct{}

var FfiConverterOptionalIntegrationINSTANCE = FfiConverterOptionalIntegration{}

func (c FfiConverterOptionalIntegration) Lift(rb RustBufferI) *Integration {
	return LiftFromRustBuffer[*Integration](c, rb)
}

func (_ FfiConverterOptionalIntegration) Read(reader io.Reader) *Integration {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterIntegrationINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalIntegration) Lower(value *Integration) C.RustBuffer {
	return LowerIntoRustBuffer[*Integration](c, value)
}

func (c FfiConverterOptionalIntegration) LowerExternal(value *Integration) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[*Integration](c, value))
}

func (_ FfiConverterOptionalIntegration) Write(writer io.Writer, value *Integration) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterIntegrationINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalIntegration struct{}

func (_ FfiDestroyerOptionalIntegration) Destroy(value *Integration) {
	if value != nil {
		FfiDestroyerIntegration{}.Destroy(*value)
	}
}

type FfiConverterOptionalLanguage struct{}

var FfiConverterOptionalLanguageINSTANCE = FfiConverterOptionalLanguage{}

func (c FfiConverterOptionalLanguage) Lift(rb RustBufferI) *Language {
	return LiftFromRustBuffer[*Language](c, rb)
}

func (_ FfiConverterOptionalLanguage) Read(reader io.Reader) *Language {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterLanguageINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalLanguage) Lower(value *Language) C.RustBuffer {
	return LowerIntoRustBuffer[*Language](c, value)
}

func (c FfiConverterOptionalLanguage) LowerExternal(value *Language) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[*Language](c, value))
}

func (_ FfiConverterOptionalLanguage) Write(writer io.Writer, value *Language) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterLanguageINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalLanguage struct{}

func (_ FfiDestroyerOptionalLanguage) Destroy(value *Language) {
	if value != nil {
		FfiDestroyerLanguage{}.Destroy(*value)
	}
}

type FfiConverterOptionalSequenceString struct{}

var FfiConverterOptionalSequenceStringINSTANCE = FfiConverterOptionalSequenceString{}

func (c FfiConverterOptionalSequenceString) Lift(rb RustBufferI) *[]string {
	return LiftFromRustBuffer[*[]string](c, rb)
}

func (_ FfiConverterOptionalSequenceString) Read(reader io.Reader) *[]string {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterSequenceStringINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalSequenceString) Lower(value *[]string) C.RustBuffer {
	return LowerIntoRustBuffer[*[]string](c, value)
}

func (c FfiConverterOptionalSequenceString) LowerExternal(value *[]string) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[*[]string](c, value))
}

func (_ FfiConverterOptionalSequenceString) Write(writer io.Writer, value *[]string) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterSequenceStringINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalSequenceString struct{}

func (_ FfiDestroyerOptionalSequenceString) Destroy(value *[]string) {
	if value != nil {
		FfiDestroyerSequenceString{}.Destroy(*value)
	}
}

type FfiConverterOptionalSequenceOpensources struct{}

var FfiConverterOptionalSequenceOpensourcesINSTANCE = FfiConverterOptionalSequenceOpensources{}

func (c FfiConverterOptionalSequenceOpensources) Lift(rb RustBufferI) *[]Opensources {
	return LiftFromRustBuffer[*[]Opensources](c, rb)
}

func (_ FfiConverterOptionalSequenceOpensources) Read(reader io.Reader) *[]Opensources {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterSequenceOpensourcesINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalSequenceOpensources) Lower(value *[]Opensources) C.RustBuffer {
	return LowerIntoRustBuffer[*[]Opensources](c, value)
}

func (c FfiConverterOptionalSequenceOpensources) LowerExternal(value *[]Opensources) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[*[]Opensources](c, value))
}

func (_ FfiConverterOptionalSequenceOpensources) Write(writer io.Writer, value *[]Opensources) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterSequenceOpensourcesINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalSequenceOpensources struct{}

func (_ FfiDestroyerOptionalSequenceOpensources) Destroy(value *[]Opensources) {
	if value != nil {
		FfiDestroyerSequenceOpensources{}.Destroy(*value)
	}
}

type FfiConverterOptionalSequenceDevelopmentkit struct{}

var FfiConverterOptionalSequenceDevelopmentkitINSTANCE = FfiConverterOptionalSequenceDevelopmentkit{}

func (c FfiConverterOptionalSequenceDevelopmentkit) Lift(rb RustBufferI) *[]Developmentkit {
	return LiftFromRustBuffer[*[]Developmentkit](c, rb)
}

func (_ FfiConverterOptionalSequenceDevelopmentkit) Read(reader io.Reader) *[]Developmentkit {
	if readInt8(reader) == 0 {
		return nil
	}
	temp := FfiConverterSequenceDevelopmentkitINSTANCE.Read(reader)
	return &temp
}

func (c FfiConverterOptionalSequenceDevelopmentkit) Lower(value *[]Developmentkit) C.RustBuffer {
	return LowerIntoRustBuffer[*[]Developmentkit](c, value)
}

func (c FfiConverterOptionalSequenceDevelopmentkit) LowerExternal(value *[]Developmentkit) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[*[]Developmentkit](c, value))
}

func (_ FfiConverterOptionalSequenceDevelopmentkit) Write(writer io.Writer, value *[]Developmentkit) {
	if value == nil {
		writeInt8(writer, 0)
	} else {
		writeInt8(writer, 1)
		FfiConverterSequenceDevelopmentkitINSTANCE.Write(writer, *value)
	}
}

type FfiDestroyerOptionalSequenceDevelopmentkit struct{}

func (_ FfiDestroyerOptionalSequenceDevelopmentkit) Destroy(value *[]Developmentkit) {
	if value != nil {
		FfiDestroyerSequenceDevelopmentkit{}.Destroy(*value)
	}
}

type FfiConverterSequenceString struct{}

var FfiConverterSequenceStringINSTANCE = FfiConverterSequenceString{}

func (c FfiConverterSequenceString) Lift(rb RustBufferI) []string {
	return LiftFromRustBuffer[[]string](c, rb)
}

func (c FfiConverterSequenceString) Read(reader io.Reader) []string {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]string, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterStringINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceString) Lower(value []string) C.RustBuffer {
	return LowerIntoRustBuffer[[]string](c, value)
}

func (c FfiConverterSequenceString) LowerExternal(value []string) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[[]string](c, value))
}

func (c FfiConverterSequenceString) Write(writer io.Writer, value []string) {
	if len(value) > math.MaxInt32 {
		panic("[]string is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterStringINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceString struct{}

func (FfiDestroyerSequenceString) Destroy(sequence []string) {
	for _, value := range sequence {
		FfiDestroyerString{}.Destroy(value)
	}
}

type FfiConverterSequenceInteroperability struct{}

var FfiConverterSequenceInteroperabilityINSTANCE = FfiConverterSequenceInteroperability{}

func (c FfiConverterSequenceInteroperability) Lift(rb RustBufferI) []Interoperability {
	return LiftFromRustBuffer[[]Interoperability](c, rb)
}

func (c FfiConverterSequenceInteroperability) Read(reader io.Reader) []Interoperability {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Interoperability, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterInteroperabilityINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceInteroperability) Lower(value []Interoperability) C.RustBuffer {
	return LowerIntoRustBuffer[[]Interoperability](c, value)
}

func (c FfiConverterSequenceInteroperability) LowerExternal(value []Interoperability) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[[]Interoperability](c, value))
}

func (c FfiConverterSequenceInteroperability) Write(writer io.Writer, value []Interoperability) {
	if len(value) > math.MaxInt32 {
		panic("[]Interoperability is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterInteroperabilityINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceInteroperability struct{}

func (FfiDestroyerSequenceInteroperability) Destroy(sequence []Interoperability) {
	for _, value := range sequence {
		FfiDestroyerInteroperability{}.Destroy(value)
	}
}

type FfiConverterSequenceOpensources struct{}

var FfiConverterSequenceOpensourcesINSTANCE = FfiConverterSequenceOpensources{}

func (c FfiConverterSequenceOpensources) Lift(rb RustBufferI) []Opensources {
	return LiftFromRustBuffer[[]Opensources](c, rb)
}

func (c FfiConverterSequenceOpensources) Read(reader io.Reader) []Opensources {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Opensources, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterOpensourcesINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceOpensources) Lower(value []Opensources) C.RustBuffer {
	return LowerIntoRustBuffer[[]Opensources](c, value)
}

func (c FfiConverterSequenceOpensources) LowerExternal(value []Opensources) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[[]Opensources](c, value))
}

func (c FfiConverterSequenceOpensources) Write(writer io.Writer, value []Opensources) {
	if len(value) > math.MaxInt32 {
		panic("[]Opensources is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterOpensourcesINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceOpensources struct{}

func (FfiDestroyerSequenceOpensources) Destroy(sequence []Opensources) {
	for _, value := range sequence {
		FfiDestroyerOpensources{}.Destroy(value)
	}
}

type FfiConverterSequenceCrates struct{}

var FfiConverterSequenceCratesINSTANCE = FfiConverterSequenceCrates{}

func (c FfiConverterSequenceCrates) Lift(rb RustBufferI) []Crates {
	return LiftFromRustBuffer[[]Crates](c, rb)
}

func (c FfiConverterSequenceCrates) Read(reader io.Reader) []Crates {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Crates, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterCratesINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceCrates) Lower(value []Crates) C.RustBuffer {
	return LowerIntoRustBuffer[[]Crates](c, value)
}

func (c FfiConverterSequenceCrates) LowerExternal(value []Crates) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[[]Crates](c, value))
}

func (c FfiConverterSequenceCrates) Write(writer io.Writer, value []Crates) {
	if len(value) > math.MaxInt32 {
		panic("[]Crates is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterCratesINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceCrates struct{}

func (FfiDestroyerSequenceCrates) Destroy(sequence []Crates) {
	for _, value := range sequence {
		FfiDestroyerCrates{}.Destroy(value)
	}
}

type FfiConverterSequenceDevelopmentkit struct{}

var FfiConverterSequenceDevelopmentkitINSTANCE = FfiConverterSequenceDevelopmentkit{}

func (c FfiConverterSequenceDevelopmentkit) Lift(rb RustBufferI) []Developmentkit {
	return LiftFromRustBuffer[[]Developmentkit](c, rb)
}

func (c FfiConverterSequenceDevelopmentkit) Read(reader io.Reader) []Developmentkit {
	length := readInt32(reader)
	if length == 0 {
		return nil
	}
	result := make([]Developmentkit, 0, length)
	for i := int32(0); i < length; i++ {
		result = append(result, FfiConverterDevelopmentkitINSTANCE.Read(reader))
	}
	return result
}

func (c FfiConverterSequenceDevelopmentkit) Lower(value []Developmentkit) C.RustBuffer {
	return LowerIntoRustBuffer[[]Developmentkit](c, value)
}

func (c FfiConverterSequenceDevelopmentkit) LowerExternal(value []Developmentkit) ExternalCRustBuffer {
	return RustBufferFromC(LowerIntoRustBuffer[[]Developmentkit](c, value))
}

func (c FfiConverterSequenceDevelopmentkit) Write(writer io.Writer, value []Developmentkit) {
	if len(value) > math.MaxInt32 {
		panic("[]Developmentkit is too large to fit into Int32")
	}

	writeInt32(writer, int32(len(value)))
	for _, item := range value {
		FfiConverterDevelopmentkitINSTANCE.Write(writer, item)
	}
}

type FfiDestroyerSequenceDevelopmentkit struct{}

func (FfiDestroyerSequenceDevelopmentkit) Destroy(sequence []Developmentkit) {
	for _, value := range sequence {
		FfiDestroyerDevelopmentkit{}.Destroy(value)
	}
}

const (
	uniffiRustFuturePollReady      int8 = 0
	uniffiRustFuturePollMaybeReady int8 = 1
)

type rustFuturePollFunc func(C.uint64_t, C.UniffiRustFutureContinuationCallback, C.uint64_t)
type rustFutureCompleteFunc[T any] func(C.uint64_t, *C.RustCallStatus) T
type rustFutureFreeFunc func(C.uint64_t)

//export interoperability_bridge_golang_uniffiFutureContinuationCallback
func interoperability_bridge_golang_uniffiFutureContinuationCallback(data C.uint64_t, pollResult C.int8_t) {
	h := cgo.Handle(uintptr(data))
	waiter := h.Value().(chan int8)
	waiter <- int8(pollResult)
}

func uniffiRustCallAsync[E any, T any, F any](
	errConverter BufReader[*E],
	completeFunc rustFutureCompleteFunc[F],
	liftFunc func(F) T,
	rustFuture C.uint64_t,
	pollFunc rustFuturePollFunc,
	freeFunc rustFutureFreeFunc,
) (T, *E) {
	defer freeFunc(rustFuture)

	pollResult := int8(-1)
	waiter := make(chan int8, 1)

	chanHandle := cgo.NewHandle(waiter)
	defer chanHandle.Delete()

	for pollResult != uniffiRustFuturePollReady {
		pollFunc(
			rustFuture,
			(C.UniffiRustFutureContinuationCallback)(C.interoperability_bridge_golang_uniffiFutureContinuationCallback),
			C.uint64_t(chanHandle),
		)
		pollResult = <-waiter
	}

	var goValue T
	var ffiValue F
	var err *E

	ffiValue, err = rustCallWithError(errConverter, func(status *C.RustCallStatus) F {
		return completeFunc(rustFuture, status)
	})
	if err != nil {
		return goValue, err
	}
	return liftFunc(ffiValue), nil
}

//export interoperability_bridge_golang_uniffiFreeGorutine
func interoperability_bridge_golang_uniffiFreeGorutine(data C.uint64_t) {
	handle := cgo.Handle(uintptr(data))
	defer handle.Delete()

	guard := handle.Value().(chan struct{})
	guard <- struct{}{}
}

func FetchInteroperability(params FilterParams) (FilterResponse, error) {
	res, err := uniffiRustCallAsync[BridgeError](
		FfiConverterBridgeErrorINSTANCE,
		// completeFn
		func(handle C.uint64_t, status *C.RustCallStatus) RustBufferI {
			res := C.ffi_interoperability_bridge_golang_rust_future_complete_rust_buffer(handle, status)
			return GoRustBuffer{
				inner: res,
			}
		},
		// liftFn
		func(ffi RustBufferI) FilterResponse {
			return FfiConverterFilterResponseINSTANCE.Lift(ffi)
		},
		C.uniffi_interoperability_bridge_golang_fn_func_fetch_interoperability(FfiConverterFilterParamsINSTANCE.Lower(params)),
		// pollFn
		func(handle C.uint64_t, continuation C.UniffiRustFutureContinuationCallback, data C.uint64_t) {
			C.ffi_interoperability_bridge_golang_rust_future_poll_rust_buffer(handle, continuation, data)
		},
		// freeFn
		func(handle C.uint64_t) {
			C.ffi_interoperability_bridge_golang_rust_future_free_rust_buffer(handle)
		},
	)

	if err == nil {
		return res, nil
	}

	return res, err
}
