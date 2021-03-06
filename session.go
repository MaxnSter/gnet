package gnet

import (
	"bufio"
	"github.com/MaxnSter/GolangDataStructure/try"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/MaxnSter/gnet/util"
)

type session struct {
	identify uint64
	rd       *bufio.Reader
	wr       *bufio.Writer
	raw      net.Conn
	wrQueue  *util.MsgQueue

	once    sync.Once
	closeCh chan struct{}
	grace   time.Duration

	guard    sync.Mutex
	priority map[string]interface{}

	manager  SessionManager
	operator Operator
}

func (s *session) ID() uint64 {
	return s.identify
}

func (s *session) LocalAddr() net.Addr {
	return s.raw.RemoteAddr()
}

func (s *session) RemoteAddr() net.Addr {
	return s.raw.RemoteAddr()
}

func (s *session) Send(message interface{}) {
	select {
	case <-s.closeCh:
		return
	default:
	}

	s.wrQueue.Put(message)
}

func (s *session) AccessManager() SessionManager {
	return s.manager
}

func (s *session) Stop() {
	select {
	case <-s.closeCh:
		return
	default:
	}

	s.once.Do(func() {
		close(s.closeCh)
		s.raw.SetDeadline(time.Now().Add(s.grace))
		s.raw.Close()
	})
}

func newSession(identify uint64, conn net.Conn, manager SessionManager,
	o Operator) NetSession {
	return &session{
		identify: identify,
		rd:       bufio.NewReader(conn),
		wr:       bufio.NewWriter(conn),
		raw:      conn,
		wrQueue:  util.NewMsgQueue(),
		closeCh:  make(chan struct{}),
		grace:    time.Second * 3,
		priority: map[string]interface{}{},
		manager:  manager,
		operator: o,
	}
}

// Run start session util session.close called
func (s *session) Run() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	if cb := s.operator.GetCallback().OnSession; cb != nil {
		cb(s)
	}

	go func() {
		s.readLoop()
		wg.Done()
	}()
	go func() {
		s.writeLoop()
		wg.Done()
	}()

	wg.Wait()

	if cb := s.operator.GetCallback().OnSessionStop; cb != nil {
		cb(s)
	}
}

func (s *session) readLoop() {
	readF := func() error {
		for {
			msg, err := s.operator.Read(s.rd)
			if err != nil {
				if err == io.EOF {
					return nil
				}

				if err, ok := err.(net.Error); ok && err.Timeout() {
					select {
					case <-s.closeCh:
						return nil
					default:
					}
				}

				return errors.Wrap(err, "read failed")
			}

			s.operator.PostEvent(&eventWrapper{eventSession: s, msg: msg})
		}
	}

	finish := func(err error) error {
		if err != nil {
			glog.Errorf("+%v", err)
		}

		s.Stop()
		return nil
	}

	try.Try(readF).Final(finish).Do()
}

func (s *session) writeLoop() {
	writeF := func() error {
		var items []interface{}
		for {
			s.wrQueue.PickWithSignal(s.closeCh, &items)
			if len(items) <= 0 {
				return nil
			}

			for i := 0; i < len(items); i++ {
				err := s.operator.Write(s.wr, items[i])
				if err != nil {
					s.wr.Flush()
					return errors.Wrap(err, "write failed")
				}
			}
			err := s.wr.Flush()
			if err != nil {
				return errors.Wrap(err, "flush writer error")
			}
			items = items[0:0]
		}
	}

	finish := func(err error) error {
		if err != nil {
			glog.Errorf("+%v", err)
		}

		s.Stop()
		return nil
	}

	try.Try(writeF).Final(finish).Do()
}
