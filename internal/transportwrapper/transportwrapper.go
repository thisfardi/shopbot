package transportwrapper

import (
	"net/http"
)

// allows setting default values to request header, and saving cookies/loading them
type TransportWrapper struct {
	http.Header
	// using map[string]Cookie to avoid duplicate, string(the key) is always equals to Cookie.Name
	Cookies map[string]*http.Cookie
	rt      http.RoundTripper
}

func Wrap(rt http.RoundTripper) *TransportWrapper {
	if rt == nil {
		rt = http.DefaultTransport
	}
	return &TransportWrapper{Header: make(http.Header), Cookies: make(map[string]*http.Cookie), rt: rt}
}

func (w *TransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range w.Header {
		req.Header[k] = v
	}
	for _, cookie := range w.Cookies {
		req.AddCookie(cookie)
	}

	resp, err := w.rt.RoundTrip(req)
	if err == nil {
		w.receiveCookie(resp)
	}
	return resp, err
}

func (w *TransportWrapper) SetCookie(cookie *http.Cookie) {
	w.Cookies[cookie.Name] = cookie
}

func (w *TransportWrapper) receiveCookie(resp *http.Response) {
	for _, cookie := range resp.Cookies() {
		w.Cookies[cookie.Name] = cookie
	}
}
