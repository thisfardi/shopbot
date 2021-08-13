package main

import (
	"fmt"
	"io/ioutil"

	"github.com/spacysky322/flasher2-src/internal/auth"
	"github.com/spacysky322/flasher2-src/internal/flasher"
	"github.com/spacysky322/flasher2-src/internal/util"
)

// temporary file name
const qrFilename = "qr.png"

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// create qr login instance
	login := auth.NewQrLogin()
	// fetch new qr code from server
	data, id, err := login.FetchQrCode()
	check(err)
	// save the fetched qr code to file
	err = ioutil.WriteFile(qrFilename, data, 0644)
	check(err)
	// open the file
	err = util.OpenFile(qrFilename)
	check(err)
	// wait until user scans the qr code
	err = login.WaitScanned(id)
	check(err)
	fmt.Println("QR Code scanned, please confirm")
	// wait until user confirmed.
	// qrToken is what we are looking for
	qrToken, err := login.WaitConfirmed(id)
	check(err)
	fmt.Println("User confirmed")
	// finally, login using qr token
	cookies, err := login.Login(qrToken)
	check(err)
	// you can now login to your shopee account with flasher
	flsher, err := flasher.NewFlasher(cookies)
	check(err)
	user := flsher.GetUser()
	fmt.Println("Hi,", user.Username())
}
