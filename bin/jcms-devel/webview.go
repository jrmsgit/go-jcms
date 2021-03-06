package main

import (
	"net/url"

	"github.com/jrmsdev/go-jcms/lib/jcms"
	xwv "github.com/zserge/webview"
)

const (
	webviewResize = true
	webviewWidth  = 800
	webviewHeight = 600
)

func Webview(addr, req string) {
	uri, err := url.Parse(jcms.Listen(addr))
	if err != nil {
		log.Panic(err.Error())
	}
	go func() {
		jcms.Serve()
	}()
	defer jcms.Stop()
	uri.Path = req
	log.D("open %s", uri.String())
	xwv.Open("JCMS Devel", uri.String(),
		webviewWidth, webviewHeight, webviewResize)
}
