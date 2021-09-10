package Config

import (
	"fmt"
	"github.com/spf13/viper"
)

type redis struct {
	Addr     string
	Password string
	Port     int
}

type jwt struct {
	JWT_SECRET string
}

type ViperConfig struct {
	Redis redis `mapstructure:"redis"`
	Jwt   jwt `mapstructure:"jwt"`
}

var Config ViperConfig

func init() {
	Viper := viper.New()
	Viper.AddConfigPath(".")
	Viper.SetConfigName("config")
	Viper.SetConfigType("toml")
	err := Viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	_ = Viper.Unmarshal(&Config)
}
