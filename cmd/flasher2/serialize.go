package main

import (
	"encoding/gob"
	"net/http"
	"os"
)

const cookieFilename = "cookie.ser"

func serialize(cookie map[string]*http.Cookie) (err error) {
	f, err := os.Create(cookieFilename)
	if err != nil {
		return
	}
	encoder := gob.NewEncoder(f)
	if err = encoder.Encode(cookie); err != nil {
		return
	}
	f.Close()
	return
}

func deserialize() (cookies map[string]*http.Cookie, err error) {
	cookies = map[string]*http.Cookie{}
	f, err := os.Open(cookieFilename)
	if err != nil {
		return
	}
	decoder := gob.NewDecoder(f)
	if err = decoder.Decode(&cookies); err != nil {
		return
	}
	f.Close()
	return
}
