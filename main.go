package main

import (
	"fmt"
	"github.com/mix-go/xcli"
	"github.com/mix-go/xutil/xenv"
	"hammer-web-api/commands"
	_ "hammer-web-api/config/dotenv"
	_ "hammer-web-api/config/viper"
	_ "hammer-web-api/di"
)

func main() {
	xcli.SetName("app").
		SetVersion("0.0.0-alpha").
		SetDebug(xenv.Getenv("APP_DEBUG").Bool(false))

	fmt.Println(xcli.App().Debug)
	xcli.AddCommand(commands.Commands...).Run()
}
