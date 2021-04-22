package postgres

import (
	"encoding/base64"
	"errors"
	"strings"
)

func parseToken(token string) (id string, secret string, err error) {
	rawBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", "", err
	}

	rawParts := strings.SplitN(string(rawBytes), ":", 2)
	if len(rawParts) != 2 {
		return "", "", errors.New("invalid token specified")
	}

	return rawParts[0], rawParts[1], nil
}

func generateToken(id, secret string) string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(
		[]string{id, secret},
		":",
	)))
}
