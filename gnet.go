package gnet

var (
	netServerCreator = map[string]func(string, Module, Operator) NetServer{}
	netClientCreator = map[string]func(string, Module, Operator) NetClient{}
)

// RegisterServerCreator 注册一个gnet server的Creator
func RegisterServerCreator(creatorName string, f func(string, Module, Operator) NetServer) {
	if _, ok := netServerCreator[creatorName]; ok {
		panic("duplicate server creator register:" + creatorName)
	} else {
		netServerCreator[creatorName] = f
	}
}

// RegisterClientCreator 注册一个gnet client的Creator
func RegisterClientCreator(creatorName string, f func(string, Module, Operator) NetClient) {
	if _, ok := netClientCreator[creatorName]; ok {
		panic("duplicate client creator register:" + creatorName)
	} else {
		netClientCreator[creatorName] = f
	}
}

// NewNetServer 通过指定的网络组件,以及Module和Operator,返回一个NetServer
// 若指定的网络组件未注册,则panic
func NewNetServer(network, name string, module Module, operator Operator) NetServer {
	checkValid(module, operator)
	if creator, ok := netServerCreator[network]; !ok {
		panic("network:" + network + " not register")
	} else {
		return creator(name, module, operator)
	}
}

// NewNetClient 通过指定的网络组件,以及Module和Operator,返回一个NetClient
// 若指定的网络组件未注册,则panic
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
