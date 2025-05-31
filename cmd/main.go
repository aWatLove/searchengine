package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"searchengine/internal/config"
	"searchengine/internal/filter"
	"searchengine/internal/index"
	"searchengine/internal/metrics"
	"searchengine/internal/rank"
	"searchengine/internal/search"
	"searchengine/internal/server"
	"searchengine/internal/subscriber"
	"syscall"
	"time"
)

func initLogger() {
	logFile, err := os.OpenFile(fmt.Sprintf("./logs/%s.log", time.Now().Format("2006-01-02_15-04-05")), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// Пишем логи И в файл, И в консоль (опционально)
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.LstdFlags)
}

func main() {
	initLogger()
	go metrics.RecordSysMetrics()

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

	// ====== Subscriber ======
	ctxSubscriber, cancelSubscriber := context.WithCancel(context.Background())
	sub := subscriber.New(cfg, indexCLi)
	sub.Start(ctxSubscriber)
	// ========================

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
	cancelSubscriber()
}
