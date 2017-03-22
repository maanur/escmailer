package cache

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// Cache implements basic interaction with InterSystems Cache ctlnetd.
type Cache struct {
	writer chan []byte
	reader chan []byte
	nsp    string
}

// DialCache initializes connection to ctelnetd with given address (ex.: "localhost:23"),
//login and password, then returns pointer to created Cache struct.
func DialCache(addr string, login string, pass string) *Cache {
	c := new(Cache)
	c.writer = make(chan []byte, 100)
	c.reader = make(chan []byte, 100)
	var err error
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		fmt.Println(err)
	}
	go func() {
		for {
			br := make([]byte, 1)
			_, err := conn.Read(br)
			if err != nil {
				log.Fatal(err)
			}
			c.reader <- br
		}
	}()
	go func() {
		for {
			bw := <-c.writer
			_, err := conn.Write(bw)
			if err != nil {
				log.Fatal(err)
			}
		}
	}()
	c.ReadFor([]byte("Username:"))
	c.WriteLine([]byte(login))
	c.ReadFor([]byte("Password:"))
	c.WriteLine([]byte(pass))
	c.ReadFor([]byte(">"))
	return c
}

// WriteLine writes slice of bytes to Cache with trailing "^M".
func (cache *Cache) WriteLine(line []byte) {
	buf := append(line, '\x1b', 'M')
	cache.writer <- buf
}

//ReadFor reads bytes from Cache until meets the given line.
func (cache *Cache) ReadFor(line []byte) {
	buf := make([]byte, 0)
	for !compare(buf, line) {
		if len(buf) == len(line) {
			buf = buf[1:]
		}
		buf = append(buf, <-cache.reader...)
		//fmt.Println(string(buf))
	}
}

func compare(sl1 []byte, sl2 []byte) bool {
	if len(sl1) != len(sl2) {
		return false
	}
	for i, t := range sl1 {
		if t != sl2[i] {
			return false
		}
	}
	return true
}

// ReadWord reads from Cache until '\n'.
//If first 2 bytes are '\n', '\r', they are skipped.
func (cache *Cache) ReadWord() []byte {
	var buf []byte
	_ = time.AfterFunc(5*time.Second, func() {
		fmt.Println("5 sec timeout over")
		cache.reader <- []byte{'\n', '\r'}
	})
	buf = append(buf, <-cache.reader...)
	buf = append(buf, <-cache.reader...)
	if compare(buf, []byte{'\n', '\r'}) {
		buf = make([]byte, 0)
	}
	for {
		buf = append(buf, <-cache.reader...)
		if buf[len(buf)-1] == '\n' {
			return buf
		}
	}
}

// Command is used to send specific command to Cache
func (cache *Cache) Command(str string) {
	cache.WriteLine([]byte(str))
	cache.ReadFor([]byte(cache.nsp + ">"))
}

// ChangeNsp is used to change current namespace within Cache session
func (cache *Cache) ChangeNsp(nsp string) {
	cache.WriteLine([]byte("zn \"" + nsp + "\""))
	cache.ReadFor([]byte(nsp + ">"))
	cache.nsp = nsp
}

// Query sends a line to Cache, then returns the answer
func (cache *Cache) Query(str string) string {
	cache.WriteLine([]byte(str))
	cache.ReadFor([]byte(str))
	out := strings.TrimSpace(string(cache.ReadWord()))
	cache.ReadFor([]byte(cache.nsp + ">"))
	return out
}
