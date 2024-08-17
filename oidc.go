package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

var openid_configuration struct {
	AuthorizationEndpoint             string     `json:"authorization_endpoint"`
	TokenEndpoint                     string     `json:"token_endpoint"`
	TokenEndpointAuthMethodsSupported ([]string) `json:"token_endpoint_auth_methods_supported"`
	JwksUri                           string     `json:"jwks_uri"`
	UserinfoEndpoint                  string     `json:"userinfo_endpoint"`
}

var openid_keyfunc keyfunc.Keyfunc

type msclaims_t struct {
	/* TODO: These may be non-portable Microsoft attributes */
	Name  string `json:"name"`  /* Scope: profile */
	Email string `json:"email"` /* Scope: email   */
	jwt.RegisteredClaims
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
func get_openid_config(endpoint string) {
	resp, err := http.Get(endpoint + "/.well-known/openid-configuration")
	e(err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatal(fmt.Sprintf(
			"Got response code %d from openid-configuration\n",
			resp.StatusCode,
		))
	}
	err = json.NewDecoder(resp.Body).Decode(&openid_configuration)
	e(err)

	resp, err = http.Get(openid_configuration.JwksUri)
	e(err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatal(fmt.Sprintf(
			"Got response code %d from JwksUri\n",
			resp.StatusCode,
		))
	}

	if config.Openid.Authorize != "" {
		openid_configuration.AuthorizationEndpoint = config.Openid.Authorize
	}

	jwks_json, err := io.ReadAll(resp.Body)
	e(err)

	/*
	 * TODO: The key set is never updated, which is technically incorrect.
	 * We could use keyfunc's auto-update mechanism, but I'd prefer
	 * controlling when to do it manually. Remember to wrap it around a
	 * mutex or some semaphores though.
	 */
	openid_keyfunc, err = keyfunc.NewJWKSetJSON(jwks_json)
	e(err)
}

func generate_authorization_url() string {
	/*
	 * TODO: Handle nonces and anti-replay. Incremental nonces would be
	 * nice on memory and speed (depending on how maps are implemented in
	 * Go, hopefully it's some sort of btree), but that requires either
	 * hacky atomics or having a multiple goroutines to handle
	 * authentication, neither of which are desirable.
	 */
	nonce := random(30)
	return fmt.Sprintf(
		"%s?client_id=%s&response_type=id_token&redirect_uri=%s%s&response_mode=form_post&scope=openid+profile+email&nonce=%s",
		openid_configuration.AuthorizationEndpoint,
		config.Openid.Client,
		config.Url,
		config.Openid.Redirect,
		nonce,
	)
}

func handle_oidc(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(405)
		w.Write([]byte("Error: The OpenID Connect authorization endpoint only accepts POST requests.\n"))
		return
	}

	err := req.ParseForm()
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte("Error: Malformed form data.\n"))
		return
	}

	returned_error := req.PostFormValue("error")
	if returned_error != "" {
		returned_error_description := req.PostFormValue("error_description")
		if returned_error_description == "" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf(
				"Error: The OpenID Connect callback returned an error %s, but did not provide an error_description.\n",
				returned_error,
			)))
			return
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf(
				"Error: The OpenID Connect callback returned an error:\n\n%s\n\n%s\n",
				returned_error,
				returned_error_description,
			)))
			return
		}
	}

	id_token_string := req.PostFormValue("id_token")
	if id_token_string == "" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Error: The OpenID Connect callback did not return an error, but no id_token was found.\n")))
		return
	}

	fmt.Println(id_token_string)

	token, err := jwt.ParseWithClaims(
		id_token_string,
		&msclaims_t{},
		openid_keyfunc.Keyfunc,
	)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Error: Error parsing JWT with custom claims.\n")))
		return
	}

	switch {
	case token.Valid:
		break
	case errors.Is(err, jwt.ErrTokenMalformed):
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Error: Malformed JWT token.\n")))
		return
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Error: Invalid signature on JWT token.\n")))
		return
	case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Error: JWT token expired or not yet valid.\n")))
		return
	default:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Error: Funny JWT token.\n")))
		return
	}

	claims, claims_ok := token.Claims.(*msclaims_t)

	if !claims_ok {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Error: JWT token's claims are not OK.\n")))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("Name: %s\nEmail: %s\nSubject: %s\n", claims.Name, claims.Email, claims.Subject)))
	return

}
