package service

import (
	"bamonC/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	RDB *redis.Client
	Ctx context.Context
}

// -------------------------------
// username -> token列表 [{},{"e":"123","toekn":"321"},{}]
// -------------------------------

// 获取 username 对应的列表
func (r *RedisService) GetTokens(username string) ([]map[string]string, error) {
	val, err := model.RDB.Get(r.Ctx, fmt.Sprintf("tokenlist:%s", username)).Result()
	if errors.Is(err, redis.Nil) {
		return []map[string]string{}, nil
	} else if err != nil {
		return nil, err
	}

	var list []map[string]string
	if err := json.Unmarshal([]byte(val), &list); err != nil {
		return nil, err
	}

	return list, nil
}

// 追加一条记录到 username 对应的列表中
func (r *RedisService) AppendUserToken(username string, item map[string]string) error {
	list, err := r.GetTokens(username)
	if err != nil {
		return err
	}

	list = append(list, item)

	data, err := json.Marshal(list)
	if err != nil {
		return err
	}

	return model.RDB.Set(r.Ctx, fmt.Sprintf("tokenlist:%s", username), data, 0).Err()
}

func (r *RedisService) ClearUserToken(username string) error {
	// 设置为空列表 []
	emptyList, err := json.Marshal([]map[string]string{})
	if err != nil {
		return err
	}

	return model.RDB.Set(r.Ctx, fmt.Sprintf("tokenlist:%s", username), emptyList, 0).Err()
}

// -------------------------------
// userid -> orderId
// -------------------------------

// 设置用户订单
func (r *RedisService) SetUserOrder(userid string, orderId string) error {
	return model.RDB.Set(r.Ctx, fmt.Sprintf("userorder:%s", userid), orderId, 0).Err()
}

// 获取用户订单
func (r *RedisService) GetUserOrder(userid string) (string, error) {
	val, err := model.RDB.Get(r.Ctx, fmt.Sprintf("userorder:%s", userid)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// -------------------------------
// 测试任务频率限制：每用户每分钟只能触发一次
// -------------------------------

// CheckAndSetTestRateLimit 如果用户在过去1分钟内已触发过测试，返回 false；否则设置标记并返回 true
func (r *RedisService) CheckAndSetTestRateLimit(username string) (bool, error) {
	key := fmt.Sprintf("test_ratelimit:%s", username)
	// SetNX：仅当key不存在时设置，TTL = 60s
	set, err := model.RDB.SetNX(r.Ctx, key, "1", 60*time.Second).Result()
	return set, err
}
