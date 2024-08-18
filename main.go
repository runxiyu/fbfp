package main

import (
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
)

func handle_index(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("e"))
}

func handle_login(w http.ResponseWriter, req *http.Request) {
	openid_authorization_url := generate_authorization_url()

	http.Redirect(w, req, openid_authorization_url, 303)
}

func main() {
	fbfp_get_config("fbfp.scfg")

	log.Printf("Registering handlers\n")
	http.HandleFunc("/", handle_index)
	http.HandleFunc("/login", handle_login)
	http.HandleFunc(config.Openid.Redirect, handle_oidc)

	log.Printf(
		"Establishing listener for net %s, addr %s\n",
		config.Net,
		config.Addr,
	)

	l, err := net.Listen(config.Net, config.Addr)
	e(err)

	log.Printf("Fetching OpenID Connect configuration\n")
	get_openid_config(config.Openid.Endpoint)

	if config.Proto == "http" {
		log.Printf("Serving http\n")
		err = http.Serve(l, nil)
	} else if config.Proto == "fcgi" {
		log.Printf("Serving fcgi\n")
		err = fcgi.Serve(l, nil)
	}
	e(err)
}
