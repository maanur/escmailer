package cache

import (
	"fmt"
)

type cachesrv struct {
	addr, login, pass string
}

type checker struct {
	name, dir, nsp string
	cmd            chan cachecmd
	sem            chan int
}

type cachecmd struct {
	cmd string // строка команды
	ret bool   // ожидаем ли ответ?
}

func newChecker(name, dir, nsp string) *checker {
	ch := new(checker)
	ch.name = name
	ch.dir = dir
	ch.nsp = nsp
	/*ch.cmd = make(chan cachecmd)
	ch.sem = make(chan int)*/
	return ch
}

func (ch *checker) start(srv cachesrv) {
	go func() {
		conn := DialCache(srv.addr, srv.login, srv.pass)
		conn.ChangeNsp(ch.nsp)
		for {
			cmd := <-ch.cmd
			if cmd.ret {
				fmt.Println(conn.Query(cmd.cmd))
			} else {
				conn.Command(cmd.cmd)
			}
		}
	}()
	ch.sem <- 1
}
