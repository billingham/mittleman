package main

import (
  "testing"
)

func TestHashKey(t *testing.T) {
  h := HashKey("hello")
  if h != "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d" {
    t.Fail()
  }
}

func TestNewCacheContent(t *testing.T) {
  cc,_ := NewCacheContent("key","body")
  if cc.Key != "key" {
    t.Fail()
  }
  if cc.Body != "body" {
    t.Fail()
  }
}

func TestInMemoryCacheGet(t *testing.T) {
  c := NewInMemoryCache()
  _, present := c.Get("key-not-found")
  if present {
    t.Fail()
  }
}

func TestInMemoryCacheSet(t *testing.T) {
  c := NewInMemoryCache()
  k := "key-to-add"
  v := "value-of-key"
  cc,_ := NewCacheContent(k,v)
  c.Set(k,cc)
  ncc, present := c.Get(k)
  if !present{
    t.Fail()
  }
  if v != ncc.Body {
    t.Fail()
  }
}

func TestInMemoryCacheDelete(t *testing.T) {
  c := NewInMemoryCache()
  k := "key-to-add"
  v := "value-of-key"
  cc,_ := NewCacheContent(k,v)
  c.Set(k,cc)
  c.Delete(k)
  _, present := c.Get("key-not-found")
  if present {
    t.Fail()
  }
}
