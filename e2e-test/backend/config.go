//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"strconv"
)

type E2EConfig struct {
	MySQL      MySQLConfig
	PostgreSQL PostgreSQLConfig
}

type MySQLConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type PostgreSQLConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SSLMode  string
}

func LoadConfig() (*E2EConfig, error) {
	return &E2EConfig{
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "127.0.0.1"),
			Port:     getEnvInt("MYSQL_PORT", 13306),
			Username: getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", "test123456"),
			Database: getEnv("MYSQL_DATABASE", "mystisql_test"),
		},
		PostgreSQL: PostgreSQLConfig{
			Host:     getEnv("POSTGRES_HOST", "127.0.0.1"),
			Port:     getEnvInt("POSTGRES_PORT", 15432),
			Username: getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "test123456"),
			Database: getEnv("POSTGRES_DATABASE", "mystisql_test"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

func (c *PostgreSQLConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}
