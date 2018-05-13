package codec

import "github.com/MaxnSter/gnet/iface"

var (
	coders = map[string]iface.Coder{}
)

func RegisterCoder(name string, c iface.Coder) {
	if _, ok := coders[name]; ok {
		panic("duplicate register Coder :" + name)
	}

	coders[name] = c
}

func MustGetCoder(name string) iface.Coder {
	if c, ok := coders[name]; ok {
		return c
	} else {
		panic("Coder not register, name :" + name)
	}
}
