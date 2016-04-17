package main

import (
	//"io"
  //"io/ioutil"
	"net/http"
  "log"
)


func main() {
	host := "localhost:9090"
	log.Print("Starting Mittleman Server! Say Hi to George Bluth.")
	log.Printf("==> http://%s",host)

	proxy := NewPathBasedReverseProxy()
	log.Fatal(http.ListenAndServe(host, proxy))
}
