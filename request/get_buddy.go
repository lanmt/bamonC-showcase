package request

import (
	"bamonC/model"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/tidwall/gjson"
)

func GetBuddies(user *model.User) ([]*model.Buddy, error) {

	//point := user.Point
	auth := user.Auth

	timestamp := getTimestamp()
	paramsMap := map[string]interface{}{
		"page":    0,
		"size":    20,
		"nocache": timestamp,
	}
	paramsJson, err := json.Marshal(paramsMap)
	if err != nil {
		return nil, err
	}

	sign, err := Sign(timestamp, upstream.BuddiesPath, string(paramsJson))
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("page", "0")
	params.Add("size", "20")
	params.Add("nocache", timestamp)

	// 构造请求
	apiUrl := upstreamURL(upstream.BuddiesPath)
	reqUrl := fmt.Sprintf("%s?%s", apiUrl, params.Encode())

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic("defer Body.Close err:" + err.Error())
		}
	}(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	arrstr := gjson.Get(string(body), "data.content").String()
	var arr []map[string]interface{}
	err = json.Unmarshal([]byte(arrstr), &arr)
	if err != nil {
		return nil, err
	}
	var buddies []*model.Buddy

	for _, item := range arr {
		buddies = append(buddies, &model.Buddy{
			UserID:    user.ID,
			BuddyID:   uint64(item["id"].(float64)),
			BuddyName: item["name"].(string),
		})
	}
	return buddies, nil
}
