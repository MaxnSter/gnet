package pack

import "github.com/MaxnSter/gnet/iface"

var (
	packers = map[string]iface.Packer{}
)

//注册packer
func RegisterPacker(name string, p iface.Packer) {
	if _, ok := packers[name]; ok {
		panic("dup register packer, name :" + name)
	}

	packers[name] = p
}

// 获取制定名字对应的packer,若未注册,则panic
func MustGetPacker(name string) iface.Packer {
	if p, ok := packers[name]; ok {
		return p
	}

	panic("packer not register, name :" + name)
}
