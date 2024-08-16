package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"git.sr.ht/~emersion/go-scfg"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

var config_with_pointers struct {
	Listen struct {
		Port *int    `scfg:"port"`
		Bind *string `scfg:"bind"`
	} `scfg:"listen"`
	Msal struct {
		Client   *string     `scfg:"client"`
		Tenant   *string     `scfg:"tenant"`
		Secret   *string     `scfg:"secret"`
		Callback *string     `scfg:"callback"`
		Scopes   *([]string) `scfg:"scopes"`
	} `scfg:"msal"`
}

var config struct {
	Listen struct {
		Port int
		Bind string
	}
	Msal struct {
		Client   string
		Tenant   string
		Secret   string
		Callback string
		Scopes   []string
	}
}

func main() {
	f, err := os.Open("fbfp.scfg")
	e(err)

	err = scfg.NewDecoder(bufio.NewReader(f)).Decode(&config_with_pointers)
	e(err)

	config.Listen.Port = *(config_with_pointers.Listen.Port)
	config.Listen.Bind = *(config_with_pointers.Listen.Bind)
	config.Msal.Client = *(config_with_pointers.Msal.Client)
	config.Msal.Tenant = *(config_with_pointers.Msal.Tenant)
	config.Msal.Secret = *(config_with_pointers.Msal.Secret)
	config.Msal.Callback = *(config_with_pointers.Msal.Callback)
	config.Msal.Scopes = *(config_with_pointers.Msal.Scopes)

	fmt.Println(config)

	cred, err := confidential.NewCredFromSecret(config.Msal.Secret)
	e(err)

	confidential_client, err := confidential.New(
		"https://login.microsoftonline.com/"+config.Msal.Tenant,
		config.Msal.Client,
		cred,
	)
	e(err)

	result, err := confidential_client.AcquireTokenByCredential(
		context.Background(),
		config.Msal.Scopes,
	)
	e(err)

	fmt.Println(result)
}
