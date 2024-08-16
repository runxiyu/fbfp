package main

import (
	"bufio"
	"os"

	"git.sr.ht/~emersion/go-scfg"
)

var config_with_pointers struct {
	Listen *string `scfg:"listen"`
	Url    *string `scfg:"url"`
	Openid struct {
		Client   *string `scfg:"client"`
		Secret   *string `scfg:"secret"`
		Endpoint *string `scfg:"endpoint"`
		Redirect *string `scfg:"redirect"`
	} `scfg:"openid"`
}

var config struct {
	Listen string
	Url    string
	Openid struct {
		Client   string
		Secret   string
		Endpoint string
		Redirect string
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
	config.Openid.Secret = *(config_with_pointers.Openid.Secret)
	config.Openid.Redirect = *(config_with_pointers.Openid.Redirect)
}
