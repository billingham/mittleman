package main

import (
	"io"
  "io/ioutil"
	"net/http"
  "fmt"
  "strings"
)

func splitPath(p string) (string, string, string){
  s := strings.Split(p[1:], "/")
  prefix, domain, path := s[0], s[1], strings.Join(s[2:],"/")
  return prefix, domain, path
}

func root(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Welcome to Mittleman!")
}

func proxy(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "HTTP Proxy")
  //fmt.Println(r.URL.Path[1:])

  proto, domain, path := splitPath(r.URL.Path)
  fmt.Println(proto, domain, path)


  client := &http.Client{
  	//CheckRedirect: redirectPolicyFunc,
  }

  url := fmt.Sprintf("%s://%s/%s", proto, domain, path)
  req, err := http.NewRequest("GET", url, nil)

  //req.Header.Add("If-None-Match", `W/"wyzzy"`)
  resp, err := client.Do(req)
  if err != nil {
    fmt.Println(err)
  }
  defer resp.Body.Close()
  //body, err := ioutil.ReadAll(resp.Body)

  io.WriteString(w, resp.Status)


}


func main() {
	http.HandleFunc("/", root)
  http.HandleFunc("/http/", proxy)

	fmt.Println("Starting Mittleman Server!\nSay Hi to George Bluth.")
	http.ListenAndServe(":8000", nil)
}
