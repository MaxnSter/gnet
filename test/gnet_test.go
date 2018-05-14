package test

import (
	"bytes"
	"testing"

	"github.com/MaxnSter/gnet/codec"
	"github.com/MaxnSter/gnet/codec/codec_json"
	"github.com/MaxnSter/gnet/message/protocol/message_demo"
	"github.com/MaxnSter/gnet/pack"
	"github.com/MaxnSter/gnet/pack/pack_type_length_value"
)

func TestMessagePack(t *testing.T) {
	packer := pack.MustGetPacker(pack_type_length_value.TlvPackerName)
	coder := codec.MustGetCoder(codec_json.CoderJsonTypeName)
	message := message_demo.NewDemoMessage("12dsgnijgongoigasdlgnaerwoghisdognaeotraierhgoefnhao3")

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

	if msg.Val != "12dsgnijgongoigasdlgnaerwoghisdognaeotraierhgoefnhao3" {
		t.Fatal("val")
	}

}
