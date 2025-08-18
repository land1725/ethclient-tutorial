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
	InfuraProjectID      string
	EthereumNetwork      string
	EthereumWSURL        string
	EthereumHTTPURL      string
	TestPrivateKey       string
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
		InfuraProjectID:      getEnv("INFURA_PROJECT_ID", ""),
		EthereumNetwork:      getEnv("ETHEREUM_NETWORK", "mainnet"),
		EthereumWSURL:        getEnv("ETHEREUM_WS_URL", "wss://mainnet.infura.io/ws/v3/"),
		EthereumHTTPURL:      getEnv("ETHEREUM_HTTP_URL", "https://mainnet.infura.io/v3/"),
		TestPrivateKey:       getEnv("TEST_PRIVATE_KEY", ""),
		TestRecipientAddress: getEnv("TEST_RECIPIENT_ADDRESS", "0x742d35Cc6634C0532925a3b8D4C9db96C5C7F4C1"),
		ContractAddress:      getEnv("CONTRACT_ADDRESS", ""),
		ContractABIPath:      getEnv("CONTRACT_ABI_PATH", "./contracts/abi/"),
		DefaultGasLimit:      getEnvAsUint64("DEFAULT_GAS_LIMIT", 21000),
		GasPriceMultiplier:   getEnvAsFloat64("GAS_PRICE_MULTIPLIER", 1.1),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		LogOutput:            getEnv("LOG_OUTPUT", "console"),
	}

	GlobalConfig = config
	return config
}

// GetEthereumURL 获取完整的以太坊连接URL
func (c *Config) GetEthereumURL() string {
	if c.InfuraProjectID == "" {
		log.Fatal("INFURA_PROJECT_ID is required in .env file")
	}
	return c.EthereumWSURL + c.InfuraProjectID
}

// GetHTTPURL 获取HTTP连接URL
func (c *Config) GetHTTPURL() string {
	if c.InfuraProjectID == "" {
		log.Fatal("INFURA_PROJECT_ID is required in .env file")
	}
	return c.EthereumHTTPURL + c.InfuraProjectID
}

// ValidateConfig 验证配置
func (c *Config) ValidateConfig() error {
	if c.InfuraProjectID == "" {
		log.Println("Warning: INFURA_PROJECT_ID not set - network functions will not work")
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
