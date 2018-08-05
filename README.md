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

快速构建一个chat服务器
--
```go
package main

import (
	"fmt"
	"net"

	"github.com/MaxnSter/gnet"

	//注册我们使用的组件
	_ "github.com/MaxnSter/gnet/codec/codec_byte"                      //注册编解码组件
	_ "github.com/MaxnSter/gnet/message_pack/pack/pack_line"           //注册封解包组件
	_ "github.com/MaxnSter/gnet/net/tcp"                               //注册网络组件
	_ "github.com/MaxnSter/gnet/worker_pool/worker_session_race_other" //注册并发池组件
)

// 加入聊天组
func onConnected(s gnet.NetSession) {
	msg := fmt.Sprintf("user:%s, join", s.Raw().(net.Conn).RemoteAddr())

	//广播有用户加入
	s.AccessManager().Broadcast(func(session gnet.NetSession) {
		session.Send(msg)
	})
}

// 离开聊天组
func onClose(s gnet.NetSession) {
	msg := fmt.Sprintf("user:%s, leave", s.Raw().(net.Conn).RemoteAddr())

	//广播有用户离开
	s.AccessManager().Broadcast(func(session gnet.NetSession) {
		session.Send(msg)
	})
}

// 接收聊天消息
func onMessage(ev gnet.Event) {
	s := ev.Session()
	msg := fmt.Sprintf("user:%s, talk:%s", s.Raw().(net.Conn).RemoteAddr(),
		ev.Message())

	// 广播聊天消息给所有用户
	s.AccessManager().Broadcast(func(session gnet.NetSession) {
		session.Send(msg)
	})
}

func main() {
	//使用裸包编解码方式,文本协议的封解包方式,单EventLoop的并发模型,创建一个module
	module := gnet.NewModule(gnet.WithCoder("byte"), gnet.WithPacker("line"),
		gnet.WithPool("poolRaceOther"))

	//创建module控制器,注册相关回调
	operator := gnet.NewOperator(onMessage)
	operator.SetOnConnected(onConnected)
	operator.SetOnClose(onClose)

	//使用tcp网络组件,传入module及其控制器,创建我们的gnet server
	//启动成功后,使用telnet或netcat测试
	server := gnet.NewNetServer("tcp", "chat", module, operator)
	server.ListenAndServe("127.0.0.1:8000")
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
    3.把module及其控制器"装载"至指定网络组件上,启动!

    
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

    消息接收回调,连接建立或断开回调
	通过添加hook,调用方可以实现packer,coder,消息元的运行时改动,甚至改变读写对象

gnet的并发模型
--


gnet示例
--

仍待完善的工作
--
1.统一,完整的error handing

2.统一配置文件

3.vgo包管理

4.性能测试&与主流框架的性能比对

5.消息派发过程的gc性能分析

目前展望
--

1.http网络组件支持

22web socket网络组件支持

3.rpc组件支持

特别感谢
--

备注
--
