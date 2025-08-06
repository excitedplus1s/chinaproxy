package main

import (
	"encoding/binary"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	chinadns "github.com/excitedplus1s/gfwutils/dns"
)

const BadGateway = "HTTP/1.0 503 Bad Gateway\r\n\r\n"
const ConnectedOk = "HTTP/1.0 200 Connection Established\r\n\r\n"

type ChinaProxy struct {
	Addr string
	Port string
}

func main() {

	addr := flag.String("addr", "0.0.0.0", "监听地址，默认0.0.0.0")
	port := flag.String("port", "8080", "监听端口，默认8080")
	flag.Parse()
	Run(*addr, *port)
}

func Run(addr string, port string) {
	proxy := NewProxy(addr, port)
	bindAddr := net.JoinHostPort(addr, port)
	log.Printf("ChinaProxy is runing on %s \n", bindAddr)
	http.ListenAndServe(bindAddr, proxy)
}

func NewProxy(addr string, port string) *ChinaProxy {
	return &ChinaProxy{
		Addr: addr,
		Port: port,
	}
}

func (p *ChinaProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		p.HTTP(rw, req)
	} else {
		p.HTTPS(rw, req)
	}

}

func (p *ChinaProxy) HTTP(rw http.ResponseWriter, req *http.Request) {

	transport := http.DefaultTransport
	outReq := new(http.Request)
	*outReq = *req
	res, err := transport.RoundTrip(outReq)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		rw.Write([]byte(err.Error()))
		return
	}
	for key, value := range res.Header {
		for _, v := range value {
			rw.Header().Add(key, v)
		}
	}
	rw.WriteHeader(res.StatusCode)
	io.Copy(rw, res.Body)
	res.Body.Close()
}

type CHReader struct {
	Conn net.Conn
}

func (ch *CHReader) Read(p []byte) (int, error) {
	n, err := ch.Conn.Read(p)
	if err != nil {
		return n, err
	}
	if n < 48 {
		return n, err
	}
	if p[0] != 0x16 {
		return n, err
	}
	if p[5] != 0x1 {
		return n, err
	}
	var offset uint16 = 43
	sessionLength := p[offset]
	offset += 1
	offset += uint16(sessionLength)
	if n < int(offset) {
		return n, err
	}
	buf := make([]byte, n+5)
	recordLength := binary.BigEndian.Uint16(p[3:5])
	tmpBuf := make([]byte, 4)
	copy(tmpBuf[1:], p[6:9])
	copy(buf, p[:offset])
	copy(buf[offset:], p[:5])
	copy(buf[offset+5:], p[offset:])
	binary.BigEndian.PutUint16(buf[3:5], offset-5)
	binary.BigEndian.PutUint16(buf[offset+3:offset+5], recordLength+5-offset)
	if len(p) < n+5 {
		p = make([]byte, n+5)
	}
	copy(p, buf)
	return n + 5, nil
}

func (p *ChinaProxy) HTTPS(rw http.ResponseWriter, req *http.Request) {
	host := req.URL.Host
	hij, ok := rw.(http.Hijacker)
	if !ok {
		return
	}

	client, _, err := hij.Hijack()
	if err != nil {
		return
	}
	domain, port, err := net.SplitHostPort(host)
	if err != nil {
		client.Write([]byte(BadGateway))
		return
	}

	ips, err := chinadns.Client(nil).LookupIP4(domain)
	if err != nil || len(ips) == 0 {
		client.Write([]byte(BadGateway))
		return
	}
	addr := net.JoinHostPort(ips[0], port)

	server, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		client.Write([]byte(BadGateway))
		return
	}

	client.Write([]byte(ConnectedOk))

	clientReader := &CHReader{
		Conn: client,
	}
	go io.Copy(server, clientReader)
	go io.Copy(client, server)
}
