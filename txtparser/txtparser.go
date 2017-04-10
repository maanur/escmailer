package txtparser

import (
	"bytes"
	"log"
	"path/filepath"
	"strconv"
	"sync"
	"text/template"
	"time"

	"github.com/maanur/escmailer/tui"
)

// Parser is a text parser fo /text/template with built-in common methods and data
type Parser struct {
	count     int
	countOnce sync.Once
}

// Count возвращает значение внутреннего счетчика парсера p
func (p *Parser) Count() int {
	p.countOnce.Do(func() {
		for {
			cnt, err := strconv.Atoi(tui.Prompt("Значение {{Count}}", "0"))
			if err != nil {
				log.Println(err)
			} else {
				p.count = cnt
			}
		}
	})
	return p.count
}

// Today возвращает текущую дату в формате ггггммдд
func (p *Parser) Today() string {
	return time.Now().Format("20060102")
}

// Yesterday возвращает предыдущую дату в формате ггггммдд
func (p *Parser) Yesterday() string {
	return time.Now().AddDate(0, 0, -1).Format("20060102")
}

// LastOne обрабатывает переданную строку как шаблон адреса файла, находит все совпадения и возвращает максимальное
func (p *Parser) LastOne(pattern string) string {
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

// ParseString проводит переданную строку через парсер, возвращая строку с выполненными методами парсера.
func (p *Parser) ParseString(input string) string {
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
