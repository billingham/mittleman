package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
  "log"
  "fmt"
  "crypto/sha1"
  //"bytes"
  //"sort"
)

// https://github.com/pkulak/SimpleTransport

// An HTTP RoundTripper that doesn't pool connections. Most of this is ripped from http.Transport.

type Cache map[string]*http.Response


type SurrogateTransport struct {
	ReadTimeout time.Duration

	// RequestTimeout isn't exact. In the worst case, the actual timeout can come at RequestTimeout * 2.
	RequestTimeout time.Duration
}

var SurrogateCache = Cache{}

func buildKey(req *http.Request) (string){

  //headers := sortedHeader(req.Header)
  key := fmt.Sprintf("Transport\nURL: %s\nMethod: %s\nScheme: %s\nHost: %s",req.URL,req.Method,req.URL.Scheme,req.URL.Host)

  return key
}

// func sortedHeader(h http.Header) (string){
//
//   keys := []string
//   for k := range h {
//     keys = append(keys, k)
//   }
//   sort.Ints(keys)
//
//   sort.Strings(h)
//
//   buffer := bytes.Buffer
//   for _, k := range keys {
//     buffer.WriteString(fmt.Sprintf("%s=%s ",k,h[k]))
//   }
//
//   return buffer.String()
// }

func (t *SurrogateTransport) RoundTrip(req *http.Request) (*http.Response, error) {
  key := buildKey(req)
  log.Print(key)

  hash := fmt.Sprintf("%x",sha1.Sum([]byte(key)))
  log.Print(hash)

  log.Print(SurrogateCache)

  if SurrogateCache[hash] != nil {
    log.Print("Cache HIT -> ",hash)
    log.Print(SurrogateCache[hash])
    return SurrogateCache[hash], nil
  }

  res, err := t.originRequest(req)
  if err != nil {
    return nil, err
  }

  SurrogateCache[hash] = &http.Response{
    Status: res.Status,
    Header: res.Header,
    Body: res.Body,
  }
  log.Print(SurrogateCache[hash])
  return SurrogateCache[hash], nil
}

func (t *SurrogateTransport) originRequest(req *http.Request) (*http.Response, error) {
	conn, err := t.dial(req)

	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	readDone := make(chan responseAndError, 1)
	writeDone := make(chan error, 1)

	// Always request GZIP.
	req.Header.Set("Accept-Encoding", "gzip")

	// Write the request.
	go func() {
		err := req.Write(writer)

		if err == nil {
			writer.Flush()
		}

		writeDone <- err
	}()

	// And read the response.
	go func() {
		resp, err := http.ReadResponse(reader, req)

		if err != nil {
			readDone <- responseAndError{nil, err}
			return
		}

		resp.Body = &connCloser{resp.Body, conn}

		if resp.Header.Get("Content-Encoding") == "gzip" {
			resp.Header.Del("Content-Encoding")
			resp.Header.Del("Content-Length")
			resp.ContentLength = -1

			reader, err := gzip.NewReader(resp.Body)

			if err != nil {
				resp.Body.Close()
				readDone <- responseAndError{nil, err}
				return
			} else {
				resp.Body = &readerAndCloser{reader, resp.Body}
			}
		}

		readDone <- responseAndError{resp, nil}
	}()

	if err = <-writeDone; err != nil {
		return nil, err
	}

	r := <-readDone

	if r.err != nil {
		return nil, r.err
	}

	return r.res, nil
}

func (t *SurrogateTransport) dial(req *http.Request) (net.Conn, error) {
	targetAddr := canonicalAddr(req.URL)

	c, err := net.Dial("tcp", targetAddr)

	if err != nil {
		return c, err
	}

	if t.RequestTimeout > 0 && t.ReadTimeout == 0 {
		t.ReadTimeout = t.RequestTimeout
	}

	if t.ReadTimeout > 0 {
		c = newDeadlineConn(c, t.ReadTimeout)

		if t.RequestTimeout > 0 {
			c = newTimeoutConn(c, t.RequestTimeout)
		}
	}

	if req.URL.Scheme == "https" {
		c = tls.Client(c, &tls.Config{ServerName: req.URL.Host})

		if err = c.(*tls.Conn).Handshake(); err != nil {
			return nil, err
		}

		if err = c.(*tls.Conn).VerifyHostname(req.URL.Host); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// canonicalAddr returns url.Host but always with a ":port" suffix
func canonicalAddr(url *url.URL) string {
	addr := url.Host

	if !hasPort(addr) {
		if url.Scheme == "http" {
			return addr + ":80"
		} else {
			return addr + ":443"
		}
	}

	return addr
}

func hasPort(s string) bool {
	return strings.LastIndex(s, ":") > strings.LastIndex(s, "]")
}

type readerAndCloser struct {
	io.Reader
	io.Closer
}

type responseAndError struct {
	res *http.Response
	err error
}

type connCloser struct {
	io.ReadCloser
	conn net.Conn
}

func (c *connCloser) Close() error {
	c.conn.Close()
	return c.ReadCloser.Close()
}

// A connection wrapper that times out after a period of time with no data sent.
type deadlineConn struct {
	net.Conn
	deadline time.Duration
}

func newDeadlineConn(conn net.Conn, deadline time.Duration) *deadlineConn {
	c := &deadlineConn{Conn: conn, deadline: deadline}
	conn.SetReadDeadline(time.Now().Add(deadline))
	return c
}

func (c *deadlineConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)

	if err != nil {
		return
	}

	c.Conn.SetReadDeadline(time.Now().Add(c.deadline))
	return
}

// A connection wrapper that times out after an absolute amount of time.
// Must wrap a deadlineConn or a hung connection may not trigger an error.
type timeoutConn struct {
	net.Conn
	timeout time.Time
}

func newTimeoutConn(conn net.Conn, timeout time.Duration) *timeoutConn {
	return &timeoutConn{Conn: conn, timeout: time.Now().Add(timeout)}
}

func (t *timeoutConn) Read(b []byte) (int, error) {
	if time.Now().After(t.timeout) {
		return 0, errors.New("connection timeout")
	}

	return t.Conn.Read(b)
}
