package ws

import (
	"github.com/MaxnSter/gnet"
	"github.com/MaxnSter/gnet/iface"
	"github.com/MaxnSter/gnet/logger"
	"github.com/MaxnSter/gnet/util"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var upgrader = websocket.Upgrader{}

type wsServer struct {
	wg       sync.WaitGroup
	sessions sync.Map

	l        net.Listener
	name     string
	module   gnet.Module
	operator gnet.Operator
}

func newWsServer(name string, m gnet.Module, o gnet.Operator) gnet.NetServer {
	return &wsServer{
		name:     name,
		sessions: sync.Map{},
		wg:       sync.WaitGroup{},
		module:   m,
		operator: o,
	}
}

// Serve启动服务器,调用方阻塞直到服务器关闭完成
// Serve必须在Listen成功后才可调用
func (ws *wsServer) Serve() {
	ws.wg.Add(1)
	ws.operator.StartModule(ws.module)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Warning(r)
			}

			ws.wg.Done()
		}()

		if err := http.Serve(ws.l, nil); err != nil {
			panic(err)
		}
	}()

	ws.wg.Wait()
	ws.operator.StopModule(ws.module)
}

// Listen开始监听指定的addr
func (ws *wsServer) Listen(addr string) error {
	http.HandleFunc("/", ws.serveHome)
	http.HandleFunc("/ws", ws.newSession)

	var err error
	if ws.l, err = net.Listen("tcp", addr); err != nil {
		panic(err)
	}

	return nil
}

func (ws *wsServer) serveHome(w http.ResponseWriter, r *http.Request) {
	logger.Debugln(r.URL)
	logger.Debugln(os.Getwd())
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

// webSocket链接开始
func (ws *wsServer) newSession(w http.ResponseWriter, r *http.Request) {
	sid := util.GetUUID()
	s := newWsSession(sid, ws.module, ws.operator,
		ws, w, r,
		func(s *wsSession) {
			ws.sessions.Delete(s.ID())
			ws.wg.Done()
		})

	if s == nil {
		// FIXME warning
		return
	}

	ws.wg.Add(1)
	ws.sessions.Store(sid, s)
}

// ListenAndServe监听指定addr并启动服务器
// 若Listen成功,则阻塞直到服务器关闭完成
// 若Listen失败,则panic
func (ws *wsServer) ListenAndServe(addr string) {
	if err := ws.Listen(addr); err != nil {
		panic(err)
	}

	ws.Serve()
}

// Stop停止服务器,调用方阻塞知道服务器所有模块关闭完成
func (ws *wsServer) Stop() {
	// 停止接受新连接
	ws.l.Close()

	// 关闭所有在线连接
	ws.Broadcast(func(s gnet.NetSession) {
		s.Stop()
	})
}

// BroadCast对所有NetSession连接执行fn
// 若module设置Pool,则fn全部投入Pool中,否则在当前goroutine执行
func (ws *wsServer) Broadcast(fn func(session gnet.NetSession)) {
	if ws.module.Pool() == nil {
		ws.sessions.Range(func(id, session interface{}) bool {
			fn(session.(gnet.NetSession))
			return true
		})
		return
	}

	// FIXME callback hell
	ws.sessions.Range(func(id, session interface{}) bool {
		ws.module.Pool().Put(session, func(ctx iface.Context) {
			fn(ctx.(gnet.NetSession))
		})
		return true
	})
}

// GetSession返回指定id对应的NetSession
func (ws *wsServer) GetSession(id int64) (gnet.NetSession, bool) {
	s, ok := ws.sessions.Load(id)
	if !ok {
		return nil, ok
	}

	return s.(gnet.NetSession), ok
}

// FIXME setSignal做为一个lib api
func (ws *wsServer) setSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGHUP)
	signal.Ignore(syscall.SIGPIPE)

	//监听信号
	go func() {
		sig := <-sigCh
		logger.WithField("signal", sig).Infoln("catch signal")

		ws.Stop()
	}()
}

func init() {
	gnet.RegisterServerCreator("ws", newWsServer)
}
