package memcached_server

import (
	"bytes"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type mcDataMap map[string]*item

type Command interface {
	CmdName() string
	Operate(fullCmd string, mc mcDataMap, out *bytes.Buffer) (err error)
}

var pool = &sync.Pool{
	New: func() interface{} {
		return new(item)
	},
}

// GET command
// GET key
type cmdGet struct{}

func (c *cmdGet) CmdName() string {
	return "GET"
}

func (c *cmdGet) Operate(fullCmd string, mc mcDataMap, out *bytes.Buffer) (err error) {
	key := strings.Split(fullCmd, " ")[1]

	if item, ok := mc[key]; !ok {
		_, err = out.WriteString("NOTFOUND")
	} else {
		_, err = out.WriteString("VALUE ")
		if err != nil {
			return err
		}

		_, err = out.Write(item.Data)

	}

	return nil
}

// SET command
// SET key flags exptime byteLen
type cmdSet struct{}

func (c *cmdSet) CmdName() string {
	return "SET"
}

func (c *cmdSet) Operate(fullCmd string, mc mcDataMap, out *bytes.Buffer) (err error) {
	fullCmd = strings.Replace(fullCmd, "\r\n", "", -1)
	subCmds := strings.Split(fullCmd, " ")
	if len(subCmds) < 5 {
		return errors.New("invalid command")
	}

	key := subCmds[1]
	flags, err := strconv.Atoi(subCmds[2])
	exptime, err := strconv.Atoi(subCmds[3])
	byteLen, err := strconv.Atoi(subCmds[4])
	if err != nil {
		return err
	}

	item := pool.Get().(*item)
	item.Flag = flags
	item.Exptime = exptime
	item.Data = make([]byte, byteLen)

	if _, ok := mc[key]; ok {
		pool.Put(mc[key])
		mc[key] = nil
	}
	mc[key] = item

	out.WriteString("STORE")
	return nil
}

func init() {
	cmdG := &cmdGet{}
	cmdS := &cmdSet{}

	RegisterMcOption(cmdG.CmdName(), cmdG)
	RegisterMcOption(cmdS.CmdName(), cmdS)
}
