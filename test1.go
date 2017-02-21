package main

/* Начну с отправки тествого сообщения */

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	//"io"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"strconv"
	//"strings"

	"github.com/go-ini/ini"
	"github.com/jpoehls/gophermail"
)

// "github.com/go-ini/ini"

func main() {
	srv, _ := readConf()
	fmt.Println(srv.name)
	/*m.send(srv)*/
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
	attach []struct {
		name  string
		files []string
	} // файлы в аттаче
}

func (msg *escMsg) send(srv server) {
	err := smtp.SendMail(srv.addr(), srv.auth(), msg.from, msg.to, msg.ready())
	if err != nil {
		log.Fatal(err)
	}
}

func (msg *escMsg) ready() []byte {
	r := new(gophermail.Message)
	for i := 0; i < len(msg.to); i++ {
		err := r.AddTo(msg.to[i])
		if err != nil {
			log.Fatal(err)
		}
	}
	for i := 0; i < len(msg.cc); i++ {
		err := r.AddCc(msg.cc[i])
		if err != nil {
			log.Fatal(err)
		}
	}
	for i := 0; i < len(msg.bcc); i++ {
		err := r.AddBcc(msg.bcc[i])
		if err != nil {
			log.Fatal(err)
		}
	}
	err := r.SetFrom(msg.from)
	if err != nil {
		log.Fatal(err)
	}
	r.Subject = msg.subj
	r.Body = msg.body
	r.Attachments = msg.convAttach()
	output, err := r.Bytes()
	if err != nil {
		log.Fatal(err)
	}
	return output
}

func (msg escMsg) convAttach() []gophermail.Attachment {
	g := make([]gophermail.Attachment, len(msg.attach))
	for i := 0; i < len(msg.attach); i++ {
		g[i].Name = msg.attach[i].name
		g[i].Data = pack(msg.attach[i].files)
	}
	return g
}

/*func readConfOld() (m *escMsg) {
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
}*/

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
			log.Println(err)
			break
		} else {
			_, err = f.Write(b) // Сюда прочитать []byte из файла
			if err != nil {
				log.Fatal(err)
			}
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
	port int
}

func (srv server) auth() smtp.Auth {
	if srv.name == "" {
		log.Println("Fail making smtp.Auth: no server.name")
		return nil
	}
	u := prompt("mailsrv user?", "")
	p := prompt("mailsrv passwd?", "")
	return smtp.PlainAuth("", u, p, srv.name)
}

func (srv *server) addr() string {
	return srv.name + ":" + strconv.Itoa(srv.port)
}

func customServer() (srv server) {
	// пока минимум...
	var err error
	srv.name = prompt("srv address?", "mail")
	for {
		srv.port, err = strconv.Atoi(prompt("srv port?", "25"))
		if err != nil {
			fmt.Println(err)
			fmt.Println("Try again!")
		} else {
			break
		}
	}
	return
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

func readConf() (srv server, msg *escMsg) {
	msg = new(escMsg)
	conf, err := ini.Load("config.ini")
	if err != nil {
		log.Fatal(err)
	}
	//server START
	srv, err = readConfServer(conf)
	if err != nil {
		log.Println(err)
		log.Println("Failed to read server from .ini")
		srv = customServer()
	}
	//server END
	//message START
	sections := conf.Sections()
	for _, sect := range sections {
		if sect.Name() != "mailsrv" && sect.Name() != "cachesrv" {
			msg.subj = sect.Key("addr").String()
		}
	}
	//message END
	return srv, msg
}

func readConfServer(conf *ini.File) (server, error) {
	var srv server
	s, err := conf.GetSection("mailsrv")
	if err!=nil {
		return srv, err
	}
	r,err:=s.GetKey("name")
	if err!=nil {
		return srv, err
	}
	srv.name=r.String()
	r,err=s.GetKey("port")
	if err!=nil {
		return srv, err
	}
	srv.port,err=r.Int()
	return srv, err
}