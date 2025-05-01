package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"searchengine/internal/config"
	"searchengine/internal/filter"
	"searchengine/internal/index"
	"searchengine/internal/rank"
	"searchengine/internal/search"
	"searchengine/internal/server"
	"syscall"
	"time"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// ====== Config ======
	log.Println("[SERVICE] INITIALIZING CONFIG")
	cfg := config.LoadConfig()
	// ====================

	// ====== Index ======
	log.Println("[SERVICE] INITIALIZING INDEX")
	indexCLi := index.New(cfg)
	// ====================

	// ====== Filter ======
	log.Println("[SERVICE] INITIALIZING FILTER CLIENT")
	filterCli := filter.New(cfg)
	// ====================

	// ====== Ranking ======
	log.Println("[SERVICE] INITIALIZING RANKING CLIENT")
	rankCli := rank.New(cfg.RankCfg)
	// ====================

	// ====== Search Client ======
	log.Println("[SERVICE] INITIALIZING SEARCH CLIENT")
	searchCli := search.NewSearchClient(indexCLi, rankCli, filterCli)
	// ====================

	// ====== Server ======
	log.Println("[SERVICE] START SERVER")
	srv := server.New(cfg, indexCLi, searchCli, filterCli)
	log.Println("[SERVER] Start")
	srv.Start()
	// ====================

	<-stop
	ctxClose, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := srv.Stop(ctxClose)
	if err != nil {
		log.Fatalln("[SERVER][ERROR] error while stopping: ", err)
	}
}
