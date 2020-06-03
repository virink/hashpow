package hashpow

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	version = "0.1"
	appName = "Hashpow"
	desc    = "A tool for ctfer which make hash collision faster"
	author  = "Virink"
	letter  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	done                               chan struct{}
	err                                error
	pos, port                          int
	code, prefix, suffix, hash, result string
	server                             bool
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

func doMD5(str []byte) string {
	h := md5.New()
	_, err = h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}

func doSha1(str []byte) string {
	h := sha1.New()
	_, err = h.Write(str)
	return hex.EncodeToString(h.Sum(nil))
}

// DoRandom -
func DoRandom(wg *sync.WaitGroup, code, prefix, suffix, hash string, pos, posend int) {
	defer wg.Done()
	var _hash func(str []byte) string
	if hash == "sha1" {
		_hash = doSha1
	} else if hash == "md5" {
		_hash = doMD5
	} else {
		result = "Error hash type!"
		return
	}
	var buffer bytes.Buffer
	r := newRandbo()
	for {
		select {
		case <-done:
			return
		default:
			buffer.Reset()
			tmp := make([]byte, 8)
			if _, err = r.Read(tmp); err != nil {
				close(done)
				return
			}
			if len(prefix) > 0 {
				buffer.WriteString(prefix)
			}
			buffer.Write(tmp)
			if len(suffix) > 0 {
				buffer.WriteString(suffix)
			}
			if _hash(buffer.Bytes())[pos:posend] == code {
				result = string(tmp)
				fmt.Println(result)
				close(done)
				return
			}
		}
	}
}

func usageHandler(c *gin.Context) {
	u := c.Request.URL.Hostname()
	c.String(http.StatusOK, `Usage:
request: %s/hashpow?c=[code]&h=[hash type]&pf=[prefix string]&sf=[suffix sstring]&p=[pos]&r=[true]
Params:
- c [string] Code (**require**)
- t [string] hash Type : md5 sha1 (**require**)
- p [int] starting Position of hash
- pf [string] text Prefix
- sf [string] text Suffix
- r [boolean] Raw resopnse
like: %s/hashpow?c=abcdef&h=md5
like: %s/hashpow?c=abcdef&h=md5&pf=v&sf=k&p=6`, u)
}

// Resp - Response Struct
type Resp struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Data struct {
		Code   string `json:"code"`
		Hash   string `json:"hash"`
		Pos    int    `json:"pos"`
		Prefix string `json:"prefix"`
		Suffix string `json:"suffix"`
		Result string `json:"result"`
	}
}

// Running -
func Running(wg *sync.WaitGroup, code, prefix, suffix, hash string, pos, posend int) Resp {
	result = ""
	done = make(chan struct{})
	wg.Add(16)
	for i := 0; i < 16; i++ {
		go DoRandom(wg, code, prefix, suffix, hash, pos, posend)
	}
	go func() {
		time.Sleep(10 * time.Second)
		select {
		case <-done:
			return
		default:
			fmt.Println("[-] Timeout")
			close(done)
			return
		}
	}()
	wg.Wait()
	resp := Resp{}
	if len(result) > 0 {
		resp = Resp{Code: 0, Msg: "success"}
		resp.Data.Result = result
		resp.Data.Code = code
		resp.Data.Hash = hash
		resp.Data.Pos = pos
		resp.Data.Prefix = prefix
		resp.Data.Suffix = suffix
	} else {
		resp = Resp{
			Code: 1,
			Msg:  fmt.Sprintf("Oops, something error [%s]", result),
		}
		resp.Data.Result = result
	}
	return resp
}

func hashpowHandler(c *gin.Context) {
	code = c.Query("c")
	prefix = c.Query("pf")
	suffix = c.Query("sf")
	hash = c.Query("h")
	_pos := c.Query("p")
	raw := c.Query("r")
	if len(code) == 0 {
		c.JSON(http.StatusInternalServerError, &Resp{Code: 1, Msg: "param code is empty"})
		return
	}
	if len(hash) == 0 {
		c.JSON(http.StatusInternalServerError, &Resp{Code: 1, Msg: "param hash is empty"})
		return
	}
	pos, err = strconv.Atoi(_pos)
	if err != nil {
		pos = 0
	}
	posend := len(code) + pos
	wg := sync.WaitGroup{}
	resp := Running(&wg, code, prefix, suffix, hash, pos, posend)
	if len(raw) > 0 {
		c.String(http.StatusOK, resp.Data.Result)
		return
	}
	if resp.Code == 0 {
		c.JSON(http.StatusOK, resp)
	} else {
		c.JSON(http.StatusInternalServerError, resp)
	}
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	done = make(chan struct{})
	result = ""

	flag.StringVar(&code, "c", "", "part of hash code")
	flag.StringVar(&prefix, "pf", "", "text prefix")
	flag.StringVar(&suffix, "sf", "", "text suffix")
	flag.StringVar(&hash, "t", "", "hash type : md5 sha1")
	flag.IntVar(&pos, "p", 0, "starting position of hash")
	flag.BoolVar(&server, "s", false, "Run as a web server to provide api")
	flag.IntVar(&port, "port", 3000, "Web server port")
	flag.Parse()

	if len(code) == 0 || len(hash) == 0 {
		fmt.Printf("%s %s - %s By %s\n", appName, version, desc, author)
		flag.Usage()
	}

}

// Execute -
func Execute() {
	if server {
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
		r := gin.Default()
		r.GET("/", usageHandler)
		r.GET("/hashpow", hashpowHandler)
		fmt.Printf("WEB Server Listen on http://0.0.0.0:%d\n", port)
		s := &http.Server{
			Addr:           fmt.Sprintf(":%d", port),
			Handler:        r,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		if err = s.ListenAndServe(); err != nil {
			fmt.Println(err.Error())
		}
	} else {
		posend := len(code) + pos
		wg := sync.WaitGroup{}
		wg.Add(16)
		for i := 0; i < 16; i++ {
			go DoRandom(&wg, code, prefix, suffix, hash, pos, posend)
		}
		wg.Wait()
	}
}
