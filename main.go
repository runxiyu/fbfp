package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var tmpl *template.Template

func handle_index(w http.ResponseWriter, req *http.Request) {
	tmpl.ExecuteTemplate(w, "index", nil)
}

func handle_login(w http.ResponseWriter, req *http.Request) {
	/* TODO: Invalidate current session */
	openid_authorization_url := generate_authorization_url()
	http.Redirect(w, req, openid_authorization_url, 303)
}

func main() {
	var err error
	fbfp_get_config("fbfp.scfg")

	log.Printf("Opening database\n")
	switch config.Db.Type {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(config.Db.Conn), &gorm.Config{})
		e(err)
	default:
		e(fmt.Errorf("Database type %s unsupported", config.Db.Type))
	}

	log.Printf("Setting up templates\n")
	tmpl, err = template.ParseGlob("tmpl/*")
	e(err)

	log.Printf("Registering handlers\n")
	http.HandleFunc("/{$}", handle_index)
	http.HandleFunc("/login", handle_login)
	http.HandleFunc(config.Openid.Redirect, handle_oidc)

	log.Printf(
		"Establishing listener for net %s, addr %s\n",
		config.Listen.Net,
		config.Listen.Addr,
	)
	l, err := net.Listen(config.Listen.Net, config.Listen.Addr)
	e(err)

	log.Printf("Fetching OpenID Connect configuration\n")
	get_openid_config(config.Openid.Endpoint)

	if config.Listen.Proto == "http" {
		log.Printf("Serving http\n")
		err = http.Serve(l, nil)
	} else if config.Listen.Proto == "fcgi" {
		log.Printf("Serving fcgi\n")
		err = fcgi.Serve(l, nil)
	}
	e(err)
}
