package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// generate random string
func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// decode json from []byte
func JsonDecode(data []byte) (result map[string]interface{}, err error) {
	d := json.NewDecoder(bytes.NewBuffer(data))
	d.UseNumber()
	err = d.Decode(&result)
	return
}

// decode json from response
func JsonDecodeResp(resp *http.Response) (result map[string]interface{}, err error) {
	d := json.NewDecoder(resp.Body)
	d.UseNumber()
	err = d.Decode(&result)
	return
}

// get int64 from json.Number, does not return an error, but panic if error not nil
func MustInt64(jsonNumber interface{}) int64 {
	num, err := jsonNumber.(json.Number).Int64()
	if err != nil {
		panic(err)
	}
	return num
}

// encode json to *bytes.Buffer
func JsonEncodeBuff(data map[string]interface{}) *bytes.Buffer {
	res, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return bytes.NewBuffer(res)
}

// open file.
// must use Termux on android, or it won't work
func OpenFile(filename string) error {
	var cmd *exec.Cmd
	switch name := runtime.GOOS; name {
	case "android", "linux":
		// must use Termux on android, or it won't work
		cmd = exec.Command("xdg-open", filename)
	case "windows":
		cmd = exec.Command("start", filename)
	default:
		return fmt.Errorf("unable to open file, unknown os: %s", name)
	}
	return cmd.Start()
}

// just a simple assert function to save lines of code
func Assert(cond bool) {
	if !cond {
		panic("assertion error")
	}
}
