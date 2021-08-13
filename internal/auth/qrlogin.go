package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spacysky322/flasher2-src/internal/constants"
	"github.com/spacysky322/flasher2-src/internal/transportwrapper"
	"github.com/spacysky322/flasher2-src/internal/util"
)

const (
	statusNew       = "NEW"       // new qr code
	statusScanned   = "SCANNED"   // the user has scanned the qr code
	statusConfirmed = "CONFIRMED" // the user has confirmed
	statusExpired   = "EXPIRED"   // qr code expired
	csrftokenLength = 32
	waitDuration    = 1 * time.Second
)

type QrLogin struct {
	client *http.Client // used for all requests
	rt     *transportwrapper.TransportWrapper
}

func (login *QrLogin) FetchQrCode() (data []byte, id string, err error) {
	req, err := http.NewRequest("GET", "https://shopee.co.id/api/v2/authentication/gen_qrcode", nil)
	if err != nil {
		return
	}
	resp, err := login.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	jsondata, err := util.JsonDecodeResp(resp)
	if err != nil {
		return
	}
	jsondata = jsondata["data"].(map[string]interface{})
	data, err = base64.StdEncoding.DecodeString(jsondata["qrcode_base64"].(string))
	id = jsondata["qrcode_id"].(string)
	return
}

// wait until qr code scanned
func (login *QrLogin) WaitScanned(id string) error {
	_, err := login.wait(id, statusScanned)
	return err
}

// wait until user confirmed
func (login *QrLogin) WaitConfirmed(id string) (string, error) {
	return login.wait(id, statusConfirmed)
}

// login using token, return the cookie
func (login *QrLogin) Login(qrcodeToken string) (cookies map[string]*http.Cookie, err error) {
	req, err := http.NewRequest("POST", "https://shopee.co.id/api/v2/authentication/qrcode_login", util.JsonEncodeBuff(
		map[string]interface{}{
			"qrcode_token": qrcodeToken,
			"support_ivs":  true,
		},
	))
	if err != nil {
		return
	}
	resp, err := login.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	jsondata, err := util.JsonDecodeResp(resp)
	if err != nil {
		return
	}

	if errcode := jsondata["error"]; errcode != nil {
		code, _ := errcode.(json.Number).Int64()
		err = fmt.Errorf("error: %d", code)
		return
	}

	return login.rt.Cookies, nil
}

func (login *QrLogin) wait(id, status string) (qrcodeToken string, err error) {
	for currStatus := statusNew; err == nil && currStatus != status; currStatus, qrcodeToken, err = login.getStatus(id) {
		time.Sleep(waitDuration)
	}
	return
}

func (login *QrLogin) getStatus(id string) (status string, qrcodeToken string, err error) {
	req, err := http.NewRequest("GET", "https://shopee.co.id/api/v2/authentication/qrcode_status", nil)
	if err != nil {
		return
	}
	q := req.URL.Query()
	q.Add("qrcode_id", id)
	req.URL.RawQuery = q.Encode()
	resp, err := login.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	jsondata, err := util.JsonDecodeResp(resp)
	if err != nil {
		return
	}

	if errcode := jsondata["error"]; errcode != nil {
		code, _ := errcode.(json.Number).Int64()
		err = fmt.Errorf("error: %d", code)
		return
	}

	// check status
	jsondata = jsondata["data"].(map[string]interface{})
	status = jsondata["status"].(string)
	switch status {
	case statusConfirmed:
		qrcodeToken = jsondata["qrcode_token"].(string)
	case statusExpired:
		err = fmt.Errorf("qr code expired")
	}

	return
}

// fetch initial cookie
func (login *QrLogin) init() (err error) {
	req, err := http.NewRequest("POST", "https://shopee.co.id/buyer/login", nil)
	if err != nil {
		return
	}
	resp, err := login.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return
}

func NewQrLogin() *QrLogin {
	login := &QrLogin{}
	client := &http.Client{}
	rt := transportwrapper.Wrap(client.Transport)
	token := util.RandomString(csrftokenLength)

	rt.Set("Referer", "https://shopee.co.id")
	rt.Set("If-None-Match-", "*")
	rt.Set("Content-Type", "application/json")
	rt.Set("User-Agent", constants.UserAgent)
	rt.Set("X-Csrftoken", token)
	rt.Set("X-Api-Source", "pc")
	rt.SetCookie(&http.Cookie{Name: "csrftoken", Value: token})

	client.Transport = rt
	login.client = client
	login.rt = rt
	login.init()
	return login
}
