package gnet

import (
	"bytes"
	"testing"

	"gnet/codec"
	"gnet/codec/codec_json"
	_ "gnet/codec/codec_json"
	"gnet/message/protocol/message_demo"
	_ "gnet/message/protocol/message_demo"
	"gnet/pack"
	"gnet/pack/pack_type_length_value"
	_ "gnet/pack/pack_type_length_value"
)

func TestMessagePack(t *testing.T) {
	packer := pack.MustGetPacker(pack_type_length_value.TlvPackerName)
	coder := codec.MustGetCoder(codec_json.CoderJsonTypeName)
	message := message_demo.NewDemoMessage("123")

	b := new(bytes.Buffer)
	err := packer.Pack(b, coder, message)
	if err != nil {
		t.Fatal(err)
	}

	n, err := packer.Unpack(b, coder)
	if err != nil {
		t.Fatal(err)
	}

	msg, ok := n.(*message_demo.DemoMessage)
	if !ok {
		t.Fatal("type")
	}

	if msg.Val != "123" {
		t.Fatal("val")
	}

}
