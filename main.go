package main

import (
	"flag"
	"net/http"

	_ "net/http/pprof"

	_ "github.com/haha4github/trojan-go/component"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/option"
)

func main() {
	go func() {
		// 开启pprof，默认端口为http://localhost:6060/debug/pprof/
		http.ListenAndServe("localhost:6060", nil)
	}()

	flag.Parse()
	for {
		h, err := option.PopOptionHandler()
		if err != nil {
			log.Fatal("invalid options")
		}
		err = h.Handle()
		if err == nil {
			break
		}
	}
}
