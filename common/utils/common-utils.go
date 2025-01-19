package utils

import "encoding/base64"

func EncodeBase64(src string) string {
	return base64.StdEncoding.EncodeToString([]byte(src))
}
func DecodeBase64(src string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", err
	}
	return string(b), err
}
