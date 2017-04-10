package main

import (
	"bytes"
	"log"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	"github.com/maanur/escmailer/tui"
)

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
		log.Println("{{LastOne}} не нашел: " + pattern)
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
