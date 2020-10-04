package config

import (
	"github.com/spf13/viper"
)

// App is the configuration structure for the application exclude the database
type App struct {
	JWTSecret string `yaml:"jwtSecret"`
}

// DB is the configuration structure for the database
type DB struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Mailer is the configuration structure for the database
type Mailer struct {
	Host  string `yaml:"host"`
	Port  string `yaml:"port"`
	Email string `yaml:"email"`
	PWD   string `yaml:"pwd"`
}

type configuration struct {
	App    App    `yaml:"app"`
	DB     DB     `yaml:"db"`
	Mailer Mailer `yaml:"mailer"`
}

// Load parse .env file to load configuration struct
func Load() (configuration, error) {
	c := &configuration{}

	viper.SetConfigName(".env")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		return *c, err
	}

	err = viper.Unmarshal(c)
	if err != nil {
		return *c, err
	}

	return *c, nil
}
