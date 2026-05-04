package config

type Config struct {
	DB *DB
}

func New() (*Config, error) {
	cfg := &Config{
		DB: &DB{
			Host:     "localhost",
			Port:     5432,
			User:     "user",
			Password: "secret_password",
			DB:       "db",
		},
	}

	return cfg, nil
}
