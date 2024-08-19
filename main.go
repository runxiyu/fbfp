/*
 * The main logic for fbfp.
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
	"html/template"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
)

var tmpl *template.Template

func main() {
	/*
	 * This seems necessary because in the following sections I need to
	 * assign to global variables like db and tmpl. If I use = there, and
	 * don't declare err here, then err is undeclared; but if I use :=
	 * there, my global variables are treated as local variables.
	 */
	var err error

	fbfp_get_config("fbfp.scfg")

	log.Printf("Setting up database\n")
	e(setup_database())

	log.Printf("Setting up templates\n")
	tmpl = er(template.ParseGlob(config.Tmpl + "/*"))

	if config.Static {
		log.Printf("Registering static handle\n")
		fs := http.FileServer(http.Dir("./static"))
		http.Handle("/static/", http.StripPrefix("/static/", fs))
	}

	log.Printf("Registering handlers\n")
	http.HandleFunc("/{$}", handle_index)
	http.HandleFunc("/oidc", handle_oidc)

	log.Printf("Fetching OpenID Connect configuration\n")
	get_openid_config(config.Openid.Endpoint)

	log.Printf(
		"Establishing listener for net \"%s\", addr \"%s\"\n",
		config.Listen.Net,
		config.Listen.Addr,
	)
	l := er(net.Listen(config.Listen.Net, config.Listen.Addr))

	if config.Listen.Proto == "http" {
		log.Printf("Serving http\n")
		err = http.Serve(l, nil)
	} else if config.Listen.Proto == "fcgi" {
		log.Printf("Serving fcgi\n")
		err = fcgi.Serve(l, nil)
	}
	e(err)
}
