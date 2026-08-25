package request

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Production endpoints and credentials belong to an internal campus system.
// They are intentionally omitted from this public portfolio version.
var upstream = struct {
	BaseURL           string
	AppKey            string
	AuthHeader        string
	BuddiesPath       string
	CaptchaGetPath    string
	CaptchaCheckPath  string
	ReservationPath  string
	CancellationPath string
}{
	BaseURL:           envOrDefault("UPSTREAM_API_BASE_URL", "https://api.example.invalid"),
	AppKey:            envOrDefault("UPSTREAM_APP_KEY", "portfolio-demo-key"),
	AuthHeader:        envOrDefault("UPSTREAM_AUTH_HEADER", "X-API-Authorization"),
	BuddiesPath:       envOrDefault("UPSTREAM_BUDDIES_PATH", "/redacted/buddies"),
	CaptchaGetPath:    envOrDefault("UPSTREAM_CAPTCHA_GET_PATH", "/redacted/captcha/challenge"),
	CaptchaCheckPath:  envOrDefault("UPSTREAM_CAPTCHA_CHECK_PATH", "/redacted/captcha/verify"),
	ReservationPath:  envOrDefault("UPSTREAM_RESERVATION_PATH", "/redacted/reservations"),
	CancellationPath: envOrDefault("UPSTREAM_CANCELLATION_PATH", "/redacted/reservations/cancel"),
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func upstreamURL(path string) string {
	return strings.TrimRight(upstream.BaseURL, "/") + "/" + strings.TrimLeft(path, "/")
}

func getDate() string {
	now := time.Now()

	var target time.Time
	hour := now.Hour()

	if hour >= 0 && hour < 5 {
		// 凌晨 0~5 点，+1 天 ，这里是为了测试
		target = now.AddDate(0, 0, 1)
	} else {
		// 其他时间 +2 天
		target = now.AddDate(0, 0, 2)
	}

	return target.Format("2006-01-02")
}

func getTimestamp() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}
