package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"git.sr.ht/~emersion/go-scfg"
)

var config struct {
	Port *int `scfg:"port"`
}

func e(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func main() {
	f, err := os.Open("fbfp.scfg")
	e(err)

	err = scfg.NewDecoder(bufio.NewReader(f)).Decode(&config)
	e(err)

	fmt.Println(*(config.Port))
}
