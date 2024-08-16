package main

import (
	"fmt"
)

func main() {
	fbfp_get_config("fbfp.scfg")

	fmt.Println(config)

	openid_configuration_response := get_openid_config(config.Openid.Endpoint)

	openid_authorization_url := generate_authorization_url(
		*(openid_configuration_response.AuthorizationEndpoint),
		config.Openid.Client,
		config.Openid.RedirectUri,
	)

	fmt.Println(openid_authorization_url)
}
