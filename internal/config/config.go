package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App        AppConfig
	DB         DBConfig
	Valkey     ValkeyConfig
	JWT        JWTConfig
	Auth       AuthConfig
	R2         R2Config
	Blockchain BlockchainConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type ValkeyConfig struct {
	Addr     string
	Password string
	DB       int
	TLS      bool
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

type AuthConfig struct {
	NonceTTLMinutes int
	EIP712Name      string
	EIP712Version   string
	EIP712ChainID   int64
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Endpoint        string
	Region          string
}

type BlockchainConfig struct {
	MantleRPCURL    string
	OwnaFarmNFTAddr string
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return fallback
}

func LoadConfig() *Config {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("warning: .env not loaded (will use OS environment variables):", err)
	}

	valkeyDB, err := strconv.Atoi(getEnv("VALKEY_DB", "0"))
	if err != nil {
		log.Fatal("env: VALKEY_DB must be an integer")
	}

	valkeyTLS, err := strconv.ParseBool(getEnv("VALKEY_TLS", "false"))
	if err != nil {
		log.Fatal("env: VALKEY_TLS must be a boolean")
	}

	jwtExpirationHours, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		log.Fatal("env: JWT_EXPIRATION_HOURS must be an integer")
	}

	nonceTTLMinutes, err := strconv.Atoi(getEnv("NONCE_TTL_MINUTES", "5"))
	if err != nil {
		log.Fatal("env: NONCE_TTL_MINUTES must be an integer")
	}

	eip712ChainID, err := strconv.ParseInt(getEnv("EIP712_CHAIN_ID", "5000"), 10, 64)
	if err != nil {
		log.Fatal("env: EIP712_CHAIN_ID must be an integer")
	}

	return &Config{
		App: AppConfig{
			Port: getEnv("APP_PORT", "8080"),
			Env:  getEnv("APP_ENV", "development"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "postgres"),
		},
		Valkey: ValkeyConfig{
			Addr:     getEnv("VALKEY_ADDR", "localhost:6379"),
			Password: getEnv("VALKEY_PASSWORD", ""),
			DB:       valkeyDB,
			TLS:      valkeyTLS,
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", ""),
			ExpirationHours: jwtExpirationHours,
		},
		Auth: AuthConfig{
			NonceTTLMinutes: nonceTTLMinutes,
			EIP712Name:      getEnv("EIP712_NAME", "OwnaFarm"),
			EIP712Version:   getEnv("EIP712_VERSION", "1"),
			EIP712ChainID:   eip712ChainID,
		},
		R2: R2Config{
			AccountID:       getEnv("R2_ACCOUNT_ID", ""),
			AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
			Bucket:          getEnv("R2_BUCKET", ""),
			Endpoint:        getEnv("R2_ENDPOINT", ""),
			Region:          getEnv("R2_REGION", "auto"),
		},
		Blockchain: BlockchainConfig{
			MantleRPCURL:    getEnv("MANTLE_RPC_URL", "https://rpc.sepolia.mantle.xyz"),
			OwnaFarmNFTAddr: getEnv("OWNAFARM_NFT_ADDRESS", "0xC51601dde25775bA2740EE14D633FA54e12Ef6C7"),
		},
	}
}
