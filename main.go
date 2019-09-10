package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"runtime"
	"time"
)

const (
	Version = "0.1"
	AppName = "Hashpow"
	Desc    = "Just for Venom Team"
	Author  = "Virink"
	letter  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	done                       chan bool
	start, end, pos            int
	code, prefix, suffix, hash string
)

type randbo struct {
	rand.Source
}

func (r *randbo) Read(p []byte) (n int, err error) {
	todo := len(p)
	for {
		val := r.Int63()
		for todo > 0 {
			p[todo-1] = letter[int(val&(1<<6-1))%52]
			todo--
			if todo == 0 {
				return len(p), nil
			}
			val >>= 6
		}
	}
}
func newFrom(src rand.Source) io.Reader {
	return &randbo{src}
}
func newRandbo() io.Reader {
	return newFrom(rand.NewSource(time.Now().UnixNano()))
}
func init() {
	done = make(chan bool)
}

func vMD5(str []byte) string {
	h := md5.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}

func vSha1(str []byte) string {
	h := sha1.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}

func runFuckRandom(code, prefix, suffix, hash string, pos, posend int) {
	var _hash func(str []byte) string
	if hash == "sha1" {
		_hash = vSha1
	} else {
		_hash = vMD5
	}

	var buffer bytes.Buffer
	r := newRandbo()
	for {
		buffer.Reset()
		tmp := make([]byte, 8)
		r.Read(tmp)
		if len(prefix) > 0 {
			buffer.WriteString(prefix)
		}
		buffer.Write(tmp)
		if len(suffix) > 0 {
			buffer.WriteString(suffix)
		}
		if _hash(buffer.Bytes())[pos:posend] == code {
			fmt.Println(string(tmp))
			done <- true
			return
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.StringVar(&code, "c", "", "code")
	flag.StringVar(&prefix, "pf", "", "text prefix")
	flag.StringVar(&suffix, "sf", "", "text suffix")
	flag.StringVar(&hash, "h", "", "hash type : md5 ")
	flag.IntVar(&pos, "p", 0, "starting position of hash")
	flag.Parse()

	if len(code) == 0 {
		fmt.Printf(`
[*]***********************************************[*]
[*] %s %s - %s - By %s [*]
[*]***********************************************[*]

`, AppName, Version, Desc, Author)
		return
	}

	posend := len(code) + pos
	for i := 0; i < 16; i++ {
		go runFuckRandom(code, prefix, suffix, hash, pos, posend)
	}
	<-done
}
