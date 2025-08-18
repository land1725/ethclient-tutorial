package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config 配置结构体
type Config struct {
	AlchemyAPIKey        string
	EthereumNetwork      string
	EthereumWSURL        string
	EthereumHTTPURL      string
	TestPrivateKey       string
	TestSendAddress      string
	TestRecipientAddress string
	ContractAddress      string
	ContractABIPath      string
	DefaultGasLimit      uint64
	GasPriceMultiplier   float64
	LogLevel             string
	LogOutput            string
}

var GlobalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig() *Config {
	// 加载.env文件
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	config := &Config{
		AlchemyAPIKey:        getEnv("ALCHEMY_API_KEY", ""),
		EthereumNetwork:      getEnv("ETHEREUM_NETWORK", ""),
		EthereumWSURL:        getEnv("ETHEREUM_WS_URL", ""),
		EthereumHTTPURL:      getEnv("ETHEREUM_HTTP_URL", ""),
		TestPrivateKey:       getEnv("TEST_PRIVATE_KEY", ""),
		TestSendAddress:      getEnv("TEST_SEND_ADDRESS", ""),
		TestRecipientAddress: getEnv("TEST_RECIPIENT_ADDRESS", ""),
		ContractAddress:      getEnv("CONTRACT_ADDRESS", ""),
		ContractABIPath:      getEnv("CONTRACT_ABI_PATH", "./contracts/abi/"),
		DefaultGasLimit:      getEnvAsUint64("DEFAULT_GAS_LIMIT", 0),
		GasPriceMultiplier:   getEnvAsFloat64("GAS_PRICE_MULTIPLIER", 1.1),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		LogOutput:            getEnv("LOG_OUTPUT", "console"),
	}

	GlobalConfig = config
	return config
}

// GetEthereumURL 获取完整的以太坊连接URL（优先使用HTTP连接）
func (c *Config) GetEthereumURL() string {
	if c.AlchemyAPIKey == "" {
		log.Fatal("ALCHEMY_API_KEY is required in .env file")
	}
	// 优先使用HTTP连接，因为它更稳定
	return c.EthereumHTTPURL + c.AlchemyAPIKey
}

// GetWebSocketURL 获取WebSocket连接URL（用于需要实时订阅的功能）
func (c *Config) GetWebSocketURL() string {
	if c.AlchemyAPIKey == "" {
		log.Fatal("ALCHEMY_API_KEY is required in .env file")
	}
	return c.EthereumWSURL + c.AlchemyAPIKey
}

// GetHTTPURL 获取HTTP连接URL
func (c *Config) GetHTTPURL() string {
	if c.AlchemyAPIKey == "" {
		log.Fatal("ALCHEMY_API_KEY is required in .env file")
	}
	return c.EthereumHTTPURL + c.AlchemyAPIKey
}

// ValidateConfig 验证配置
func (c *Config) ValidateConfig() error {
	if c.AlchemyAPIKey == "" {
		log.Println("Warning: ALCHEMY_API_KEY not set - network functions will not work")
	}
	if c.TestPrivateKey == "" {
		log.Println("Warning: TEST_PRIVATE_KEY not set - transfer functions will not work")
	}
	return nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsUint64 获取环境变量并转换为uint64
func getEnvAsUint64(key string, defaultValue uint64) uint64 {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		log.Printf("Warning: Invalid value for %s, using default %d", key, defaultValue)
		return defaultValue
	}
	return value
}

// getEnvAsFloat64 获取环境变量并转换为float64
func getEnvAsFloat64(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Printf("Warning: Invalid value for %s, using default %f", key, defaultValue)
		return defaultValue
	}
	return value
}

// IsProductionMode 检查是否为生产模式
func (c *Config) IsProductionMode() bool {
	return strings.ToLower(c.EthereumNetwork) == "mainnet"
}

// IsTestMode 检查是否为测试模式
func (c *Config) IsTestMode() bool {
	network := strings.ToLower(c.EthereumNetwork)
	return network == "sepolia" || network == "goerli" || network == "localhost"
}
