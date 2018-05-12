package pack

var (
	packers = map[string]Packer{}
)

func RegisterPacker(name string, p Packer) {
	if _, ok := packers[name]; ok {
		panic("dup register packer, name :" + name)
	}

	packers[name] = p
}

func MustGetPacker(name string) Packer {
	if p, ok := packers[name]; ok {
		return p
	}

	panic("packer not register, name :" + name)
}
