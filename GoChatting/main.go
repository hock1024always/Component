package main

import (
	"GoChatting/conf"
	"GoChatting/router"
	"GoChatting/service"
)

func main() {
	conf.Init()
	go service.Manager.Start()
	r := router.NewRouter()
	_ = r.Run(conf.HttpPort)
}
