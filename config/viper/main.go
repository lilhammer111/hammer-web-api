package viper

import (
	"fmt"
	"github.com/mix-go/xcli/argv"
	"github.com/spf13/viper"
	"hammer-web-api/config"
)

func init() {
	// Conf support JSON, TOML, YAML, HCL, INI, envfile
	viper.SetConfigFile(fmt.Sprintf("%s/../conf/config.yml", argv.Program().Dir))
	//viper.SetConfigFile("conf/config.yml")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(&config.Config); err != nil {
		panic(err)
	}
}
