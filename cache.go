package main

import (
  "crypto/sha1"
  "fmt"
	"net/http"
  "io/ioutil"
  "bytes"
)

type Cache interface {
  Get(string) (CacheContent, bool)
  Set(string, CacheContent) error
  Delete(string) error
}

type CacheContent struct {
  Key string
  Body string
  Header map[string][]string
  Status string
}

func HashKey(k string) string {
  return fmt.Sprintf("%x",sha1.Sum([]byte(k)))
}

func NewCacheContent(k string, body string) (CacheContent, error){
  cc := CacheContent{
    Key: k,
    Body: string(body),
  }
  return cc, nil
}

func NewCacheContentHttp(k string, r *http.Response) (CacheContent, error){
  defer r.Body.Close()
  body, err := ioutil.ReadAll(r.Body)
  if err != nil {
    return CacheContent{}, err
  }
  cc := CacheContent{
    Key: k,
    Body: string(body),
    Status: r.Status,
    Header: r.Header,
  }

  return cc, nil
}

func NewHttpResponseFromCache(cc CacheContent) *http.Response {
  body := nopCloser{bytes.NewBufferString(cc.Body)}
  return &http.Response{
    Status: cc.Status,
    Body: body,
  }

}



type InMemoryCache struct {
  Store map[string]CacheContent
}

func NewInMemoryCache() *InMemoryCache {
  s := make(map[string]CacheContent)
  return &InMemoryCache{Store: s}
}

func (c *InMemoryCache) Get(k string) (CacheContent, bool) {
  val, present := c.Store[k]
  return val, present
}

func (c *InMemoryCache) Set(k string, v CacheContent) error{
  c.Store[k] = v
  return nil
}

func (c *InMemoryCache) Delete(k string) error{
  delete(c.Store, k)
  return nil
}
