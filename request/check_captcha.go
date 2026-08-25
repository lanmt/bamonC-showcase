package request

import (
	"bamonC/model"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/tidwall/gjson"
)

func CheckCaptcha(user *model.User, coordsJSON string, secretKey, backToken string) (string, error) {
	count := len(gjson.Parse(coordsJSON).Array())
	if count != 3 {
		return "", errors.New("获取的坐标不足3个")
	}
	var ee, pointJson string
	if secretKey != "" {
		encryptedEE, err := D(backToken+"---"+coordsJSON, secretKey)
		if err != nil {
			return "", err
		}
		ee = encryptedEE
		encryptedPoint, err := D(coordsJSON, secretKey)
		if err != nil {
			return "", err
		}
		pointJson = encryptedPoint
	} else {
		ee = backToken + "---" + coordsJSON
		pointJson = coordsJSON
	}

	timestamp := getTimestamp()
	dataMap := map[string]interface{}{
		"captchaType": "clickWord",
		"pointJson":   pointJson,
		"token":       backToken,
	}
	dataJSONBytes, _ := json.Marshal(dataMap)
	dataJSON := string(dataJSONBytes)

	sign, err := Sign(timestamp, upstream.CaptchaCheckPath, dataJSON)
	if err != nil {
		return "", err
	}

	form := url.Values{}
	for k, v := range dataMap {
		form.Set(k, fmt.Sprintf("%v", v))
	}

	req, err := http.NewRequest("POST", upstreamURL(upstream.CaptchaCheckPath), bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", err
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
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic("defer Body.Close err:" + err.Error())
		}
	}(resp.Body)

	body, _ := ioutil.ReadAll(resp.Body)
	success := gjson.GetBytes(body, "data.success").Bool()
	if !success {
		return "", errors.New("验证码未通过")
	}
	return ee, nil
}
