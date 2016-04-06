package main

import (
	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/server"
	"github.com/clawio/service-localfs-data/service"
)

func main() {
	var cfg *service.Config
	config.LoadJSONFile("./config.json", &cfg)

	server.Init("service-localfs-data", cfg.Server)

	svc, err := service.New(cfg)
	if err != nil {
		server.Log.Fatal("unable to create service: ", err)
	}
	err = server.Register(svc)
	if err != nil {
		server.Log.Fatal("unable to register service: ", err)
	}

	err = server.Run()
	if err != nil {
		server.Log.Fatal("server encountered a fatal error: ", err)
	}
}
