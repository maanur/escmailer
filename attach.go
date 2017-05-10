package main

import (
	"github.com/jpoehls/gophermail"
)

type attach struct {
	name      string
	files     []string
	checkDir  string
	checkNsp  string
	checkFunc []string
}

func (msg escMsg) convAttach() []gophermail.Attachment {
	g := make([]gophermail.Attachment, len(msg.attach))
	for i := 0; i < len(msg.attach); i++ {
		checkAttach(msg.attach[i])
		g[i].Name = msg.attach[i].name
		g[i].Data = pack(msg.attach[i].files)
	}
	return g
}
