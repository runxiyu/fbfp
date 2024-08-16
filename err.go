package main

import "log"

func e(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
