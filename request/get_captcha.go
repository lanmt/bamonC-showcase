package request

import (
	"bamonC/model"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/tidwall/gjson"
)

// GetCaptcha /*
/*
params:
	user: 用户
	username: 该用户的username
return:
	验证码响应，string json格式
	error
*/
func GetCaptcha(user *model.User, username string) (string, error) {
	point := user.Point
	auth := user.Auth

	timestamp := getTimestamp()
	paramsMap := map[string]interface{}{
		"captchaType": "clickWord",
		"clientUid":   point,
		"ts":          timestamp,
		"nocache":     timestamp,
	}
	paramsJson, err := json.Marshal(paramsMap)
	if err != nil {
		return "", err
	}

	sign, err := Sign(timestamp, upstream.CaptchaGetPath, string(paramsJson))
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Add("captchaType", "clickWord")
	params.Add("clientUid", point)
	params.Add("ts", timestamp)
	params.Add("nocache", timestamp)

	// 构造请求
	apiUrl := upstreamURL(upstream.CaptchaGetPath)
	reqUrl := fmt.Sprintf("%s?%s", apiUrl, params.Encode())

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("app-key", upstream.AppKey)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("sign", sign)
	req.Header.Set(upstream.AuthHeader, auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "en,zh-CN;q=0.9")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic("defer Body.Close err:" + err.Error())
		}
	}(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	data := gjson.Get(string(body), "data").String()
	if gjson.Get(data, "repData.originalImageBase64").String() == "" {
		return "", errors.New(data)
	}
	return data, nil
}
