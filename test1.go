package main

/* Начну с отправки тествого сообщения */

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/jpoehls/gophermail"
)

func main() {
	m := readConf()
	fmt.Println(m)
	r := m.ready()
	fmt.Println(string(r))
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
	attach gophermail.Attachment // файлы в аттаче

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
	r.Attachments = []gophermail.Attachment{m.attach}
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
	// test message START
	m.id = 1
	m.from = "me@one.com"
	m.to = []string{"you@two.ru"}
	m.subj = "Превед, Чукотка!"
	m.body = "Проверка! Рас-Рас."
	m.attach.Name = "test.zip"
	var t []string
	for {
		data, err := rdr.ReadString(13)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		t = append(t, data)
		if err == io.EOF {
			break
		}
	}
	m.attach.Data = new(bytes.Buffer)
	_, err = m.attach.Data.Read(pack(t)) // panic: runtime error: invalid memory address or nil pointer dereference
	if err != nil && err!=io.EOF {
		log.Fatal(err)
	}
	// test message END
	return m
}

func pack(files []string) []byte {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for _, file := range files {
		fi, err := os.Lstat(file)
		if err != nil {
			log.Fatal(err)
		}
		f, err := w.Create(fi.Name())
		if err != nil {
			log.Fatal(err)
		}
		b, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write(b) // Сюда прочитать []byte из файла
		if err != nil {
			log.Fatal(err)
		}
	}
	err := w.Close()
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

func customServer() (srv smtp.ServerInfo) {
	// пока минимум...
	srv.Name=prompt("srv addr?","",0)
	srv.TLS=false
	u:=prompt("user?","",0)
	p:=prompt("passwd?","",0)
	h:=prompt("host?","",0)
	srv.Auth=smtp.PlainAuth("",u,p,h)
	return
}

func prompt(ask string, dft string, repeat bool) (output string) {
	// в библитеку бы
	consolereader := bufio.NewReader(os.Stdin)
	fmt.Println(ask)
	rn, err := consolereader.ReadBytes('\r') // this will prompt the user for input
	if err != nil {
		log.Fatal(err)
	}
	if !repeat {
		output = string(rn[:len(rn)-1])
	} else {
		output = string(rn[1 : len(rn)-1])
	}
	if output == "" {
		return dft
	}
	return output
}

func sendall(msgs [][]byte, srv smtp.ServerInfo) {
	

}