gnet
--
 [![Build Status][3]][4] [![Go Report Card][5]][6] [![MIT licensed][11]][12] [![GoDoc][1]][2]

[1]: https://godoc.org/github.com/MaxnSter/gnet?status.svg
[2]: https://godoc.org/github.com/MaxnSter/gnet
[3]: https://travis-ci.org/MaxnSter/gnet.svg?branch=master
[4]: https://travis-ci.org/MaxnSter/gnet
[5]: https://goreportcard.com/badge/github.com/MaxnSter/gnet
[6]: https://goreportcard.com/report/github.com/MaxnSter/gnet
[11]: https://img.shields.io/badge/license-MIT-blue.svg
[12]: LICENSE

gnet是一个简单易用,高度自由,完全组件化,高扩展性的golang网络库

安装
--
go get -u github.com/MaxnSter/gnet

快速构建一个echo服务器
--
```go
package main
import (
	"github.com/MaxnSter/gnet"

	// 注册coder组件
	_ "github.com/MaxnSter/gnet/codec/codec_byte"

	// 注册packer组件
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_line"

	// 注册网络组件
	_ "github.com/MaxnSter/gnet/net/tcp"

	// 注册goroutine pool组件
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other"
)

func main() {
	// 指定封包组件packer,本处使用文本协议
	// 指定本次编解码组件coder
	// 指定并发模型
	// 最后,我们创建一个了module对象
	module := gnet.NewModule(gnet.WithPacker("line"), gnet.WithCoder("byte"),
		gnet.WithPool("sessionRaceOther"))

	// 传入业务逻辑回调,创建一个module控制器
	operator := gnet.NewOperator(onMessage)

	// 指定我们的网络组件,传入module和module控制器,我们创建了一个NetServer
	server := gnet.NewNetServer("tcp", "echo", module, operator)

	// 启动服务器,可以用nc,telnet,或client测试
	server.ListenAndServe("127.0.0.1:8000")
}

// 业务逻辑入口
func onMessage(ev gnet.Event) {
	ev.Session().Send(ev.Message())
}
```

gnet的设计
--
在学习和使用网络库时,常常看到有这样的问题:"请问支持protobuf吗","请问能提供x语言的客户端吗"等等诸如此类问题.
要么是网络库中编解码,封解包模块与其它模块耦合,使用者不熟悉框架情况下难以扩展.要么编解码封包方式已在网络库中写死,
使用者别无选择.对此,本人愚见如下:是否能有这样的一个网络库,用户根据自己的需求可以随意指定,扩展通信过程的code&Decode,
pack&Unpack,设置是业务逻辑的并发模型?各个模块尽可能保证无耦合,可以独立用于任何地方(不只是网络库).

gnet是怎么做的?gnet中所有组件均以隐式import的方式创建,各个组件互不干扰,互不知晓,最大限度保证无耦合.
gnet主要由三个模块构成:module,Operator,网络组件.使用者只需以下三步,便可在gnet的基础上"构建另一个网络库",业务方不用关心业务逻辑之外的任何事

    1."拆装module"(coder,packer,并发模型),
    2.注册"module控制器"(注册相关回调和hook),
    3.把module及其控制器"装载"至指定组件上,启动!

    
gnet提供的组件
--
- coder 

    byte(裸包),json(标准库),protobuf,msgpack
- packer

    line(文本协议),lv(length-value),tlv(type-length-value)
- pool

    poolRaceOther(单事件循环),poolRaceSelf(round-robin派发),poolNoRace(无状态服务)
- timer

    gnet内置高精度定时器,与pool组件结合使用
- 网络

    tcp, ...(更多组件后期增加)
    
- 回调及hook
    消息接收回调(onMessage),

gnet的并发模型
--


gnet示例
--

仍待完善的工作
--
1.统一,完整的error handing

2.统一配置文件

3.vgo包管理

4.性能测试&与主流框架的性能比对(因为本人硬件条件有限,不敢用本机测试和小水管云主机的测试数据下结论... 欢迎提供性能测试数据!)

5.消息派发过程的gc性能分析

目前展望
--
1.http网络组件支持

2.rpc组件支持

特别感谢
--
感谢[@davyxu](https://github.com/davyxu),gnet设计灵感来源于[cellnet](https://github.com/davyxu)的学习过程,并借鉴了cellnet的消息元模块

备注
--
此项目伴随作为本人由浅入深学习Go的过程,可能有很多不够成熟的设计.欢迎各位看官拍砖!