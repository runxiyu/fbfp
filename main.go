package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"git.sr.ht/~emersion/go-scfg"
)

var config_with_pointers struct {
	Listen struct {
		Port *int    `scfg:"port"`
		Bind *string `scfg:"bind"`
	} `scfg:"listen"`
}

var config struct {
	Listen struct {
		Port int
		Bind string
	}
}

func e(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func main() {
	f, err := os.Open("fbfp.scfg")
	e(err)

	err = scfg.NewDecoder(bufio.NewReader(f)).Decode(&config_with_pointers)
	e(err)

	config.Listen.Port = *(config_with_pointers.Listen.Port)
	config.Listen.Bind = *(config_with_pointers.Listen.Bind)

	fmt.Println(config)
}
