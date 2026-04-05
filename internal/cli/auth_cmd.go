package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"MystiSql/internal/service/auth"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "认证管理命令",
	Long:  `管理 MystiSql 认证 Token`,
}

var authTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "生成新的认证 Token",
	Long: `生成新的 JWT Token 用于 API 认证。

需要指定用户ID和角色，生成的 Token 将在指定的过期时间后失效。`,
	Example: `  # 生成管理员 Token
  mystisql auth token --user-id admin --role admin

  # 生成只读用户 Token
  mystisql auth token --user-id readonly --role readonly`,
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		role, _ := cmd.Flags().GetString("role")
		serverURL, _ := cmd.Flags().GetString("server")
		adminToken, _ := cmd.Flags().GetString("admin-token")

		if userID == "" || role == "" {
			return fmt.Errorf("必须指定 --user-id 和 --role")
		}

		if serverURL == "" {
			serverURL = "http://localhost:8080"
		}

		if adminToken == "" {
			adminToken = GetToken()
		}

		if adminToken == "" {
			return fmt.Errorf("必须提供管理员 Token，请使用 --admin-token 参数或配置环境变量 MYSTISQL_TOKEN")
		}

		client := &http.Client{Timeout: 30 * time.Second}

		body, err := json.Marshal(map[string]string{
			"user_id": userID,
			"role":    role,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}

		req, err := http.NewRequest("POST",
			serverURL+"/api/v1/auth/token",
			bytes.NewReader(body),
		)
		if err != nil {
			return fmt.Errorf("创建请求失败: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("请求服务器失败: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取响应失败: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("生成 Token 失败: %s", string(respBody))
		}

		fmt.Println(string(respBody))
		return nil
	},
}

var authRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "撤销认证 Token",
	Long: `撤销指定的 JWT Token，使其立即失效。

需要提供要撤销的 Token。`,
	Example: `  # 撤销 Token
  mystisql auth revoke --token <jwt_token>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _ := cmd.Flags().GetString("token")
		serverURL, _ := cmd.Flags().GetString("server")

		if token == "" {
			return fmt.Errorf("必须指定 --token")
		}

		if serverURL == "" {
			serverURL = "http://localhost:8080"
		}

		client := &http.Client{Timeout: 30 * time.Second}
		req, err := http.NewRequest("DELETE", serverURL+"/api/v1/auth/token", nil)
		if err != nil {
			return fmt.Errorf("创建请求失败: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("请求服务器失败: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取响应失败: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("撤销 Token 失败: %s", string(body))
		}

		fmt.Println("Token 已成功撤销")
		return nil
	},
}

var authInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "查看 Token 信息",
	Long:  `查看指定 Token 的详细信息，包括用户ID、角色、过期时间等。`,
	Example: `  # 查看 Token 信息
  mystisql auth info --token <jwt_token>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _ := cmd.Flags().GetString("token")
		serverURL, _ := cmd.Flags().GetString("server")

		if token == "" {
			return fmt.Errorf("必须指定 --token")
		}

		if serverURL == "" {
			serverURL = "http://localhost:8080"
		}

		client := &http.Client{Timeout: 30 * time.Second}
		req, err := http.NewRequest("GET", serverURL+"/api/v1/auth/token/info", nil)
		if err != nil {
			return fmt.Errorf("创建请求失败: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("请求服务器失败: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取响应失败: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("获取 Token 信息失败: %s", string(body))
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}

		fmt.Println("Token 信息:")
		fmt.Printf("  用户ID: %v\n", result["userId"])
		fmt.Printf("  角色: %v\n", result["role"])
		fmt.Printf("  Token ID: %v\n", result["tokenId"])
		fmt.Printf("  签发时间: %v\n", result["issuedAt"])
		fmt.Printf("  过期时间: %v\n", result["expiresAt"])

		return nil
	},
}

var authBootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "生成初始管理员 Token（首次部署使用）",
	Long: `直接使用配置文件中的 JWT Secret 生成管理员 Token，无需连接服务器。

此命令仅用于首次部署时生成初始管理员 Token。
生成后可使用该 Token 调用 API 生成其他用户 Token。`,
	Example: `  # 使用配置文件中的 secret 生成管理员 Token
  mystisql auth bootstrap

  # 指定配置文件
  mystisql auth bootstrap --config /path/to/config.yaml

  # 指定 secret 和过期时间（覆盖配置文件）
  mystisql auth bootstrap --secret my-secret-key --expire 48h`,
	RunE: func(cmd *cobra.Command, args []string) error {
		secret, _ := cmd.Flags().GetString("secret")
		expireStr, _ := cmd.Flags().GetString("expire")

		if secret == "" {
			if cfg == nil || cfg.Auth.Token.Secret == "" {
				return fmt.Errorf("必须通过 --secret 参数或配置文件 auth.token.secret 提供 JWT Secret")
			}
			secret = cfg.Auth.Token.Secret
		}

		if expireStr == "" {
			if cfg != nil && cfg.Auth.Token.Expire != "" {
				expireStr = cfg.Auth.Token.Expire
			} else {
				expireStr = "24h"
			}
		}

		expire, err := time.ParseDuration(expireStr)
		if err != nil {
			return fmt.Errorf("无效的过期时间格式 %q: %w", expireStr, err)
		}

		generator, err := auth.NewTokenGenerator(secret, expire)
		if err != nil {
			return fmt.Errorf("创建 Token 生成器失败: %w", err)
		}

		token, err := generator.GenerateToken("admin", "admin")
		if err != nil {
			return fmt.Errorf("生成 Token 失败: %w", err)
		}

		fmt.Println(token)
		return nil
	},
}

func init() {
	authTokenCmd.Flags().String("user-id", "", "用户ID")
	authTokenCmd.Flags().String("role", "", "用户角色")
	authTokenCmd.Flags().String("server", "http://localhost:8080", "服务器地址")
	authTokenCmd.Flags().String("admin-token", "", "管理员 Token（用于认证）")

	authRevokeCmd.Flags().String("token", "", "要撤销的 Token")
	authRevokeCmd.Flags().String("server", "http://localhost:8080", "服务器地址")

	authInfoCmd.Flags().String("token", "", "要查看的 Token")
	authInfoCmd.Flags().String("server", "http://localhost:8080", "服务器地址")

	authBootstrapCmd.Flags().String("secret", "", "JWT Secret（覆盖配置文件）")
	authBootstrapCmd.Flags().String("expire", "", "Token 过期时间（覆盖配置文件，默认 24h）")

	authCmd.AddCommand(authTokenCmd)
	authCmd.AddCommand(authRevokeCmd)
	authCmd.AddCommand(authInfoCmd)
	authCmd.AddCommand(authBootstrapCmd)

	rootCmd.AddCommand(authCmd)
}
