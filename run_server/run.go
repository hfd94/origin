package main

import (
	"github.com/duanhf2012/origin/v2/config"
	"github.com/duanhf2012/origin/v2/node"
)

func main() {
	config.SetClusterPath("D:\\DevelopProject\\golang\\lg-game\\bin\\cluster")
	config.NewConfig("Mysql", "D:\\DevelopProject\\golang\\lg-game\\bin\\config\\config.json")

	node.Start()
}
