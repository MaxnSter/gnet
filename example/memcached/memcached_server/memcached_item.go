package memcached_server

import (
	"reflect"
	"unsafe"
)

type item struct {
	Flag    int
	Exptime int
	Data    []byte
}

func NewItem(key string, flag int, expTime int, dataLen int) *item {
	return &item{
		Flag:    flag,
		Exptime: expTime,
		Data:    make([]byte, dataLen),
	}
}

func (im *item) Size() uintptr {
	dataHead := (*reflect.SliceHeader)(unsafe.Pointer(&im.Data))
	return unsafe.Sizeof(im.Flag) + unsafe.Sizeof(im.Exptime) +
		unsafe.Sizeof(dataHead.Len) + unsafe.Sizeof(dataHead.Cap) + unsafe.Sizeof(dataHead.Data) +
		uintptr(dataHead.Cap)
}
