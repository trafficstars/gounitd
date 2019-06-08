package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/trafficstars/gounit/server"
)

func main() {
	enableLogging := flag.Bool("enable-logging", false, `should we print about everything?`)
	configPath := flag.String("config", "/etc/gounit.yaml", `path to the config file [default: "/etc/gounit.yaml"]`)
	flag.Parse()

	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	cfg := server.NewConfig()
	err = cfg.Parse(configFile)
	if err != nil {
		log.Fatal(err)
	}
	if *enableLogging {
		fmt.Println("config:", *cfg)
	}
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if *enableLogging {
		srv.AccessLogger = newLogger(`access`)
		srv.ErrorLogger = newLogger(`error`)
	}
	err = srv.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(srv.Wait())
}
