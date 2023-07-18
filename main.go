package main

import (
	"easyssh/authorized"
	"easyssh/config"
)

func main() {
	config := config.Init()

	for _, host := range config.Hosts {
		authorized.Init(config, &host)
	}
}
