package gnet

var (
	netServerCreator = map[string]func(string, Module, Operator) NetServer{}
	netClientCreator = map[string]func(string, Module, Operator) NetClient{}
)

func RegisterServerCreator(creatorName string, f func(string, Module, Operator) NetServer) {
	if _, ok := netServerCreator[creatorName]; ok {
		panic("duplicate server creator register:" + creatorName)
	} else {
		netServerCreator[creatorName] = f
	}
}

func RegisterClientCreator(creatorName string, f func(string, Module, Operator) NetClient) {
	if _, ok := netClientCreator[creatorName]; ok {
		panic("duplicate client creator register:" + creatorName)
	} else {
		netClientCreator[creatorName] = f
	}
}

// only support tcp for now
func NewNetServer(network, name string, module Module, operator Operator) NetServer {
	checkValid(module, operator)
	if creator, ok := netServerCreator[network]; !ok {
		panic("network:" + network + " not register")
	} else {
		return creator(name, module, operator)
	}
}

func NewNetClient(network, name string, module Module, operator Operator) NetClient {
	checkValid(module, operator)
	if creator, ok := netClientCreator[network]; !ok {
		panic("network:" + network + " not register")
	} else {
		return creator(name, module, operator)
	}
}

func checkValid(m Module, o Operator) {
	if m.Coder() == nil || m.Packer() == nil {
		panic("coder and pack can not be nil")
	}

	if o.GetOnMessage() == nil {
		panic("onMessage can not be nil")
	}
}
