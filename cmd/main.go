package main

import (
	"log"
	"searchengine/internal/config"
	"searchengine/internal/index"
)

func main() {

	cfg := config.LoadConfig()

	storage := index.New(cfg)

	message := struct {
		Id   string
		From string
		Body string
	}{
		Id:   "example",
		From: "marty.schoch@gmail.com",
		Body: "bleve indexing is easy",
	}

	err := storage.Add(message.Id, message)
	if err != nil {
		log.Fatalln(err)
	}

}
