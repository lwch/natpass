package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lwch/runtime"
)

var cli = &http.Client{}
var upgrader = websocket.Upgrader{}
var dialer = websocket.Dialer{}

func main() {
	http.HandleFunc("/", next)
	http.ListenAndServe(":8001", nil)
}

func normal(w http.ResponseWriter, r *http.Request) {
	u := r.URL
	u.Scheme = "http"
	u.Host = "127.0.0.1:8000"
	req, err := http.NewRequest(r.Method, u.String(), r.Body)
	runtime.Assert(err)

	for key, values := range r.Header {
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}

	rep, err := cli.Do(req)
	runtime.Assert(err)
	defer rep.Body.Close()

	for key, values := range rep.Header {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}

	w.WriteHeader(rep.StatusCode)

	_, err = io.Copy(w, rep.Body)
	runtime.Assert(err)
}

func ws(w http.ResponseWriter, r *http.Request) {
	u := r.URL
	u.Scheme = "ws"
	u.Host = "127.0.0.1:8000"

	hdr := make(http.Header)
	for key, values := range r.Header {
		if strings.HasPrefix(key, "Sec-") {
			continue
		}
		for _, value := range values {
			hdr.Add(key, value)
		}
	}

	hdr.Del("Connection")
	hdr.Del("Upgrade")

	remote, resp, err := dialer.Dial(u.String(), hdr)
	runtime.Assert(err)
	defer resp.Body.Close()
	defer remote.Close()

	local, err := upgrader.Upgrade(w, r, nil)
	runtime.Assert(err)
	defer local.Close()

	cp := func(wg *sync.WaitGroup, dst, src *websocket.Conn) {
		defer wg.Done()
		defer dst.Close()
		defer src.Close()
		for {
			t, data, err := src.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			err = dst.WriteMessage(t, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}

	var wg sync.WaitGroup

	wg.Add(2)
	go cp(&wg, local, remote)
	go cp(&wg, remote, local)

	wg.Wait()
}

func next(w http.ResponseWriter, r *http.Request) {
	upgrade := r.Header.Get("Connection")

	if upgrade == "Upgrade" {
		ws(w, r)
		return
	}
	normal(w, r)
}
