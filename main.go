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
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	Version = "0.1"
	AppName = "Hashpow"
	Desc    = "Just for Venom Team"
	Author  = "Virink"
	letter  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var (
	done                               chan struct{}
	start, end, pos, port              int
	code, prefix, suffix, hash, result string
	server                             bool
	wg                                 sync.WaitGroup
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
	done = make(chan struct{})
	result = ""
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

func runFuckRandom(wg *sync.WaitGroup, code, prefix, suffix, hash string, pos, posend int) {
	defer wg.Done()
	var _hash func(str []byte) string
	if hash == "sha1" {
		_hash = vSha1
	} else {
		_hash = vMD5
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
			r.Read(tmp)
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

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.StringVar(&code, "c", "", "code")
	flag.StringVar(&prefix, "pf", "", "text prefix")
	flag.StringVar(&suffix, "sf", "", "text suffix")
	flag.StringVar(&hash, "h", "", "hash type : md5 sha1")
	flag.IntVar(&pos, "p", 0, "starting position of hash")
	flag.BoolVar(&server, "s", false, "Run as a web server provide api")
	flag.IntVar(&port, "port", 3000, "Web server port")
	flag.Parse()

	if server {
		gin.SetMode(gin.ReleaseMode)
		gin.DisableConsoleColor()
		r := gin.Default()
		r.GET("/hashpow", func(c *gin.Context) {
			done = make(chan struct{})
			var err error
			result = ""
			code = c.Query("c")
			prefix = c.Query("pf")
			suffix = c.Query("sf")
			hash = c.Query("h")
			_pos := c.Query("p")
			pos, err = strconv.Atoi(_pos)
			if err != nil {
				pos = 0
			}
			if len(code) > 0 {
				posend := len(code) + pos
				wg.Add(16)
				for i := 0; i < 16; i++ {
					go runFuckRandom(&wg, code, prefix, suffix, hash, pos, posend)
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
			}
			if len(result) > 0 {
				c.JSON(200, gin.H{
					"code":   code,
					"hash":   hash,
					"pos":    pos,
					"prefix": prefix,
					"suffix": suffix,
					"result": result,
					"msg":    "success",
				})
			} else {
				c.JSON(500, gin.H{"msg": "Oops, something error"})
			}

		})
		fmt.Printf("WEB Server Listen on http://0.0.0.0:%d\n", port)
		// r.Run(fmt.Sprintf(":%d", port))
		s := &http.Server{
			Addr:           fmt.Sprintf(":%d", port),
			Handler:        r,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		s.ListenAndServe()
	} else if len(code) == 0 {
		fmt.Printf(`
[*]***********************************************[*]
[*] %s %s - %s - By %s [*]
[*]***********************************************[*]

`, AppName, Version, Desc, Author)
		return
	} else {
		posend := len(code) + pos
		wg.Add(16)
		for i := 0; i < 16; i++ {
			go runFuckRandom(&wg, code, prefix, suffix, hash, pos, posend)
		}
		wg.Wait()
	}
}
