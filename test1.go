package main

/* Начну с отправки тествого сообщения */

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/jpoehls/gophermail"
)

func main() {
	m := readConf()
	fmt.Println(m)
	log.Fatal("Not yet")
}

type escMsg struct {
	// совокупность параметров для письма
	id     int      // идентификактор письма
	to     []string // кому
	cc     []string // копия
	bcc    []string // скрытая копия
	from   string   // от
	subj   string   // тема
	body   string
	attach []io.Reader // файлы в аттаче

}

func (m *escMsg) ready() []byte {
	r := new(gophermail.Message)
	for i := 0; i < len(m.to); i++ {
		err := r.AddTo(m.to[i])
		if err != nil {
			log.Fatal(err)
		}
	}
	for i := 0; i < len(m.cc); i++ {
		err := r.AddCc(m.cc[i])
		if err != nil {
			log.Fatal(err)
		}
	}
	for i := 0; i < len(m.bcc); i++ {
		err := r.AddBcc(m.bcc[i])
		if err != nil {
			log.Fatal(err)
		}
	}
	err := r.SetFrom(m.from)
	r.Subject = m.subj
	r.Body = m.body
	output, err := r.Bytes()
	if err != nil {
		log.Fatal(err)
	}
	return output
}

func readConf() (m *escMsg) {
	m = new(escMsg)
	var file io.Reader
	file, err := os.Open("config.conf")
	if err != nil {
		log.Fatal(err)
	}
	rdr := bufio.NewReader(file)
	for {
		data, err := rdr.ReadString(13)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		m.body = m.body + data
		if err == io.EOF {
			break
		}
	}
	return m
}
