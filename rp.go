package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"
)

var LISTEN = flag.String("listen", "0.0.0.0:80", "Host and Port to listen")
var CONFIG = flag.String("config", "rp.json", "Default config json (map url to backend)")
var REWRITE_HOST = flag.Bool("rewrite_host", false, "Rewrite host header to backend")
var VERBOSE = flag.Bool("verbose", true, "Verbose logging")

var r = mux.NewRouter()

type URLList [][2]string

func (a URLList) Len() int           { return len(a) }
func (a URLList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a URLList) Less(i, j int) bool { return len(a[i][0]) > len(a[j][0]) }

type ProxyHandler struct {
	Data          string
	TargetHandler http.Handler
}

func (v *ProxyHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	v.TargetHandler.ServeHTTP(resp, req)
}

func main() {
	flag.Parse()
	fmt.Println("Listen:", *LISTEN)
	fmt.Println("Config JSON to read:", *CONFIG)
	fmt.Println("Rewrite Host?:", *REWRITE_HOST)
	fmt.Println("Verbose?:", *VERBOSE)
	var urlMaps map[string]string = make(map[string]string)
	raw, err := ioutil.ReadFile(*CONFIG)
	if err != nil {
		fmt.Println(*CONFIG, "can't be read. Only reading command line prefix:dest formats")
	} else {
		json.Unmarshal(raw, &urlMaps)
	}

	otherArgs := flag.Args()
	for _, nextArg := range otherArgs {
		idx := strings.Index(nextArg, ":")
		if idx >= 0 {
			prefix, dest := nextArg[:idx], nextArg[idx+1:]
			urlMaps[prefix] = dest
		}
	}
	urlList := make([][2]string, len(urlMaps))

	counter := 0
	for key, value := range urlMaps {
		urlList[counter] = [2]string{key, value}
		counter = counter + 1
	}
	sort.Sort(URLList(urlList))

	mappedCount := 0
	for _, nextMapping := range urlList {
		key, value := nextMapping[0], nextMapping[1]
		mapURL(key, value)
		mappedCount = mappedCount + 1
	}

	if mappedCount <= 0 {
		log.Fatal("No URL mapping specified.")
	}
	log.Fatal(http.ListenAndServe(*LISTEN, r))
}

func mapURL(prefix, dest string) {
	remote, err := url.Parse(dest)
	if err != nil {
		fmt.Println("Invalid url:", dest)
		return
	}
	proxyObj := httputil.NewSingleHostReverseProxy(remote)
	defaultDirector := proxyObj.Director

	thisDomain := ""
	if *REWRITE_HOST {
		startIndex := strings.Index(dest, "//")

		if startIndex >= 0 {
			trimmed := dest[startIndex+2:]
			endIndex := strings.Index(trimmed, "/")
			if endIndex >= 0 {
				thisDomain = trimmed[:endIndex]
			} else {
				thisDomain = trimmed
			}
		}
		fmt.Println(prefix, "=>", dest, fmt.Sprintf("Host = [%s]", thisDomain))
	} else {
		fmt.Println(prefix, "=>", dest, "Host = [Default]")
	}

	proxyObj.Director = func(request *http.Request) {
		if *REWRITE_HOST && thisDomain != "" {
			request.Host = thisDomain
		}
		if *VERBOSE {
			fmt.Println(request.Method, request.URL, request.Host, fmt.Sprintf("=> [%s]", dest))
		}
		defaultDirector(request)
	}

	r.PathPrefix(prefix).Handler(&ProxyHandler{dest, proxyObj})
}
