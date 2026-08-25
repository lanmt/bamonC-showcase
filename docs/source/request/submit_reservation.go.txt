package request

import (
	"bamonC/model"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/tidwall/gjson"
)

// SubmitReservation 提交预约订单
func SubmitReservation(user *model.User, court *model.Court, e, captchaToken string) (string, error) {
	//reservationOrderJson
	/*
		[{"timeId":8791,"spaceId":130},{"timeId":8792,"spaceId":130}]
	*/
	reservationOrderJson, _ := BuildReservationJson(court.CourtID, court.Time1ID, court.Time2ID)
	venueSiteId := strconv.FormatUint(court.VenueSiteID, 10)
	timestamp := getTimestamp() // 毫秒时间戳
	data := url.Values{}
	data.Set("captchaVerification", e)
	data.Set("captchaToken", captchaToken)
	data.Set("reservationOrderJson", reservationOrderJson)
	data.Set("reservationDate", getDate())
	data.Set("weekStartDate", getDate())
	data.Set("reservationType", "-1")
	data.Set("orderPrice", "25")
	data.Set("orderPin", os.Getenv("UPSTREAM_ORDER_PIN"))
	data.Set("venueSiteId", venueSiteId)
	data.Set("phone", os.Getenv("RESERVATION_CONTACT_PHONE"))
	data.Set("manNum", "0")
	data.Set("childNum", "0")
	data.Set("buddyUids", "")
	data.Set("buddyIds", strconv.FormatUint(*user.SelectBuddyID, 10))
	formBody := data.Encode() // urlencoded 格式

	jsonData := make(map[string]string)
	for k, v := range data {
		if len(v) > 0 {
			jsonData[k] = v[0]
		}
	}
	jsonStr, _ := json.Marshal(jsonData)
	sign, err := Sign(timestamp, upstream.ReservationPath, string(jsonStr))
	if err != nil {
		return "", err
	}

	// 构建请求
	req, err := http.NewRequest("POST", upstreamURL(upstream.ReservationPath), bytes.NewBufferString(formBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("app-key", upstream.AppKey)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("sign", sign)
	req.Header.Set(upstream.AuthHeader, user.Auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("accept", "application/json, text/plain, */*")

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

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(bodyBytes))
	if err != nil {
		return "", err
	}
	tradeNo := gjson.GetBytes(bodyBytes, "data.tradeNo").String()

	if tradeNo == "" {
		message := user.Username + gjson.GetBytes(bodyBytes, "message").String()
		return "", errors.New(message)
	}

	return tradeNo, nil
}

func BuildReservationJson(spaceId uint64, timeId1 *uint64, timeId2 *uint64) (string, error) {
	type Item struct {
		TimeID  uint64 `json:"timeId"`
		SpaceID uint64 `json:"spaceId"`
	}

	list := []Item{
		{
			TimeID:  *timeId1,
			SpaceID: spaceId,
		},
	}

	// timeId2 不为 nil 才追加第二个元素
	if timeId2 != nil {
		list = append(list, Item{
			TimeID:  *timeId2,
			SpaceID: spaceId,
		})
	}

	b, err := json.Marshal(list)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
