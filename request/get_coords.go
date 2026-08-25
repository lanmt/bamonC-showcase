package request

import (
	"bamonC/model"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/tidwall/gjson"
)

// getCoords /*
/*
params:
	user: 某个用户
	jsonStr: 验证码响应
	username: 用户名
return:
	ocr检测出的坐标，string json格式
*/
func GetCoords(user *model.User, jsonStr string) (string, error) {
	apiUrl := os.Getenv("OCR_API")
	reqJson := map[string]string{
		"base64_str": gjson.Get(jsonStr, "repData.originalImageBase64").String(),
		"word_list":  gjson.Get(jsonStr, "repData.wordList").String(),
	}
	jsonByte, _ := json.Marshal(reqJson)
	reqBody := bytes.NewBuffer(jsonByte)
	req, err := http.NewRequest("POST", apiUrl, reqBody)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic("defer Body.Close err:" + err.Error())
		}
	}(resp.Body)
	return string(body), err
}
