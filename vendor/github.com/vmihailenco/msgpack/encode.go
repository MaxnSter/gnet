package msgpack

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"time"
)

type writer interface {
	io.Writer
	WriteByte(byte) error
	WriteString(string) (int, error)
}

type writeByte struct {
	io.Writer
}

func (w *writeByte) WriteByte(b byte) error {
	n, err := w.Write([]byte{b})
	if err != nil {
		return err
	}
	if n != 1 {
		return io.ErrShortWrite
	}
	return nil
}

func (w *writeByte) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func Marshal(v ...interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := NewEncoder(buf).Encode(v...)
	return buf.Bytes(), err
}

type Encoder struct {
	W writer
}

func NewEncoder(w io.Writer) *Encoder {
	ww, ok := w.(writer)
	if !ok {
		ww = &writeByte{Writer: w}
	}
	return &Encoder{
		W: ww,
	}
}

func (e *Encoder) Encode(v ...interface{}) error {
	for _, vv := range v {
		if err := e.encode(vv); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encode(iv interface{}) error {
	if iv == nil {
		return e.EncodeNil()
	}

	switch v := iv.(type) {
	case string:
		return e.EncodeString(v)
	case []byte:
		return e.EncodeBytes(v)
	case int:
		return e.EncodeInt64(int64(v))
	case int64:
		return e.EncodeInt64(v)
	case uint:
		return e.EncodeUint64(uint64(v))
	case uint64:
		return e.EncodeUint64(v)
	case bool:
		return e.EncodeBool(v)
	case float32:
		return e.EncodeFloat32(v)
	case float64:
		return e.EncodeFloat64(v)
	case []string:
		return e.encodeStringSlice(v)
	case map[string]string:
		return e.encodeMapStringString(v)
	case time.Duration:
		return e.EncodeInt64(int64(v))
	case time.Time:
		return e.EncodeTime(v)
	case encoder:
		return v.EncodeMsgpack(e.W)
	}
	return e.EncodeValue(reflect.ValueOf(iv))
}

func (e *Encoder) EncodeValue(v reflect.Value) error {
	switch v.Kind() {
	case reflect.String:
		return e.EncodeString(v.String())
	case reflect.Bool:
		return e.EncodeBool(v.Bool())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return e.EncodeUint64(v.Uint())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return e.EncodeInt64(v.Int())
	case reflect.Float32:
		return e.EncodeFloat32(float32(v.Float()))
	case reflect.Float64:
		return e.EncodeFloat64(v.Float())
	case reflect.Array:
		return e.encodeSlice(v)
	case reflect.Slice:
		if v.IsNil() {
			return e.EncodeNil()
		}
		return e.encodeSlice(v)
	case reflect.Map:
		return e.encodeMap(v)
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return e.EncodeNil()
		}
		if enc, ok := typEncMap[v.Type()]; ok {
			return enc(e, v)
		}
		if enc, ok := v.Interface().(encoder); ok {
			return enc.EncodeMsgpack(e.W)
		}
		return e.EncodeValue(v.Elem())
	case reflect.Struct:
		typ := v.Type()
		if enc, ok := typEncMap[typ]; ok {
			return enc(e, v)
		}
		if enc, ok := v.Interface().(encoder); ok {
			return enc.EncodeMsgpack(e.W)
		}
		return e.encodeStruct(v)
	default:
		return fmt.Errorf("msgpack: unsupported type %v", v.Type().String())
	}
	panic("not reached")
}

func (e *Encoder) EncodeNil() error {
	return e.W.WriteByte(nilCode)
}

func (e *Encoder) EncodeBool(value bool) error {
	if value {
		return e.W.WriteByte(trueCode)
	}
	return e.W.WriteByte(falseCode)
}

func (e *Encoder) encodeStruct(v reflect.Value) error {
	fields := structs.Fields(v.Type())
	switch l := len(fields); {
	case l < 16:
		if err := e.W.WriteByte(fixMapLowCode | byte(l)); err != nil {
			return err
		}
	case l < 65536:
		if err := e.write([]byte{
			map16Code,
			byte(l >> 8),
			byte(l),
		}); err != nil {
			return err
		}
	default:
		if err := e.write([]byte{
			map32Code,
			byte(l >> 24),
			byte(l >> 16),
			byte(l >> 8),
			byte(l),
		}); err != nil {
			return err
		}
	}
	for _, f := range fields {
		if err := e.EncodeString(f.Name()); err != nil {
			return err
		}
		if err := f.EncodeValue(e, v); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) write(data []byte) error {
	n, err := e.W.Write(data)
	if err != nil {
		return err
	}
	if n < len(data) {
		return io.ErrShortWrite
	}
	return nil
}

func (e *Encoder) writeString(s string) error {
	n, err := e.W.WriteString(s)
	if err != nil {
		return err
	}
	if n < len(s) {
		return io.ErrShortWrite
	}
	return nil
}
