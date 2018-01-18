package main

import (
	"Taonet/work"
	"Taonet/conf"
	"github.com/fatih/color"
)

func main(){

	const (
		banner = `
 ______)
   (, /
     /   _   _____    _ _/_
  ) /   (_(_(_) / (__(/_(__
 (_/

         `
    )

    color.Cyan(banner)

    color.Blue("[Taonet version]	")

    color.Green("%s", conf.TaoConf["version"])

	color.Blue("[Listen tcp port]	")

	color.Green("%s", conf.TaoConf["socket_tcp_port"])

	color.Blue("[Connect redis address]	")

	color.Green("%s", conf.TaoConf["redis_addr"])

	go work.StartTaoNet()

	select {}
}




