package main

/* Начну с отправки тествого сообщения */

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	//"io"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"strconv"
	//"strings"
	"path/filepath"
	"strings"
	//"sync"
	"text/template"

	"github.com/maanur/escmailer/tui"

	"time"

	"github.com/go-ini/ini"
	"github.com/jpoehls/gophermail"
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

type escMsg struct {
	// совокупность параметров для письма
	name             string   // идентификактор письма
	to, cc, bcc      []string // кому, копия, скрытая копия
	from, subj, body string   // от, тема, текст
	attach           []struct {
		name     string
		files    []string
		checkDir string
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
		if msg.attach[i].checkDir != "" && tui.PromptYN("Проверить файлы для "+msg.attach[i].name+"?", false) {
			copyToDir(msg.attach[i].files, msg.attach[i].checkDir)
			if !tui.PromptYN("Файлы корректны?", false) {
				log.Fatal("Файлы не корректны...")
			}
		}
		g[i].Name = msg.attach[i].name
		g[i].Data = pack(msg.attach[i].files)
	}
	return g
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
	for i:=0 ; i<len(files); i++ {
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
	for _, sect := range msgchosen { //здесь надо будет добавить выбор писем по regexp
		var msg escMsg
		msg.name = sect.Key("name").String()
		p := newParser(msg.name)
		msg.to = sect.Key("to").Strings(",")
		msg.cc = sect.Key("cc").Strings(",")
		msg.bcc = sect.Key("bcc").Strings(",")
		msg.from = sect.Key("from").String()
		msg.subj = p.parseString(sect.Key("subj").String())
		//add template
		msg.body = p.parseString(sect.Key("body").String() + string(10) + "--" + string(10) + signature)
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
	name     string
	files    []string
	checkDir string
}) {
	p := new(parser)
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

type parser struct {
	count int
}

func (p *parser) Count() int {
	return p.count
}

func (p *parser) Today() string {
	return time.Now().Format("20060102")
}

func (p *parser) Yesterday() string {
	return time.Now().AddDate(0, 0, -1).Format("20060102")
}

func (p *parser) LastOne(pattern string) string {
	matches, err := filepath.Glob(filepath.FromSlash(pattern))
	if err != nil {
		log.Println(err)
	}
	if len(matches) == 0 {
		fmt.Println("{{LastOne}} не нашел: " + pattern)
		return ""
	}
	return func() string {
		max := matches[0]
		for _, i := range matches {
			if i > max {
				max = i
			}
		}
		return max
	}()
}

func (p *parser) parseString(input string) string {
	tmpl, err := template.New("test").Parse(input)
	if err != nil {
		log.Fatal(err)
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, p)
	if err != nil {
		log.Fatal(err)
	}
	return string(buf.Bytes())
}

func newParser(name string) *parser {
	p := new(parser)
	p.count = func() int {
		for {
			cnt, err := strconv.Atoi(tui.Prompt("Значение {{Count}} для письма : "+name, "0"))
			if err != nil {
				log.Println(err)
			} else {
				return cnt
			}
		}
	}()
	return p
}

/*func findLastDir(dir string) string { //поиск последней поддиректории. Актуальность под вопросом.
	d, err := os.Open(dir)
	if err != nil {
		log.Fatal(err)
	}
	sd, err := d.Readdirnames(0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sd[len(sd)])
	return sd[len(sd)]
}*/
