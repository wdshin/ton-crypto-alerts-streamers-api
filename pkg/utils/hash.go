package utils

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"
)

func СomputeSign(walletAddress string, amount string, nickname string, message string, lt string) string {
	stringToHash := strings.Join([]string{
		walletAddress,
		amount,
		nickname,
		message,
		lt,
	}, "")

	bytes := md5.Sum([]byte(stringToHash))
	return fmt.Sprintf("%x", bytes)
}

func СomputeSign2(walletAddress string, amount string, nickname string, message string, lt string) string {
	stringToHash := strings.Join([]string{
		walletAddress,
		amount,
		nickname,
		message,
		lt,
	}, "")

	sEnc := base64.StdEncoding.EncodeToString([]byte(stringToHash))
	return sEnc
}
