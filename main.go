package main

import (
	"log"
	"net/http"
)

func handle_index(w http.ResponseWriter, req *http.Request) {
	openid_authorization_url := generate_authorization_url()

	http.Redirect(w, req, openid_authorization_url, 303)
}

func main() {
	fbfp_get_config("fbfp.scfg")

	get_openid_config(config.Openid.Endpoint)

	http.HandleFunc("/", handle_index)
	http.HandleFunc(config.Openid.Redirect, handle_oidc)

	log.Printf("Listening on %s\n", config.Listen)
	http.ListenAndServe(config.Listen, nil)
}
