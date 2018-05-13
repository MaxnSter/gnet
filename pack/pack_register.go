package pack

import "github.com/MaxnSter/gnet/iface"

var (
	packers = map[string]iface.Packer{}
)

func RegisterPacker(name string, p iface.Packer) {
	if _, ok := packers[name]; ok {
		panic("dup register packer, name :" + name)
	}

	packers[name] = p
}

func MustGetPacker(name string) iface.Packer {
	if p, ok := packers[name]; ok {
		return p
	}

	panic("packer not register, name :" + name)
}
