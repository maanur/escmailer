package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"

	"github.com/jpoehls/gophermail"
	//"github.com/maanur/escmailer/cache"
	"github.com/maanur/escmailer/tui"
)

type escMsg struct {
	// совокупность параметров для письма
	name             string   // идентификактор письма
	to, cc, bcc      []string // кому, копия, скрытая копия
	from, subj, body string   // от, тема, текст
	attach           []struct {
		name      string
		files     []string
		checkDir  string
		checkNsp  string
		checkFunc []string
	} // файлы в аттаче
}

func (msg escMsg) Name() string {
	return msg.name
}

func (msg escMsg) send(srv server) {
	err := smtp.SendMail(srv.addr(), srv.auth, msg.from, append(msg.to, append(msg.cc, msg.bcc...)...), msg.ready())
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
		checkAttach(msg.attach[i])
		g[i].Name = msg.attach[i].name
		g[i].Data = pack(msg.attach[i].files)
	}
	return g
}

func checkAttach(attach struct {
	name      string
	files     []string
	checkDir  string
	checkNsp  string
	checkFunc []string
}) {
	if attach.checkDir != "" && tui.PromptYN("Проверить файлы для "+attach.name+"?", false) {
		copyToDir(attach.files, attach.checkDir)
		if !tui.PromptYN("Файлы корректны?", false) {
			log.Fatal("Файлы не корректны...")
		}
	}
}

func pack(files []string) *bytes.Buffer { //поменял вывод с []bytes на bytes.Buffer, который должен бы быть io.Reader
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for _, file := range files {
		fmt.Println("Архивирую: " + file)
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
