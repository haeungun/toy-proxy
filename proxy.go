package main

import (
	"io"
	"log"
	"net/http"
)

type Proxy struct {}

type HttpConnection struct {
	Request  *http.Request
	Response *http.Response
}

func removeProxyHeaders(r *http.Request) {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers
	// Hop-by-hop headers
	//    These headers are meaningful only for a single transport-level connection, 
	//    and must not be retransmitted by proxies or cached.
	r.Header.Del("Connection")
	r.Header.Del("Keep-alive")
	r.Header.Del("Transfer-Encoding")
	r.Header.Del("Upgrade")

	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")
	r.Header.Del("Proxy-Connection")	// not standard, but curl lib use it.
}

func PrintHTTP(conn *HttpConnection) {
	log.Println("==============================")
	log.Printf("%v %v %v\n", conn.Request.Method, conn.Request.RequestURI, conn.Request.Proto)
	for k, v := range conn.Request.Header {
		log.Println(k, ":", v)
	}
	log.Println("------------------------------")
	log.Printf("HTTP/1.1 %v\n", conn.Response.Status)
	for k, v := range conn.Response.Header {
		log.Println(k, ":", v)
	}
	log.Println(conn.Response.Body)
	log.Println("==============================")
}

func NewProxy() *Proxy { return &Proxy{} }

func (p *Proxy) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	
	var res *http.Response
	var req *http.Request
	var err error

	client := &http.Client{}

	removeProxyHeaders(r)

	req, err = http.NewRequest(r.Method, r.RequestURI, r.Body)
	for name, value := range r.Header {
		req.Header.Set(name, value[0])
	}

	res, err = client.Do(req)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		log.Printf("%v %v", http.StatusInternalServerError, err.Error())
		return
	}
	defer res.Body.Close()

	conn := &HttpConnection{r, res}

	for k, v := range res.Header {
		wr.Header().Set(k, v[0])
	}
	
	wr.WriteHeader(res.StatusCode)
	io.Copy(wr, res.Body)

	PrintHTTP(conn)
}

func main() {
	proxy := NewProxy()
	err := http.ListenAndServe(":12345", proxy)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
