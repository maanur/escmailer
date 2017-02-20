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
	"net/smtp"
	"os"
	"strconv"
	"strings"

	"github.com/jpoehls/gophermail"
)

// "github.com/go-ini/ini"

func main() {
	m := readConf()
	srv := customServer()
	sendall([]*escMsg{m}, srv)
	os.Exit(0)
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
	if err != nil {
		log.Fatal(err)
	}
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
	m.from = "mal@esc.ru"
	m.to = []string{"grushin_m@esc.ru"}
	m.subj = "Превед, Чукотка!"
	m.body = "Проверка! Рас-Рас."
	m.attach.Name = "test.zip"
	var t []string
	for {
		data, err := rdr.ReadString(10)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		t = append(t, strings.TrimSpace(data))
		fmt.Println(t)
		if err == io.EOF {
			break
		}
	}
	m.attach.Data = pack(t)
	// test message END
	return m
}

func pack(files []string) *bytes.Buffer { //поменял вывод с []bytes на bytes.Buffer, который должен бы быть io.Reader
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
			fmt.Println(err)
			break
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
	return buf
}

type server struct {
	name string
	auth smtp.Auth
	port int
}

func customServer() (srv server) {
	// пока минимум...
	var err error
	srv.name = prompt("srv addr?", "")
	for {
		srv.port, err = strconv.Atoi(prompt("srv port?", ""))
		if err != nil {
			fmt.Println(err)
			fmt.Println("Try again!")
		} else {
			fmt.Println("OK")
			break
		}
	}
	u := prompt("user?", "")
	p := prompt("passwd?", "")
	srv.auth = smtp.PlainAuth("", u, p, srv.name)
	return
}

func sendall(msgs []*escMsg, srv server) {
	for i := 0; i < len(msgs); i++ {
		err := smtp.SendMail(srv.name+":"+strconv.Itoa(srv.port), srv.auth, msgs[i].from, msgs[i].to, msgs[i].ready())
		if err != nil {
			log.Fatal(err)
		}
	}
}

func prompt(ask string, dft string) (output string) {
	// в библитеку бы
	consolereader := bufio.NewReader(os.Stdin)
	fmt.Println(ask)
	rn, err := consolereader.ReadBytes('\r') // this will prompt the user for input
	if err != nil {
		log.Fatal(err)
	}
	output = string(rn[:len(rn)-1])
	if output == "" {
		return dft
	}
	return output
}
