package util

import (
	"reflect"
	"unsafe"
)

// BytesToString convert []byte type to string type
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToBytes convert string type to []byte type
func StringToBytes(s string) []byte {
	sp := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bp := reflect.SliceHeader{Data: sp.Data, Len: sp.Len, Cap: sp.Len}
	return *(*[]byte)(unsafe.Pointer(&bp))
}
