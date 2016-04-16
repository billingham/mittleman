package main

import (
	"io"
  //"io/ioutil"
	"net/http"
  "fmt"
  "log"
)



func root(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Welcome to Mittleman!")
}




func main() {
	http.HandleFunc("/", root)

	fmt.Println("Starting Mittleman Server!\nSay Hi to George Bluth.")

	proxy := NewPathBasedReverseProxy()
	log.Fatal(http.ListenAndServe(":9090", proxy))
}
