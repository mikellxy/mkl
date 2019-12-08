package main

import (
	"fmt"
	"github.com/mikellxy/mkl/api/deploy"
	"github.com/mikellxy/mkl/config"
)

func errorHandler(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	//image := "mikellxy/test-ci:ef95f0d98da9d5f76cb75c49e3541f8c98a9327f"

	config.LoadConfig()

	s := deploy.GetServer()
	conf := config.Conf.PanelService
	s.Run(fmt.Sprintf("%s:%s", conf.Host, conf.Port))
}
