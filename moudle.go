package gnet

import (
	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/packer"
	"github.com/MaxnSter/gnet/pool"
	"github.com/MaxnSter/gnet/timer"
)

type Runner interface {
	Run()
	Stop()
}

// Module 是gnet的所有组件集合
type Module interface {
	Pool() pool.Pool
	Coder() codec.Coder
	Packer() packer.Packer
}

type moduleWrapper struct {
	pool   pool.Pool
	coder  codec.Coder
	packer packer.Packer

	timer timer.Timer
}

func (m *moduleWrapper) Pool() pool.Pool {
	return m.pool
}

func (m *moduleWrapper) Coder() codec.Coder {
	return m.coder
}

func (m *moduleWrapper) Packer() packer.Packer {
	return m.packer
}

func NewModule(pool pool.Pool, c codec.Coder, packer packer.Packer) Module {
	return &moduleWrapper{
		pool:   pool,
		coder:  c,
		packer: packer,
	}
}
