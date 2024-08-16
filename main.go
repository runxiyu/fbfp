package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"git.sr.ht/~emersion/go-scfg"
)

var config_with_pointers struct {
	Listen struct {
		Port *int    `scfg:"port"`
		Bind *string `scfg:"bind"`
	} `scfg:"listen"`
	Openid struct {
		Client      *string `scfg:"client"`
		Secret      *string `scfg:"secret"`
		Endpoint    *string `scfg:"endpoint"`
		RedirectUri *string `scfg:"redirect_uri"`
	} `scfg:"openid"`
}

var config struct {
	Listen struct {
		Port int
		Bind string
	}
	Openid struct {
		Client      string
		Secret      string
		Endpoint    string
		RedirectUri string
	}
}

type openid_configuration_response_t struct {
	AuthorizationEndpoint             *string     `json:"authorization_endpoint"`
	TokenEndpoint                     *string     `json:"token_endpoint"`
	TokenEndpointAuthMethodsSupported *([]string) `json:"token_endpoint_auth_methods_supported"`
	JwksUri                           *string     `json:"jwks_uri"`
	UserinfoEndpoint                  *string     `json:"userinfo_endpoint"`
}

func main() {
	f, err := os.Open("fbfp.scfg")
	e(err)

	err = scfg.NewDecoder(bufio.NewReader(f)).Decode(&config_with_pointers)
	e(err)

	/*
	 * TODO: We segfault when there are missing configuration options.
	 * There should be better ways to handle this.
	 */
	config.Listen.Port = *(config_with_pointers.Listen.Port)
	config.Listen.Bind = *(config_with_pointers.Listen.Bind)
	config.Openid.Client = *(config_with_pointers.Openid.Client)
	config.Openid.Endpoint = *(config_with_pointers.Openid.Endpoint)
	config.Openid.Secret = *(config_with_pointers.Openid.Secret)
	config.Openid.RedirectUri = *(config_with_pointers.Openid.RedirectUri)

	fmt.Println(config)

	/*
	 * Fetch the OpenID Connect configuration. The endpoint specified in
	 * the configuration is incomplete and we fetch the OpenID Connect
	 * configuration from the well-known endpoint.
	 * This seems to be supported by many authentication providers.
	 * The following work, as config.Openid.Endpoint:
	 * - https://login.microsoftonline.com/common
	 * - https://accounts.google.com/.well-known/openid-configuration
	 */
	var openid_configuration_response openid_configuration_response_t
	resp, err := http.Get(config.Openid.Endpoint + "/.well-known/openid-configuration")
	e(err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatal(fmt.Sprintf(
			"Got response code %d from openid-configuration\n",
			resp.StatusCode,
		))
	}
	err = json.NewDecoder(resp.Body).Decode(&openid_configuration_response)
	e(err)

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
