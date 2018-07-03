package message_pack

var (
	packers = map[string]Packer{}
)

// RegisterPacker 注册一个packer.
// 如果name已存在,则panic
func RegisterPacker(name string, p Packer) {
	if _, ok := packers[name]; ok {
		panic("dup register packer, name :" + name)
	}

	packers[name] = p
}

// MustGetPacker 获取指定名字对应的packer.
// 若未注册,则panic
func MustGetPacker(name string) Packer {
	if p, ok := packers[name]; ok {
		return p
	}

	panic("packer not register, name :" + name)
}
