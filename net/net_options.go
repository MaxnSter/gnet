package net

import (
	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/pack"
)

type NetOptions struct {
	Coder codec.Coder
	Packer pack.Packer

	//TODO more, such as socket options...
}

type NetOpFunc func(options *NetOptions)
