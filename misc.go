package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func e(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// Generate a random url-safe string.
// Note that the "len" parameter specifies the number of bytes taken from the
// random source divided by three and does NOT represent the length of the
// encoded string.
func random(len int) string {
	r := make([]byte, 3*len)
	rand.Read(r)
	return base64.RawURLEncoding.EncodeToString(r)
}
