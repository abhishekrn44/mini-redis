package main

import (
	"flag"
	"log"
	"mini-redis/config"
	"mini-redis/server"
)

func configValues() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Mini-Redis Host")
	flag.IntVar(&config.Port, "port", 6379, "Mini-Redis Port")
	flag.Parse()
}

func main() {
	configValues()
	log.Println("Starting mini-redis")
	server.StartAsyncTCPServer()
}
