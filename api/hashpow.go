package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/virink/hashpow/hashpow"
)

// Resp -
type Resp = hashpow.Resp

func toJSON(r *Resp) string {
	res, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(res)
}

// Handler -
func Handler(w http.ResponseWriter, r *http.Request) {
	wg := sync.WaitGroup{}
	q := r.URL.Query()
	code := q.Get("c")
	prefix := q.Get("pf")
	suffix := q.Get("sf")
	hash := q.Get("h")
	_pos := q.Get("p")
	raw := c.Query("r")
	if len(code) == 0 || len(hash) == 0 {
		fmt.Fprintf(w, `Usage:
request: /?c=[code]&h=[hash type]&pf=[prefix string]&sf=[suffix sstring]&p=[pos]&r=[true]
Params:
- c [string] Code (**require**)
- t [string] hash Type : md5 sha1 (**require**)
- p [int] starting Position of hash
- pf [string] text Prefix
- sf [string] text Suffix
- r [boolean] Raw resopnse
eg: /?c=abcdef&h=md5
    /?c=abcdef&h=md5&pf=v&sf=k&p=6`)
		return
	}
	pos, err := strconv.Atoi(_pos)
	if err != nil {
		pos = 0
	}
	posend := len(code) + pos
	wg := sync.WaitGroup{}
	resp := Running(&wg, code, prefix, suffix, hash, pos, posend)
	if len(raw) > 0 {
		fmt.Fprintf(w, resp.Data.Result)
		return
	}
	if resp.Code == 0 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Fprintf(w, toJSON(resp))
}
