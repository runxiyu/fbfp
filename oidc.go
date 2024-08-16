package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type openid_configuration_response_t struct {
	AuthorizationEndpoint             *string     `json:"authorization_endpoint"`
	TokenEndpoint                     *string     `json:"token_endpoint"`
	TokenEndpointAuthMethodsSupported *([]string) `json:"token_endpoint_auth_methods_supported"`
	JwksUri                           *string     `json:"jwks_uri"`
	UserinfoEndpoint                  *string     `json:"userinfo_endpoint"`
}

/*
 * Fetch the OpenID Connect configuration. The endpoint specified in
 * the configuration is incomplete and we fetch the OpenID Connect
 * configuration from the well-known endpoint.
 * This seems to be supported by many authentication providers.
 * The following work, as config.Openid.Endpoint:
 * - https://login.microsoftonline.com/common
 * - https://accounts.google.com/.well-known/openid-configuration
 */
func get_openid_config(endpoint string) openid_configuration_response_t {
	var o openid_configuration_response_t
	resp, err := http.Get(endpoint + "/.well-known/openid-configuration")
	e(err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatal(fmt.Sprintf(
			"Got response code %d from openid-configuration\n",
			resp.StatusCode,
		))
	}
	err = json.NewDecoder(resp.Body).Decode(&o)
	e(err)
	return o
}
