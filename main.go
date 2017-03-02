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
	//"text/template"

	"strings"

	"github.com/go-ini/ini"
	"github.com/jpoehls/gophermail"
)

func main() {
	srv, msgs := readConf()
	for _, msg := range msgs {
		msg.send(srv)
	}
	/*m.send(srv)*/
	os.Exit(0)
}

type escMsg struct {
	// совокупность параметров для письма
	id               int      // идентификактор письма
	to, cc, bcc      []string // кому, копия, скрытая копия
	from, subj, body string   // от, тема, текст
	attach           []struct {
		name  string
		files []string
	} // файлы в аттаче
}

func (msg *escMsg) send(srv server) {
	err := smtp.SendMail(srv.addr(), srv.auth(), msg.from, append(msg.to, append(msg.cc, msg.bcc...)...), msg.ready())
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

func readConfServer(conf *ini.File) (server, error) {
	var srv server
	s, err := conf.GetSection("mailsrv")
	if err != nil {
		return srv, err
	}
	r, err := s.GetKey("name")
	if err != nil {
		return srv, err
	}
	srv.name = r.String()
	r, err = s.GetKey("port")
	if err != nil {
		return srv, err
	}
	srv.port, err = r.Int()
	return srv, err
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

func readConf() (srv server, msgs []*escMsg) {
	conf, err := ini.Load("config.ini")
	/*t1:=template.New("MailSection")*/
	conf.BlockMode = false
	if err != nil {
		log.Fatal(err)
	}
	srv, err = readConfServer(conf)
	if err != nil {
		log.Println(err)
		log.Println("Failed to read server from .ini")
		srv = customServer()
	}
	sections := conf.Sections()
	//message START
	for _, sect := range sections {
		if sect.Name() == "test" {
			msg := new(escMsg)
			msg.id, err = sect.Key("id").Int()
			if err != nil {
				log.Fatal(err)
			}
			msg.to = sect.Key("to").Strings(",")
			msg.cc = sect.Key("cc").Strings(",")
			msg.bcc = sect.Key("bcc").Strings(",")
			msg.from = sect.Key("from").String()
			msg.subj = sect.Key("subj").String()
			msg.body = sect.Key("body").String()
			for _, att := range sect.Key("attach").Strings(",") {
				msg.attach = append(msg.attach, readAttach(conf, att))
			}
			msgs = append(msgs, msg)
		}
	}
	//message END
	return srv, msgs
}

func readAttach(conf *ini.File, sectname string) (attach struct {
	name  string
	files []string
}) {
	attsec, err := conf.GetSection(sectname)
	if err != nil {
		log.Println(err)
	} else {
		attach.name = attsec.Key("name").String()
		if attsec.HasKey("directory") && attsec.Key("directory").String() != "" {
			dir := attsec.Key("directory").String()
			files := attsec.Key("files").Strings(",")
			if strings.HasSuffix(dir, string(os.PathSeparator)) {
				for _, file := range files {
					attach.files = append(attach.files, dir+file)
				}
			} else {
				for _, file := range files {
					attach.files = append(attach.files, dir+string(os.PathSeparator)+file)
				}
			}

		} else {
			attach.files = attsec.Key("files").Strings(",")
		}
	}
	return
}
