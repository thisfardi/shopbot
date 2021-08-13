package flasher

import (
	"fmt"
	"net/http"

	"github.com/spacysky322/flasher2-src/internal/util"
)

func getUser(cookies map[string]*http.Cookie) (_ User, err error) {
	client := http.DefaultClient
	req, err := http.NewRequest("GET", "https://mall.shopee.co.id/api/v1/account_info", nil)
	if err != nil {
		return
	}
	req.Header.Add("Referer", "https://mall.shopee.co.id")
	req.Header.Add("If-None-Match-", "*")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := util.JsonDecodeResp(resp)
	if err != nil {
		return
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("invalid cookie")
	}

	return data, nil
}
