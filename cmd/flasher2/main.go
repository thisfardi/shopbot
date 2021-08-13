package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spacysky322/flasher2-src/internal/auth"
	"github.com/spacysky322/flasher2-src/internal/flasher"
	"github.com/spacysky322/flasher2-src/internal/util"
)

const qrFilename = "qr.png"

func login() {
	qrlogin := auth.NewQrLogin()
	data, id, err := qrlogin.FetchQrCode()
	if err != nil {
		panic(err)
	}
	if err = ioutil.WriteFile(qrFilename, data, 0644); err != nil {
		panic(err)
	}
	if err = util.OpenFile(qrFilename); err != nil {
		panic(err)
	}
	err = qrlogin.WaitScanned(id)
	if err != nil {
		panic(err)
	}
	fmt.Println("Qr code scanned, Please confirm")
	qrcodeToken, err := qrlogin.WaitConfirmed(id)
	if err != nil {
		panic(err)
	}
	fmt.Println("Qr confirmed")
	cookies, err := qrlogin.Login(qrcodeToken)
	if err != nil {
		panic(err)
	}
	serialize(cookies)
}

func main() {
	if _, err := os.Stat(cookieFilename); os.IsNotExist(err) {
		login()
	}
	cookies, err := deserialize()
	if err != nil {
		panic(err)
	}
	f, err := flasher.NewFlasher(cookies)
	if err != nil {
		panic(err)
	}
	itemid, shopid, err := f.ParseUrl("https://shopee.co.id/Hawaii-Ineza-Kotak-Sampah-Segi-i.120737738.1879166523")
	if err != nil {
		panic(err)
	}
	item, err := f.FetchItem(itemid, shopid)
	if err != nil {
		panic(err)
	}
	shippingInfo, err := f.FetchShippingInfo(item)
	if err != nil {
		panic(err)
	}
	for i, info := range shippingInfo {
		fmt.Println(i, info.Name())
	}
	selectedShipping := shippingInfo[iinput("select: ", "pls int")]
	cartItem, err := f.AddToCart(item, 0)
	if err != nil {
		panic(err)
	}
	err = f.Checkout(cartItem, selectedShipping, &flasher.Payment{
		Name:      "Alfamart",
		ChannelId: 8003200,
		Option:    "",
		Version:   2,
		TxnFee:    250000000,
	})
	if err != nil {
		panic(err)
	}
}
