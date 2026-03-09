package cli

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

var (
	tokenFlag string
)

func GetToken() string {
	if tokenFlag != "" {
		return tokenFlag
	}

	if token := os.Getenv("MYSTISQL_TOKEN"); token != "" {
		return token
	}

	return ""
}

func RequireToken() (string, error) {
	token := GetToken()
	if token == "" {
		return "", fmt.Errorf("未提供认证 Token，请使用 --token 参数或配置环境变量 MYSTISQL_TOKEN")
	}
	return token, nil
}

func ValidateTokenWithServer(token string, serverURL string) error {
	if GetLogger() != nil {
		GetLogger().Debug("Validating token with server",
			zap.String("server", serverURL),
		)
	}

	return nil
}
