package main

import (
	"fmt"
	"log"
	"net/http"
)

func handler_index(w http.ResponseWriter, req *http.Request) {
	openid_authorization_url := generate_authorization_url(
		config.Openid.Client,
		config.Openid.RedirectUri,
	)

	fmt.Println(openid_authorization_url)
}

func main() {
	fbfp_get_config("fbfp.scfg")

	get_openid_config(config.Openid.Endpoint)

	http.HandleFunc("/", handler_index)

	log.Printf("Listening on %s\n", config.Listen)
	http.ListenAndServe(config.Listen, nil)
}
