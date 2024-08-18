/*
 * OpenID Connect for fbfp.
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

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
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JwksUri               string `json:"jwks_uri"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
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
		openid_configuration.AuthorizationEndpoint =
			config.Openid.Authorize
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
		"%s"+
			"?client_id=%s"+
			"&response_type=id_token"+
			"&redirect_uri=%s/oidc"+
			"&response_mode=form_post"+
			"&scope=openid+profile+email"+
			"&nonce=%s",
		openid_configuration.AuthorizationEndpoint,
		config.Openid.Client,
		config.Url,
		nonce,
	)
}

func handle_oidc(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(405)
		w.Write([]byte(
			"Error\n" +
				"Only POST is allowed on the OIDC callback.\n" +
				"Please return to the login page and retry.\n",
		))
		return
	}

	err := req.ParseForm()
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf(
			"Error\n"+
				"Malformed form data.\n%s\n",
			err,
		)))
		return
	}

	returned_error := req.PostFormValue("error")
	if returned_error != "" {
		returned_error_description :=
			req.PostFormValue("error_description")
		if returned_error_description == "" {
			w.Header().Set(
				"Content-Type",
				"text/plain; charset=utf-8",
			)
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf(
				"Error\n%s\n",
				returned_error,
			)))
			return
		} else {
			w.Header().Set(
				"Content-Type",
				"text/plain; charset=utf-8",
			)
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf(
				"Error\n%s\n%s\n",
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
		w.Write([]byte(fmt.Sprintf("Error\nMissing id_token.\n")))
		return
	}

	token, err := jwt.ParseWithClaims(
		id_token_string,
		&msclaims_t{},
		openid_keyfunc.Keyfunc,
	)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf(
			"Error\n"+
				"Cannot parse claims.\n%s\n",
			err,
		)))
		return
	}

	switch {
	case token.Valid:
		break
	case errors.Is(err, jwt.ErrTokenMalformed):
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Error\nMalformed JWT token.\n")))
		return
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Error\nInvalid JWS signature.\n")))
		return
	case errors.Is(err, jwt.ErrTokenExpired) ||
		errors.Is(err, jwt.ErrTokenNotValidYet):
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf(
			"Error\n" +
				"JWT token expired or not yet valid.\n",
		)))
		return
	default:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Error\nInvalid JWT token.\n")))
		return
	}

	claims, claims_ok := token.Claims.(*msclaims_t)

	if !claims_ok {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("Error\nCannot unpack claims.\n")))
		return
	}

	cookie_value := random(20)

	cookie := http.Cookie{
		Name:     "session",
		Value:    cookie_value,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	var user user_t
	err = db.FirstOrCreate(&user, user_t{Subject: claims.Subject}).Error
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf(
			"Error\nDatabase query failed.\n%s\n",
			err,
		)))
		return
	}
	user.Name = claims.Name
	user.Email = claims.Email
	db.Save(&user)

	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(
		"Name: %s\nEmail: %s\nSubject: %s\nCookie: %s\n",
		claims.Name,
		claims.Email,
		claims.Subject,
		cookie_value,
	)))
	return

}
