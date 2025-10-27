package configs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Set test environment variables
	os.Setenv("REDIS_HOST", "test-redis")
	os.Setenv("REDIS_PORT", "7000")
	os.Setenv("REDIS_DB", "5")
	os.Setenv("RATE_LIMIT_IP", "20")
	os.Setenv("RATE_LIMIT_IP_BLOCK_TIME", "600")
	os.Setenv("RATE_LIMIT_TOKEN", "200")
	os.Setenv("RATE_LIMIT_TOKEN_BLOCK_TIME", "900")
	os.Setenv("SERVER_PORT", "9090")

	defer func() {
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_DB")
		os.Unsetenv("RATE_LIMIT_IP")
		os.Unsetenv("RATE_LIMIT_IP_BLOCK_TIME")
		os.Unsetenv("RATE_LIMIT_TOKEN")
		os.Unsetenv("RATE_LIMIT_TOKEN_BLOCK_TIME")
		os.Unsetenv("SERVER_PORT")
	}()

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "test-redis", cfg.RedisHost)
	assert.Equal(t, "7000", cfg.RedisPort)
	assert.Equal(t, 5, cfg.RedisDB)
	assert.Equal(t, 20, cfg.RateLimitIP)
	assert.Equal(t, 600, cfg.RateLimitIPBlockTime)
	assert.Equal(t, 200, cfg.RateLimitToken)
	assert.Equal(t, 900, cfg.RateLimitTokenBlockTime)
	assert.Equal(t, "9090", cfg.ServerPort)
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear all environment variables
	os.Clearenv()

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check defaults
	assert.Equal(t, "localhost", cfg.RedisHost)
	assert.Equal(t, "6379", cfg.RedisPort)
	assert.Equal(t, 0, cfg.RedisDB)
	assert.Equal(t, 10, cfg.RateLimitIP)
	assert.Equal(t, 300, cfg.RateLimitIPBlockTime)
	assert.Equal(t, 100, cfg.RateLimitToken)
	assert.Equal(t, 300, cfg.RateLimitTokenBlockTime)
	assert.Equal(t, "8080", cfg.ServerPort)
}

func TestLoadConfig_TokenConfigs(t *testing.T) {
	os.Setenv("TOKEN_PREMIUM", "premium-token-123")
	os.Setenv("TOKEN_PREMIUM_LIMIT", "1000")
	os.Setenv("TOKEN_PREMIUM_BLOCK_TIME", "120")
	os.Setenv("TOKEN_BASIC", "basic-token-456")
	os.Setenv("TOKEN_BASIC_LIMIT", "50")
	os.Setenv("TOKEN_BASIC_BLOCK_TIME", "600")

	defer func() {
		os.Unsetenv("TOKEN_PREMIUM")
		os.Unsetenv("TOKEN_PREMIUM_LIMIT")
		os.Unsetenv("TOKEN_PREMIUM_BLOCK_TIME")
		os.Unsetenv("TOKEN_BASIC")
		os.Unsetenv("TOKEN_BASIC_LIMIT")
		os.Unsetenv("TOKEN_BASIC_BLOCK_TIME")
	}()

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check premium token config (using actual token value as key)
	premiumCfg, exists := cfg.TokenConfigs["premium-token-123"]
	assert.True(t, exists)
	assert.Equal(t, 1000, premiumCfg.Limit)
	assert.Equal(t, 120, premiumCfg.BlockTime)

	// Check basic token config (using actual token value as key)
	basicCfg, exists := cfg.TokenConfigs["basic-token-456"]
	assert.True(t, exists)
	assert.Equal(t, 50, basicCfg.Limit)
	assert.Equal(t, 600, basicCfg.BlockTime)
}

func TestGetTokenConfig(t *testing.T) {
	os.Setenv("TOKEN_TEST", "test-token-789")
	os.Setenv("TOKEN_TEST_LIMIT", "500")
	os.Setenv("TOKEN_TEST_BLOCK_TIME", "180")
	os.Setenv("RATE_LIMIT_TOKEN", "100")
	os.Setenv("RATE_LIMIT_TOKEN_BLOCK_TIME", "300")

	defer func() {
		os.Unsetenv("TOKEN_TEST")
		os.Unsetenv("TOKEN_TEST_LIMIT")
		os.Unsetenv("TOKEN_TEST_BLOCK_TIME")
		os.Unsetenv("RATE_LIMIT_TOKEN")
		os.Unsetenv("RATE_LIMIT_TOKEN_BLOCK_TIME")
	}()

	cfg, _ := LoadConfig()

	// Test existing token config (using actual token value)
	tokenCfg, exists := cfg.GetTokenConfig("test-token-789")
	assert.True(t, exists)
	assert.Equal(t, 500, tokenCfg.Limit)
	assert.Equal(t, 180, tokenCfg.BlockTime)

	// Test non-existing token config (should return defaults)
	defaultCfg, exists := cfg.GetTokenConfig("nonexistent-token")
	assert.False(t, exists)
	assert.Equal(t, 100, defaultCfg.Limit)
	assert.Equal(t, 300, defaultCfg.BlockTime)
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")

	value := getEnv("TEST_VAR", "default")
	assert.Equal(t, "test-value", value)

	value = getEnv("NON_EXISTENT", "default")
	assert.Equal(t, "default", value)
}

func TestGetEnvAsInt(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	value := getEnvAsInt("TEST_INT", 10)
	assert.Equal(t, 42, value)

	value = getEnvAsInt("NON_EXISTENT", 10)
	assert.Equal(t, 10, value)

	os.Setenv("INVALID_INT", "not-a-number")
	defer os.Unsetenv("INVALID_INT")

	value = getEnvAsInt("INVALID_INT", 10)
	assert.Equal(t, 10, value)
}
