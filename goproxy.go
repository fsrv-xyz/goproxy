package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

type HttpConnection struct {
	Request  *http.Request
	Response *http.Response
}

type HttpConnectionChannel chan *HttpConnection

var connChannel = make(HttpConnectionChannel)

func PrintHTTP(conn *HttpConnection) {
	fmt.Printf("%v => %v %v\n", conn.Request.RemoteAddr, conn.Request.Method, conn.Request.RequestURI)
}

func HandleHTTP() {
	for {
		select {
		case conn := <-connChannel:
			PrintHTTP(conn)
		}
	}
}

type Proxy struct {
}

func NewProxy() *Proxy { return &Proxy{} }

func (p *Proxy) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error
	var req *http.Request
	client := &http.Client{}

	//log.Printf("%v %v", r.Method, r.RequestURI)
	req, err = http.NewRequest(r.Method, r.RequestURI, r.Body)
	for name, value := range r.Header {
		req.Header.Set(name, value[0])
	}
	resp, err = client.Do(req)
	r.Body.Close()

	// combined for GET/POST
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}

	conn := &HttpConnection{r, resp}

	for k, v := range resp.Header {
		wr.Header().Set(k, v[0])
	}
	wr.WriteHeader(resp.StatusCode)
	io.Copy(wr, resp.Body)
	resp.Body.Close()

	PrintHTTP(conn)
	//connChannel <- &HttpConnection{r,resp}
}

func main() {
	//go HandleHTTP()
	proxy := NewProxy()
	err := http.ListenAndServe(":9823", proxy)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

