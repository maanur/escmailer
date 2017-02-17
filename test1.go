package main

/*Это будет библиотечка COS-запросов к АСУЛОНу*/

import (
	"fmt"
	"log"
	"net"

	"github.com/Cristofori/kmud/telnet"
)

func main() {
	_, err := net.ResolveTCPAddr("tcp", "127.0.0.1:23")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.Dial("tcp", "127.0.0.1:23")
	if err == nil {
		fmt.Println("Connected...")
	} else {
		log.Fatal(err)
	}
	cache := telnet.NewTelnet(conn)
	for i := 0; i < 2; i++ {
		func() {
			what := make([]byte, 256)
			_, err = cache.Read(what)
			fmt.Println(string(what))
		}()
	}

}
