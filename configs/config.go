package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	ServerPort    string

	// Default rate limits
	RateLimitIP             int
	RateLimitIPBlockTime    int
	RateLimitToken          int
	RateLimitTokenBlockTime int

	// Token-specific configurations
	TokenConfigs map[string]TokenConfig
}

type TokenConfig struct {
	Limit     int
	BlockTime int
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),
		ServerPort:    getEnv("SERVER_PORT", "8080"),

		RateLimitIP:             getEnvAsInt("RATE_LIMIT_IP", 10),
		RateLimitIPBlockTime:    getEnvAsInt("RATE_LIMIT_IP_BLOCK_TIME", 300),
		RateLimitToken:          getEnvAsInt("RATE_LIMIT_TOKEN", 100),
		RateLimitTokenBlockTime: getEnvAsInt("RATE_LIMIT_TOKEN_BLOCK_TIME", 300),

		TokenConfigs: make(map[string]TokenConfig),
	}

	// Load token-specific configurations
	cfg.loadTokenConfigs()

	return cfg, nil
}

func (c *Config) loadTokenConfigs() {
	tokens := make(map[string]string)

	// First pass: find all TOKEN_* entries
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}

		key := pair[0]
		value := pair[1]

		// Check if it's a TOKEN_{NAME} entry (not _LIMIT or _BLOCK_TIME)
		if strings.HasPrefix(key, "TOKEN_") &&
			!strings.HasSuffix(key, "_LIMIT") &&
			!strings.HasSuffix(key, "_BLOCK_TIME") {
			// Extract the token name (e.g., "ONE" from "TOKEN_ONE")
			tokenName := strings.TrimPrefix(key, "TOKEN_")
			tokens[tokenName] = value
		}
	}

	// Second pass: for each token found, get its limit and block time
	for tokenName, tokenValue := range tokens {
		limitKey := fmt.Sprintf("TOKEN_%s_LIMIT", tokenName)
		blockTimeKey := fmt.Sprintf("TOKEN_%s_BLOCK_TIME", tokenName)

		limit := getEnvAsInt(limitKey, c.RateLimitToken)
		blockTime := getEnvAsInt(blockTimeKey, c.RateLimitTokenBlockTime)

		c.TokenConfigs[tokenValue] = TokenConfig{
			Limit:     limit,
			BlockTime: blockTime,
		}
	}
}

func (c *Config) GetTokenConfig(token string) (TokenConfig, bool) {
	cfg, exists := c.TokenConfigs[token]
	if !exists {
		// Return default token configuration
		return TokenConfig{
			Limit:     c.RateLimitToken,
			BlockTime: c.RateLimitTokenBlockTime,
		}, false
	}
	return cfg, true
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
