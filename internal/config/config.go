package config

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	SecretKey string
	ExpiresIn int64
}

func LoadConfig() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "19770811",
			DBName:   "avito_shop",
			SSLMode:  "disable",
		},
		JWT: JWTConfig{
			SecretKey: "your-secret-key",
			ExpiresIn: 24,
		},
	}, nil
}
