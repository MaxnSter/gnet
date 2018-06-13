package message_pack

var (
	packers = map[string]Packer{}
)

//注册packer
func RegisterPacker(name string, p Packer) {
	if _, ok := packers[name]; ok {
		panic("dup register packer, name :" + name)
	}

	packers[name] = p
}

// 获取制定名字对应的packer,若未注册,则panic
func MustGetPacker(name string) Packer {
	if p, ok := packers[name]; ok {
		return p
	}

	panic("packer not register, name :" + name)
}
