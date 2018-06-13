package gnet

import (
	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/message_pack"
	"github.com/MaxnSter/gnet/plugin"
	"github.com/MaxnSter/gnet/timer"
	"github.com/MaxnSter/gnet/worker_pool"

	_ "github.com/MaxnSter/gnet/codec/codec_byte"
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_line"
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"
)

type Module interface {
	SetPool(pool string)
	Pool() worker_pool.Pool

	SetCoder(coder string)
	Coder() codec.Coder

	SetPacker(packet string)
	Packer() message_pack.Packer

	SetRdPlugin(plugins ...plugin.PluginBeforeRead)
	RdPlugins() []plugin.PluginBeforeRead

	SetWrPlugin(plugins ...plugin.PluginBeforeWrite)
	WrPlugins() []plugin.PluginBeforeWrite

	SetTimer(pool worker_pool.Pool)
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

func WithPool(pool string) func(m *moduleWrapper) {
	return func(m *moduleWrapper) {
		m.SetPool(pool)
	}
}

func WithCoder(coder string) func(m *moduleWrapper) {
	return func(m *moduleWrapper) {
		m.SetCoder(coder)
	}
}

func WithPacker(packer string) func(m *moduleWrapper) {
	return func(m *moduleWrapper) {
		m.SetPacker(packer)
	}
}

func WithRdPlugins(plugins ...plugin.PluginBeforeRead) func(m *moduleWrapper) {
	return func(m *moduleWrapper) {
		m.SetRdPlugin(plugins...)
	}
}

func WithWrPlugins(plugins ...plugin.PluginBeforeWrite) func(m *moduleWrapper) {
	return func(m *moduleWrapper) {
		m.SetWrPlugin(plugins...)
	}
}

func NewModule(options ...func(m *moduleWrapper)) Module {
	m := &moduleWrapper{}
	for _, option := range options {
		option(m)
	}

	return m
}

func NewDefaultModule() Module {
	return NewModule(WithPacker("line"),
		WithCoder("byte"),
		WithPool("poolRaceOther"))
}

func (m *moduleWrapper) SetPool(pool string) {
	m.pool = worker_pool.MustGetWorkerPool(pool)
}

func (m *moduleWrapper) Pool() worker_pool.Pool {
	return m.pool
}

func (m *moduleWrapper) SetCoder(coder string) {
	m.coder = codec.MustGetCoder(coder)
}

func (m *moduleWrapper) Coder() codec.Coder {
	return m.coder
}

func (m *moduleWrapper) SetPacker(packer string) {
	m.packer = message_pack.MustGetPacker(packer)
}

func (m *moduleWrapper) Packer() message_pack.Packer {
	return m.packer
}

func (m *moduleWrapper) SetRdPlugin(plugins ...plugin.PluginBeforeRead) {
	m.rdPlugins = append(m.rdPlugins, plugins...)
}

func (m *moduleWrapper) RdPlugins() []plugin.PluginBeforeRead {
	return m.rdPlugins
}

func (m *moduleWrapper) SetWrPlugin(plugins ...plugin.PluginBeforeWrite) {
	m.wrPlugins = append(m.wrPlugins, plugins...)
}

func (m *moduleWrapper) WrPlugins() []plugin.PluginBeforeWrite {
	return m.wrPlugins
}

func (m *moduleWrapper) SetTimer(pool worker_pool.Pool) {
	m.timer = timer.NewTimerManager(pool)
}

func (m *moduleWrapper) Timer() timer.TimerManager {
	return m.timer
}
