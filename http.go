package main

import (
        "log"
        "net/http"
        "net/http/httputil"
        "net/url"
        "strings"
        "fmt"
        "time"
)

func extractProtoAndDomain(target *url.URL) (proto string, domain string, err error) {
        path := target.Path
        // Trim the leading `/`
        if len(path) > 1 && path[0] == '/' {
                path = path[1:]
        }
        // Explode on `/` and make sure we have at least
        // 2 elements (service name and version)
        tmp := strings.Split(path, "/")
        if len(tmp) < 2 {
                return "", "", fmt.Errorf("Invalid path %s", path)
        }
        proto, domain = tmp[0], tmp[1]


        // Rewrite the request's path without the prefix.
        target.Path = "/" + strings.Join(tmp[2:], "/")
        return proto, domain, nil
}

// NewMultipleHostReverseProxy creates a reverse proxy that will randomly
// select a host from the passed `targets`
func NewPathBasedReverseProxy() *httputil.ReverseProxy {
        director := func(req *http.Request) {
                proto, domain, err := extractProtoAndDomain(req.URL)
                if err != nil {
                    log.Print(err)
                    return
                }
                req.URL.Scheme = proto
                req.URL.Host = domain
                log.Print(proto," ", domain," ",req.URL.Path)
        }
        transport := &SurrogateTransport{
            ReadTimeout:    10 * time.Second,
            RequestTimeout: 15 * time.Second,
            Cache: NewInMemoryCache(),
        }
        return &httputil.ReverseProxy{
                Director: director,
                Transport: transport,
        }
}
