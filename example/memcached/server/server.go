package server

import (
	"bytes"
	"strings"

	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/net"
	"github.com/MaxnSter/gnet/util"
	"github.com/sirupsen/logrus"

	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/pack/pack_text"
	_ "github.com/MaxnSter/gnet/worker/worker_session_race_self"
)

// MemcachedOption是每个命令对应的操作
type MemcachedOption func(command string, data *map[string]interface{}) (ret []byte, err error)

// 记录当前支持的命令以及对应操作
var memcachedOp map[string]MemcachedOption

// 默认的 memcached server
var defaultMCServer *memcachedServer

// RegisterMemcachedOption注册命令及其操作,调用者保证线程安全
// opName重复注册则panic
// NOTE:不区分大小写
func RegisterMemcacheOption(opName string, op MemcachedOption) {
	opName = strings.ToUpper(opName)
	if _, dup := memcachedOp[opName]; dup {
		panic("dup register op, name:" + opName)
	}

	memcachedOp[opName] = op
}

// 简单的memcachedServer
type memcachedServer struct {
	netServer *net.TcpServer
	data      map[string]interface{}
}

// NewMemcachedServer创建并返回一个memcachedServer
func NewMemcachedServer() *memcachedServer {
	return &memcachedServer{
		data: make(map[string]interface{}),
	}
}

// onEvent是client消息处理函数
func (ms *memcachedServer) onEvent(ev iface.Event) {
	msg := ev.Message().([]byte)
	idx := bytes.IndexByte(msg, ' ')

	//错误的命令格式
	if idx == -1 {
		ev.Session().Send(util.StringToBytes("ERROR END"))
		return
	}

	cmd := strings.ToUpper(string(msg[0 : idx+1]))
	fullCmd := util.BytesToString(msg)
	if op, found := memcachedOp[cmd]; !found {
		ev.Session().Send(util.StringToBytes("NOT FOUND END"))
		return
	} else {

		//使用的workerPool保证了线程安全
		ret, err := op(fullCmd, &ms.data)

		if err != nil {
			logger.WithFields(logrus.Fields{
				"cmd": fullCmd, "error": err,
			}).Errorln("")

			ev.Session().Send(util.StringToBytes("FAILED END"))
			return
		}

		ev.Session().Send(ret)
	}
}

// StartAndRun根据传入的endpoint启动memcachedServer
func (ms *memcachedServer) StartAndRun(addr string) {
	ms.netServer = gnet.NewServer(addr, "memcached", ms.onEvent,
		gnet.WithCoder("byte"), gnet.WithPacker("text"), gnet.WithWorkerPool("poolRaceOther"))

	ms.netServer.StartAndRun()
}

func StartAndRun(addr string) {
	defaultMCServer.StartAndRun(addr)
}

func init() {
	defaultMCServer = NewMemcachedServer()
}
