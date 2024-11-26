/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/souravbiswassanto/concurrent-file-server/internal/client"
	"github.com/souravbiswassanto/concurrent-file-server/internal/server"
	"github.com/souravbiswassanto/concurrent-file-server/internal/util"
	"log"
	"time"
)

func main() {
	go func() {
		time.Sleep(time.Second * 2)
		fc := client.FileClient{}
		fc.DefaultSetup()
		fc.Start()
	}()

	err := server.SetupAndRunServer(util.HandleFunc{})
	log.Println(err)
	//cmd.Execute()
}
