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
	Url    *string `scfg:"url"`
	Listen struct {
		Addr  *string `scfg:"addr"`
		Net   *string `scfg:"net"`
		Proto *string `scfg:"proto"`
	} `scfg:"listen"`
	Db struct {
		Type *string `scfg:"type"`
		Conn *string `scfg:"conn"`
	} `scfg:"db"`
	Openid struct {
		Client    *string `scfg:"client"`
		Endpoint  *string `scfg:"endpoint"`
		Authorize *string `scfg:"authorize"`
		Redirect  *string `scfg:"redirect"`
	} `scfg:"openid"`
}

var config struct {
	Url    string
	Listen struct {
		Addr  string
		Net   string
		Proto string
	}
	Db struct {
		Type string
		Conn string
	}
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
	config.Url = *(config_with_pointers.Url)
	config.Listen.Addr = *(config_with_pointers.Listen.Addr)
	config.Listen.Net = *(config_with_pointers.Listen.Net)
	config.Listen.Proto = *(config_with_pointers.Listen.Proto)
	config.Db.Type = *(config_with_pointers.Db.Type)
	config.Db.Conn = *(config_with_pointers.Db.Conn)
	config.Openid.Client = *(config_with_pointers.Openid.Client)
	config.Openid.Endpoint = *(config_with_pointers.Openid.Endpoint)
	config.Openid.Redirect = *(config_with_pointers.Openid.Redirect)

	if config_with_pointers.Openid.Authorize != nil {
		config.Openid.Authorize =
			*(config_with_pointers.Openid.Authorize)
	}
}
