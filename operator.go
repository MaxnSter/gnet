package gnet

import (
	"github.com/MaxnSter/gnet/meta"
	"github.com/MaxnSter/gnet/packer/plugins/packer_type_length_value"
	"github.com/MaxnSter/gnet/pool"
	"io"
)

type Callback struct {
	OnSession     func(NetSession)
	OnMessage     func(Event)
	OnSessionStop func(NetSession)
}

type ReadInterceptor struct {
	PreRead  func(r io.Reader, m meta.Meta) (io.Reader, meta.Meta)
	InRead   func(buf []byte, m meta.Meta) ([]byte, meta.Meta)
	PostRead func(msg interface{}) interface{}
}

type WriteInterceptor struct {
	PreWrite func(w io.Writer, msg interface{}) (io.Writer, interface{})
	InWrite  func(w io.Writer, buf []byte) (io.Writer, []byte)
}

type Operator interface {
	PostEvent(ev Event)
	Read(reader io.Reader) (interface{}, error)
	Write(writer io.Writer, msg interface{}) error

	GetCallback() Callback
}

type operatorWrapper struct {
	Module
	Callback
	ReadInterceptor
	WriteInterceptor

	meta meta.Meta
}

func NewOperator(m Module, cb Callback, opts ...func(Operator)) Operator {
	s := &operatorWrapper{
		Module:   m,
		Callback: cb,
		meta:     nil,
	}

	for _, f := range opts {
		f(s)
	}
	return s
}

func WithReadHooks(i ReadInterceptor) func(Operator) {
	return func(operator Operator) {
		operator.(*operatorWrapper).ReadInterceptor = i
	}
}

func WithWriteHooks(i WriteInterceptor) func(Operator) {
	return func(operator Operator) {
		operator.(*operatorWrapper).WriteInterceptor = i
	}
}

func WithMeta(m meta.Meta) func(Operator) {
	return func(operator Operator) {
		operator.(*operatorWrapper).meta = m
	}
}

func (s *operatorWrapper) GetCallback() Callback {
	return s.Callback
}

func (s *operatorWrapper) PostEvent(ev Event) {
	if s.OnMessage != nil {
		s.Pool().Put(func() {
			s.OnMessage(ev)
		}, pool.WithIdentify(ev.Session().(interface{ ID() uint64 })))
	}
}

func (s *operatorWrapper) Read(reader io.Reader) (interface{}, error) {
	r, m := reader, s.meta
	if s.PreRead != nil {
		r, m = s.PreRead(r, m)
	}

	buf, err := s.Packer().Unpack(reader)
	if err != nil {
		return nil, err
	}
	if s.Packer().String() == packer_type_length_value.Name {
		var msgId uint32
		msgId, buf = packer_type_length_value.UnpackMsgId(buf)

		m = meta.MustGetMsgMeta(msgId)
	}
	if s.InRead != nil {
		buf, m = s.InRead(buf, m)
	}

	var msg interface{}
	if m != nil {
		msg = m.New()
	}
	if err = s.Coder().Decode(buf, msg); err != nil {
		return nil, err
	}
	if s.PostRead != nil {
		msg = s.PostRead(msg)
	}

	return msg, nil
}

func (s *operatorWrapper) Write(writer io.Writer, msg interface{}) error {
	if s.PreWrite != nil {
		writer, msg = s.PreWrite(writer, msg)
	}

	buf, err := s.Coder().Encode(msg)
	if err != nil {
		return err
	}

	if s.InWrite != nil {
		writer, buf = s.InWrite(writer, buf)
	}

	if s.Packer().String() == packer_type_length_value.Name {
		msgId := msg.(meta.Meta).Identify()
		buf = packer_type_length_value.PackMsgId(msgId, buf)
	}
	return s.Packer().Pack(writer, buf)
}
