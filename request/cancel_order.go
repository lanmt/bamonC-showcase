package request

import (
	"bamonC/model"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func CancelOrder(user *model.User, tradeNo string) error {

	dataMap := map[string]interface{}{
		"venueTradeNo": tradeNo,
		"remark":       "",
	}
	dataJSONBytes, _ := json.Marshal(dataMap)
	dataJSON := string(dataJSONBytes)

	timestamp := getTimestamp()

	sign, err := Sign(timestamp, upstream.CancellationPath, dataJSON)
	if err != nil {
		return err
	}

	form := url.Values{}
	for k, v := range dataMap {
		form.Set(k, fmt.Sprintf("%v", v))
	}

	req, err := http.NewRequest("POST", upstreamURL(upstream.CancellationPath), bytes.NewBufferString(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("app-key", upstream.AppKey)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("sign", sign)
	req.Header.Set(upstream.AuthHeader, user.Auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "en,zh-CN;q=0.9,zh;q=0.8,en-GB;q=0.7,en-US;q=0.6")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic("defer Body.Close err:" + err.Error())
		}
	}(resp.Body)

	return nil
}
