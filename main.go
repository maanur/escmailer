package main

/* Начну с отправки тествого сообщения */

import (
	"fmt"
	"io"
	//"io"

	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	//"sync"
	//"text/template"

	"github.com/maanur/escmailer/tui"
	"github.com/maanur/escmailer/txtparser"
	//"github.com/maanur/escmailer/cache"

	//"time"

	"github.com/go-ini/ini"
)

func main() {
	srv, msgs := readConf()
	/*for _, i := range msgs {
		msgs[i].send(srv)
	}*/
	var selector = make([]tui.SelectorItem, len(msgs))
	for i, d := range msgs {
		selector[i] = d
	}
	for _, msg := range msgs {
		msg.send(srv)
	}
	os.Exit(0)
}

// copyToDir параллельно копирует массив файликов в данную директорию
func copyToDir(files []string, dest string) {
	//var wg sync.WaitGroup
	_, err := os.Lstat(dest)
	if err != nil && os.IsExist(err) {
		log.Fatal(err)
	} else {
		if os.IsNotExist(err) {
			err = os.Mkdir(dest, os.FileMode(0777))
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	for i := 0; i < len(files); i++ {
		//wg.Add(1)
		func() {
			_, fname := filepath.Split(files[i])
			err := copyFileContents(files[i], filepath.Clean(dest+string(os.PathSeparator)+fname))
			if err != nil {
				log.Fatal(err)
			}
			//defer wg.Done()
		}()
	}
	//wg.Wait()
	report := ""
	for _, s := range files {
		report = report + "\n" + s
	}
	fmt.Println("Скопировали " + report + "\n" + "в директорию " + dest)
}

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

type server struct {
	name string
	port int
	auth smtp.Auth
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
	srv.name = tui.Prompt("Адрес почтового сервера?", "mail")
	for {
		srv.port, err = strconv.Atoi(tui.Prompt("Порт почтового сервера?", "25"))
		if err != nil {
			fmt.Println(err)
			fmt.Println("Try again!")
		} else {
			break
		}
	}
	return
}

func readConf() (srv server, msgs []escMsg) {
	conf, err := ini.Load("config.ini")
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
	srv.auth = func() smtp.Auth {
		if srv.name == "" {
			log.Println("Fail making smtp.Auth: no server.name")
			return nil
		}
		u := tui.Prompt("Логин пользователя на почтовом сервере?", "")
		p := tui.Prompt("Пароль пользователя на почтовом сервере?", "")
		return smtp.PlainAuth("", u, p, srv.name)
	}()
	//message START
	signature := func() string {
		signa, err := conf.GetSection("signature")
		if err != nil {
			return ""
		}
		if !signa.HasKey("value") {
			return ""
		}
		return signa.Key("value").String()
	}()
	sections := conf.Sections()
	var msgsections []*ini.Section
	for _, ms := range sections {
		if strings.Contains(ms.Name(), "letter") {
			msgsections = append(msgsections, ms)
		}
	}
	var selector = make([]tui.SelectorItem, len(msgsections))
	for i, d := range msgsections {
		selector[i] = msgSection{d}
	}
	var msgchosen []*ini.Section
	for _, msgsec := range tui.MultiChoice(selector) {
		s, ok := msgsec.(msgSection)
		if ok {
			msgchosen = append(msgchosen, s.sect)
		}
	}
	for _, sect := range msgchosen {
		var msg escMsg
		msg.name = sect.Key("name").String()
		p := new(txtparser.Parser)
		msg.to = sect.Key("to").Strings(",")
		msg.cc = sect.Key("cc").Strings(",")
		msg.bcc = sect.Key("bcc").Strings(",")
		msg.from = sect.Key("from").String()
		msg.subj = p.ParseString(sect.Key("subj").String())
		//add template
		msg.body = p.ParseString(sect.Key("body").String() + string(10) + "--" + string(10) + signature)
		for _, att := range sect.Key("attach").Strings(",") {
			msg.attach = append(msg.attach, readAttach(conf, att))
		}
		msgs = append(msgs, msg)
	}
	//message END
	return srv, msgs
}

type msgSection struct {
	sect *ini.Section
}

func (sec msgSection) Name() string {
	return sec.sect.Key("name").String()
}

func readAttach(conf *ini.File, sectname string) (attach struct {
	name      string
	files     []string
	checkDir  string
	checkNsp  string
	checkFunc []string
}) {
	p := new(txtparser.Parser)
	attsec, err := conf.GetSection(sectname)
	if err != nil {
		log.Println(err)
	} else {
		attach.name = attsec.Key("name").String()
		if attsec.HasKey("directory") && attsec.Key("directory").String() != "" {
			dir := p.parseString(attsec.Key("directory").String())
			files := attsec.Key("files").Strings(",")
			if strings.HasSuffix(dir, string(os.PathSeparator)) {
				for _, file := range files {
					attach.files = append(attach.files, checkFileName(p.parseString(dir+file))...)
				}
			} else {
				for _, file := range files {
					attach.files = append(attach.files, checkFileName(p.parseString(dir+string(os.PathSeparator)+file))...)
				}
			}

		} else {
			for _, f := range attsec.Key("files").Strings(",") {
				attach.files = append(attach.files, checkFileName(p.parseString(f))...)
			}
		}
	}
	if attsec.HasKey("checkDir") {
		attach.checkDir = attsec.Key("checkDir").String()
	}
	if attsec.HasKey("checkNsp") {
		attach.checkNsp = attsec.Key("checkNsp").String()
	}
	if attsec.HasKey("checkFunc") {
		attach.checkFunc = strings.Split(attsec.Key("checkFunc").String(), string('\n'))
	}
	return
}

func checkFileName(file string) (out []string) {
	matches, err := filepath.Glob(filepath.FromSlash(file))
	if err != nil {
		log.Println(err)
	}
	if len(matches) == 0 {
		log.Println("No matches for: " + file)
	}
	return matches
}
