package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUUIDWorker(t *testing.T) {
	w, err := NewUUIDWorker(0)
	assert.NotNil(t, w)
	assert.Nil(t, err)

	w, err = NewUUIDWorker(-1)
	assert.Nil(t, w)
	assert.NotNil(t, err)

	w, err = NewUUIDWorker(9999)
	assert.Nil(t, w)
	assert.NotNil(t, err)
}

func TestGetUUID(t *testing.T) {
	uuidNum := 100000
	uuidCh := make(chan int64, uuidNum)

	for i := 0; i < uuidNum; i++ {
		go func() {
			uuidCh <- GetUUID()
		}()
	}

	m := make(map[int64]struct{})
	for i := 0; i < uuidNum; i++ {
		uuid := <-uuidCh

		if _, ok := m[uuid]; ok {
			assert.Fail(t, "uuid repeat")
		} else {
			m[uuid] = struct{}{}
		}
	}
}
