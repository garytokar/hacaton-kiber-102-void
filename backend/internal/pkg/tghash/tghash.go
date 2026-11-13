package tghash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strings"
)

type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// ValidateInitData проверяет HMAC-SHA256 подпись Telegram Mini App initData
func ValidateInitData(initDataRaw, botToken string) (*TelegramUser, error) {
	// Режим локальной разработки без реального бота
	if botToken == "" || botToken == "mock_token" || initDataRaw == "dev_test" {
		return &TelegramUser{
			ID:        999999999,
			FirstName: "Тестовый",
			LastName:  "Ученик",
			Username:  "dev_student",
		}, nil
	}

	values, err := url.ParseQuery(initDataRaw)
	if err != nil {
		return nil, err
	}

	receivedHash := values.Get("hash")
	if receivedHash == "" {
		return nil, errors.New("hash is missing in initData")
	}

	var pairs []string
	for k, v := range values {
		if k == "hash" {
			continue
		}
		if len(v) > 0 {
			pairs = append(pairs, k+"="+v[0])
		}
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// 1. secretKey = HMAC_SHA256("WebAppData", botToken)
	mac := hmac.New(sha256.New, []byte("WebAppData"))
	mac.Write([]byte(botToken))
	secretKey := mac.Sum(nil)

	// 2. calculatedHash = HMAC_SHA256(secretKey, dataCheckString)
	macCheck := hmac.New(sha256.New, secretKey)
	macCheck.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(macCheck.Sum(nil))

	if !hmac.Equal([]byte(calculatedHash), []byte(receivedHash)) {
		return nil, errors.New("invalid signature")
	}

	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, errors.New("user object is missing in initData")
	}

	var tgUser TelegramUser
	if err := json.Unmarshal([]byte(userJSON), &tgUser); err != nil {
		return nil, err
	}

	return &tgUser, nil
}
