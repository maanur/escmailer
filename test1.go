package main

/* Начну с отправки тествого сообщения */

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"io/ioutil"

	"github.com/jpoehls/gophermail"
)

func main() {
	m := readConf()
	fmt.Println(m)
	log.Fatal("Not yet")
	pack([]string{"config.conf"})
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

func pack(files []string) string {
	buf := new(bytes.Buffer)
	arch := zip.NewWriter(buf)
	for _, file := range files {
		f, err := arch.Create(file)
		if err != nil {
			log.Fatal(err)
		}

		_, err = f.Write(readFile(file))
		if err != nil {
			log.Fatal(err)
		}
	}
	var fm os.FileMode
	err := ioutil.WriteFile("test.zip", readFile(files[0]), fm)
	if err != nil {
		log.Fatal(err)
	}
	return "test.zip"
}
func readFile(f string) []byte {
	o, err := ioutil.ReadFile(f)
	if err != nil {
		log.Fatal(err)
	}
	return o
}

/*
// Create a buffer to write our archive to.
buf := new(bytes.Buffer)

// Create a new zip archive.
w := zip.NewWriter(buf)

// Add some files to the archive.
var files = []struct {
    Name, Body string
}{
    {"readme.txt", "This archive contains some text files."},
    {"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
    {"todo.txt", "Get animal handling licence.\nWrite more examples."},
}
for _, file := range files {
    f, err := w.Create(file.Name)
    if err != nil {
        log.Fatal(err)
    }
    _, err = f.Write([]byte(file.Body))
    if err != nil {
        log.Fatal(err)
    }
}

// Make sure to check the error on Close.
err := w.Close()
if err != nil {
    log.Fatal(err)
}
*/
