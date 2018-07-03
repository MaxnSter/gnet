package gnet

import (
	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/message_pack"
	"github.com/MaxnSter/gnet/plugin"
	"github.com/MaxnSter/gnet/timer"
	"github.com/MaxnSter/gnet/worker_pool"
)

// Module 是gnet的所有组件集合
type Module interface {
	// SetPool设置一个注册过的goroutine pool组件,若组件未注册,则panic
	SetPool(pool string)
	// SetSharePool直接设置一个goroutine pool组件,所有持有本module的对象共享该pool
	SetSharePool(pool worker_pool.Pool) //sometimes we want all connections run in one pool
	// SetCoder设置一个注册过的coder组件,若组件未注册,则panic
	SetCoder(coder string)
	// SetPacker设置一个注册过的Packer组件,若组件未注册,则panic
	SetPacker(packet string)

	// Pool返回本module的pool组件
	Pool() worker_pool.Pool
	// COder返回本module的coder组件
	Coder() codec.Coder
	// Packer返回本module的packer组件
	Packer() message_pack.Packer

	// SetRdPlugin设置一个或多个inbound hook
	//TODO 插件热插拔
	SetRdPlugin(plugins ...plugin.PluginBeforeRead)
	// SetWrPlugin设置一个或多个outbound hook
	SetWrPlugin(plugins ...plugin.PluginBeforeWrite)

	// RdPlugins返回本module注册的所有inbound hooks
	RdPlugins() []plugin.PluginBeforeRead
	// WrPlugins返回本module注册的所有inbound hooks
	WrPlugins() []plugin.PluginBeforeWrite

	// SetTimer启用gnet内置定时器,指定pool用于dispatch onTimeout回调
	SetTimer(pool worker_pool.Pool)
	// Timer返回module中的timer组件,返回的timer无法使用
	// 注意:调用者不能用if Timer() == nil来判断是否注册过
	Timer() timer.TimerManager
}

type moduleWrapper struct {
	pool   worker_pool.Pool
	coder  codec.Coder
	packer message_pack.Packer
	timer  timer.TimerManager

	rdPlugins []plugin.PluginBeforeRead
	wrPlugins []plugin.PluginBeforeWrite
}

// WithPool 用于指定module的pool
func WithPool(pool string) func(m *moduleWrapper) {
	return func(m *moduleWrapper) {
		m.SetPool(pool)
	}
}

// WithCoder 用于指定module的coder
func WithCoder(coder string) func(m *moduleWrapper) {
	return func(m *moduleWrapper) {
		m.SetCoder(coder)
	}
}

// WithPacker 用于指定module的packer
func WithPacker(packer string) func(m *moduleWrapper) {
	return func(m *moduleWrapper) {
		m.SetPacker(packer)
	}
}

// WithRdPlugins 用于指定module的RdPlugins
func WithRdPlugins(plugins ...plugin.PluginBeforeRead) func(m *moduleWrapper) {
	return func(m *moduleWrapper) {
		m.SetRdPlugin(plugins...)
	}
}

// WithWrPlugins 用于指定module的WrPlugins
func WithWrPlugins(plugins ...plugin.PluginBeforeWrite) func(m *moduleWrapper) {
	return func(m *moduleWrapper) {
		m.SetWrPlugin(plugins...)
	}
}

// NewModule 传入0个或多个option,返回一个gnet module对象
func NewModule(options ...func(m *moduleWrapper)) Module {
	m := &moduleWrapper{}
	for _, option := range options {
		option(m)
	}

	return m
}

// SetPool设置一个注册过的goroutine pool组件,若组件未注册,则panic
func (m *moduleWrapper) SetPool(pool string) {
	m.pool = worker_pool.MustGetWorkerPool(pool)
}

// SetSharePool直接设置一个goroutine pool组件,所有持有本module的对象共享该pool
func (m *moduleWrapper) SetSharePool(pool worker_pool.Pool) {
	m.pool = pool
}

// Pool返回本module的pool组件
func (m *moduleWrapper) Pool() worker_pool.Pool {
	return m.pool
}

// SetCoder设置一个注册过的coder组件,若组件未注册,则panic
func (m *moduleWrapper) SetCoder(coder string) {
	m.coder = codec.MustGetCoder(coder)
}

// Coder返回本module的coder组件
func (m *moduleWrapper) Coder() codec.Coder {
	return m.coder
}

// SetPacker设置一个注册过的Packer组件,若组件未注册,则panic
func (m *moduleWrapper) SetPacker(packer string) {
	m.packer = message_pack.MustGetPacker(packer)
}

// Packer返回本module的packer组件
func (m *moduleWrapper) Packer() message_pack.Packer {
	return m.packer
}

// SetRdPlugin设置一个或多个inbound hook
func (m *moduleWrapper) SetRdPlugin(plugins ...plugin.PluginBeforeRead) {
	m.rdPlugins = append(m.rdPlugins, plugins...)
}

// RdPlugins返回本module注册的所有inbound hooks
func (m *moduleWrapper) RdPlugins() []plugin.PluginBeforeRead {
	return m.rdPlugins
}

// SetWrPlugin设置一个或多个outbound hook
func (m *moduleWrapper) SetWrPlugin(plugins ...plugin.PluginBeforeWrite) {
	m.wrPlugins = append(m.wrPlugins, plugins...)
}

// WrPlugins返回本module注册的所有inbound hooks
func (m *moduleWrapper) WrPlugins() []plugin.PluginBeforeWrite {
	return m.wrPlugins
}

// SetTimer启用gnet内置定时器,指定pool用于dispatch onTimeout回调
func (m *moduleWrapper) SetTimer(pool worker_pool.Pool) {
	m.timer = timer.NewTimerManager(pool)
}

// 注意:调用者不能用if Timer() == nil来判断是否注册过
func (m *moduleWrapper) Timer() timer.TimerManager {
	return m.timer
}
