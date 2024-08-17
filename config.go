package main

import (
	"bufio"
	"os"

	"git.sr.ht/~emersion/go-scfg"
)

/*
 * config.Openid.Authorize doesn't have to be specified. But if it is
 * specified, it replaces the once obtained from the
 * .well-known/openid-configuration endpoint. Because Microsoft doesn't seem to
 * want to put their v2.0 authorization endpoint anywhere in their
 * configuration.
 */

var config_with_pointers struct {
	Listen *string `scfg:"listen"`
	Url    *string `scfg:"url"`
	Openid struct {
		Client    *string `scfg:"client"`
		Endpoint  *string `scfg:"endpoint"`
		Authorize *string `scfg:"authorize"`
		Redirect  *string `scfg:"redirect"`
	} `scfg:"openid"`
}

var config struct {
	Listen string
	Url    string
	Openid struct {
		Client    string
		Endpoint  string
		Authorize string
		Redirect  string
	}
}

func fbfp_get_config(path string) {
	f, err := os.Open(path)
	e(err)

	err = scfg.NewDecoder(bufio.NewReader(f)).Decode(&config_with_pointers)
	e(err)

	/*
	 * TODO: We segfault when there are missing configuration options.
	 * There should be better ways to handle this.
	 */
	config.Listen = *(config_with_pointers.Listen)
	config.Url = *(config_with_pointers.Url)
	config.Openid.Client = *(config_with_pointers.Openid.Client)
	config.Openid.Endpoint = *(config_with_pointers.Openid.Endpoint)
	config.Openid.Redirect = *(config_with_pointers.Openid.Redirect)

	if config_with_pointers.Openid.Authorize != nil {
		config.Openid.Authorize = *(config_with_pointers.Openid.Authorize)
	}
}
