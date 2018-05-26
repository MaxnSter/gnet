package memcached_server

import (
	"bytes"
	"strings"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/net"
	"github.com/sirupsen/logrus"

	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/pack/pack_text"
	_ "github.com/MaxnSter/gnet/worker/worker_session_norace"
	_ "github.com/MaxnSter/gnet/worker/worker_session_race_other"
	_ "github.com/MaxnSter/gnet/worker/worker_session_race_self"
)

// 记录当前支持的命令以及对应操作
var memcachedCmd = map[string]Command{}

// 默认的 memcached memcached_server
var defaultMCServer *memcachedServer

// RegisterMemcachedOption注册命令及其操作,调用者保证线程安全
// opName重复注册则panic
// NOTE:不区分大小写
func RegisterMcOption(opName string, cmd Command) {
	opName = strings.ToUpper(opName)
	if _, dup := memcachedCmd[opName]; dup {
		panic("dup register op, name:" + opName)
	}

	memcachedCmd[opName] = cmd
}

// 简单的memcachedServer
type memcachedServer struct {
	netServer *net.TcpServer
	data      mcDataMap
	buf       *bytes.Buffer
	count     int
}

// NewMemcachedServer创建并返回一个memcachedServer
func NewMemcachedServer() *memcachedServer {
	return &memcachedServer{
		data: make(map[string]*item),
		buf:  new(bytes.Buffer),
	}
}

// onEvent是client消息处理函数
func (ms *memcachedServer) onEvent(ev iface.Event) {
	msg := ev.Message().([]byte)
	idx := bytes.IndexByte(msg, ' ')
	buf := ms.buf
	buf.Reset()

	//错误的命令格式
	if idx == -1 {
		buf.WriteString("ERROR END")
		ev.Session().Send(buf.Bytes())
		return
	}

	cmd := strings.ToUpper(string(msg[0:idx]))
	fullCmd := string(msg)

	if op, found := memcachedCmd[cmd]; !found {
		buf.WriteString("NOTFOUND")
	} else {

		//使用的workerPool保证了线程安全
		err := op.Operate(fullCmd, ms.data, buf)

		if err != nil {
			logger.WithFields(logrus.Fields{
				"cmd": fullCmd, "error": err,
			}).Errorln("")

			buf.Reset()
			buf.WriteString("ERROR")
		}

	}

	buf.WriteString(" END")
	ev.Session().Send(buf.Bytes())
}

// StartAndRun根据传入的endpoint启动memcachedServer
func (ms *memcachedServer) StartAndRun(addr string) {
	ms.netServer = gnet.NewServer(addr, "memcached", ms.onEvent,
		gnet.WithCoder("byte"),
		gnet.WithPacker("text"),
		gnet.WithWorkerPool("poolRaceOther"))

	ms.netServer.StartAndRun()
}

func StartAndRun(addr string) {
	defaultMCServer.StartAndRun(addr)
}

func init() {
	defaultMCServer = NewMemcachedServer()
}
