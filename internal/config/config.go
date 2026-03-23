package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	BotToken    string `env:"BOT_TOKEN" env-required:"true"`
	MongoURI    string `env:"MONGO_URI" env-default:"mongodb://localhost:27017"`
	MongoDB     string `env:"MONGO_DB" env-default:"go_motivation"`
	AdminChatID int64  `env:"ADMIN_CHAT_ID" env-required:"true"`

	SMTPHost string `env:"SMTP_HOST" env-default:"smtp.gmail.com"`
	SMTPPort int    `env:"SMTP_PORT" env-default:"587"`
	SMTPUser string `env:"SMTP_USER"`
	SMTPPass string `env:"SMTP_PASS"`
	SMTPFrom string `env:"SMTP_FROM"`
}

func (c *Config) SMTPEnabled() bool {
	return c.SMTPUser != "" && c.SMTPPass != "" && c.SMTPFrom != ""
}

func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}
