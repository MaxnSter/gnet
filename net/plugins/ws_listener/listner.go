package ws_listener

import (
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"time"
)

type listener struct {
	conns chan conn
	w     websocket.Upgrader

	net.Listener
}

type conn struct {
	c   net.Conn
	err error
}

type wsConn struct {
	raw         *websocket.Conn
	messageType int
}

func (w *wsConn) Read(b []byte) (n int, err error) {
	messageType, r, err := w.raw.NextReader()
	if err != nil {
		return 0, err
	}

	w.messageType = messageType
	return r.Read(b)
}

func (w *wsConn) Write(b []byte) (n int, err error) {
	writer, err := w.raw.NextWriter(w.messageType)
	if err != nil {
		return 0, err
	}
	return writer.Write(b)
}

func (w *wsConn) Close() error {
	return w.raw.Close()
}

func (w *wsConn) LocalAddr() net.Addr {
	return w.raw.LocalAddr()
}

func (w *wsConn) RemoteAddr() net.Addr {
	return w.raw.RemoteAddr()
}

func (w *wsConn) SetDeadline(t time.Time) error {
	err := w.SetReadDeadline(t)
	if err != nil {
		return err
	}

	return w.SetWriteDeadline(t)
}

func (w *wsConn) SetReadDeadline(t time.Time) error {
	return w.raw.SetReadDeadline(t)
}

func (w *wsConn) SetWriteDeadline(t time.Time) error {
	return w.raw.SetWriteDeadline(t)
}

func New(addr, url string, w websocket.Upgrader) net.Listener {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	wsListener := &listener{
		w:        w,
		Listener: l,
		conns:    make(chan conn),
	}
	http.Handle(url, wsListener.wsHandler())
	return wsListener
}

func (l *listener) wsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, err := l.w.Upgrade(w, r, nil)
		l.conns <- conn{
			c: &wsConn{
				raw: raw,

				// todo
				messageType: websocket.TextMessage,
			},
			err: err,
		}
	}
}

func (l *listener) Accept() (net.Conn, error) {
	go func() {
		if err := http.Serve(l.Listener, nil); err != nil {
			panic(err)
		}
	}()

	c := <-l.conns
	return c.c, c.err
}

func (l *listener) Close() error {
	return l.Listener.Close()
}

func (l *listener) Addr() net.Addr {
	return l.Listener.Addr()
}
