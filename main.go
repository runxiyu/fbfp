package main

import (
	"fmt"
)

func main() {
	fbfp_get_config("fbfp.scfg")

	fmt.Println(config)

	openid_configuration_response := get_openid_config(config.Openid.Endpoint)

	/*
	 * TODO: Check what this is supposed to do at all
	 */
	state := random(20)

	/*
	 * TODO: Handle nonces and anti-replay. Incremental nonces would be
	 * nice on memory and speed (depending on how maps are implemented in
	 * Go, hopefully it's some sort of btree), but that requires either
	 * hacky atomics or having a multiple goroutines to handle
	 * authentication, neither of which are desirable.
	 */
	nonce := random(20)

	fmt.Println(fmt.Sprintf(
		"%s?client_id=%s&response_type=id_token&redirect_uri=%s&response_mode=form_post&scope=openid&state=%s&nonce=%s",
		*(openid_configuration_response.AuthorizationEndpoint),
		config.Openid.Client,
		config.Openid.RedirectUri,
		state,
		nonce,
	))
}
