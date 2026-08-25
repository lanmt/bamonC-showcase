package request

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

func Sign(timestamp, path string, dataJSON string) (string, error) {
	w := os.Getenv("UPSTREAM_SIGNING_SECRET")
	if w == "" {
		return "", fmt.Errorf("UPSTREAM_SIGNING_SECRET is required")
	}
	var data map[string]interface{}
	err := json.Unmarshal([]byte(dataJSON), &data)
	if err != nil {
		return "", fmt.Errorf("invalid json: %v", err)
	}

	c := w + path

	// 获取所有 key 并排序
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := data[k]

		// 过滤掉 object、空字符串、nil
		switch val := v.(type) {
		case string:
			if val != "" {
				c += k + val
			}
		case float64, bool:
			c += fmt.Sprintf("%v%v", k, val)
		// 忽略嵌套 object、array、nil
		default:
			continue
		}
	}

	c += fmt.Sprintf("%v %v", timestamp, w)

	hash := md5.Sum([]byte(c))
	return hex.EncodeToString(hash[:]), nil
}

// PKCS7 填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	// 计算需要填充的字节数
	padLen := blockSize - len(data)%blockSize
	// 创建填充的字节
	pad := bytes.Repeat([]byte{byte(padLen)}, padLen)
	// 返回填充后的数据
	return append(data, pad...)
}

// AES-ECB 加密
func aesECBEncrypt(data, key []byte) ([]byte, error) {
	// AES 密钥必须是 16, 24, 32 字节
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("key length must be 16, 24, or 32 bytes")
	}

	// 创建 AES 密码块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 获取块的大小
	blockSize := block.BlockSize()
	// 填充数据使其块大小对齐
	data = pkcs7Pad(data, blockSize)

	// 创建加密后的字节切片
	encrypted := make([]byte, len(data))
	dst := encrypted

	// 按块加密数据
	for len(data) > 0 {
		block.Encrypt(dst, data[:blockSize])
		data = data[blockSize:]
		dst = dst[blockSize:]
	}

	return encrypted, nil
}

// AES 加密（和 JS 的 d 函数一致）
func D(e string, key string) (string, error) {
	if key == "" {
		key = os.Getenv("UPSTREAM_ENCRYPTION_KEY")
	}
	if key == "" {
		return "", fmt.Errorf("UPSTREAM_ENCRYPTION_KEY is required")
	}

	// 转换为字节
	encrypted, err := aesECBEncrypt([]byte(e), []byte(key))
	if err != nil {
		return "", err
	}

	// 返回 Base64 编码的加密结果
	return base64.StdEncoding.EncodeToString(encrypted), nil
}
