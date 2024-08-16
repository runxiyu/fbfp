package main

import (
	"bufio"
	"fmt"
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

func main() {
	f, err := os.Open("fbfp.scfg")
	e(err)

	err = scfg.NewDecoder(bufio.NewReader(f)).Decode(&config_with_pointers)
	e(err)

	config.Listen.Port = *(config_with_pointers.Listen.Port)
	config.Listen.Bind = *(config_with_pointers.Listen.Bind)
	config.Openid.Client = *(config_with_pointers.Openid.Client)
	config.Openid.Endpoint = *(config_with_pointers.Openid.Endpoint)
	config.Openid.Secret = *(config_with_pointers.Openid.Secret)
	config.Openid.RedirectUri = *(config_with_pointers.Openid.RedirectUri)

	fmt.Println(config)

}
