package config

// Config contains the configuration for the application.
type Config struct {
	RedisConfig RedisConfig `json:"redis_config"`
	Port        int         `json:"port"`
}

// RedisConfig contains the configuration for Redis.
type RedisConfig struct {
	Addr     string `json:"addr"`
	DB       int    `json:"db"`
	Password string `json:"password"`
}

// GetConfig returns the configuration for the application.
func GetConfig() *Config {
	cfg := &Config{
		RedisConfig: RedisConfig{
			Addr: "localhost:6379", // Default Redis address
		},
		Port: 8080,
	}
	return cfg
}
