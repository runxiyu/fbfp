/*
 * Handle fbfp's configuration.
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
	Prod   *bool   `scfg:"prod"`
	Tmpl   *string `scfg:"tmpl"`
	Static *bool   `scfg:"static"`
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
	} `scfg:"openid"`
}

var config struct {
	Url    string
	Prod   bool
	Tmpl   string
	Static bool
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
	config.Prod = *(config_with_pointers.Prod)
	config.Tmpl = *(config_with_pointers.Tmpl)
	config.Static = *(config_with_pointers.Static)
	config.Listen.Addr = *(config_with_pointers.Listen.Addr)
	config.Listen.Net = *(config_with_pointers.Listen.Net)
	config.Listen.Proto = *(config_with_pointers.Listen.Proto)
	config.Db.Type = *(config_with_pointers.Db.Type)
	config.Db.Conn = *(config_with_pointers.Db.Conn)
	config.Openid.Client = *(config_with_pointers.Openid.Client)
	config.Openid.Endpoint = *(config_with_pointers.Openid.Endpoint)

	if config_with_pointers.Openid.Authorize != nil {
		config.Openid.Authorize =
			*(config_with_pointers.Openid.Authorize)
	}
}
