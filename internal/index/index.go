package index

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"log"
	"searchengine/internal/config"
	"sync"
)

type Index struct {
	cfg *config.Config

	bIndex bleve.Index

	mu *sync.RWMutex
}

func New(cfg *config.Config) *Index {

	bleveIndex, err := bleve.Open(fmt.Sprintf("%s%s", cfg.IndexPath, cfg.IndexName))
	if err != nil {
		log.Println("[INDEX][ERROR] error while opening:", err)

		mapping := bleve.NewIndexMapping()

		bleveIndex, err = bleve.New(fmt.Sprintf("%s%s", cfg.IndexPath, cfg.IndexName), mapping)
		if err != nil {
			log.Fatalln("[INDEX][ERROR] error while creating:", err)
		}
	}

	return &Index{
		cfg:    cfg,
		bIndex: bleveIndex,
		mu:     new(sync.RWMutex),
	}
}

func (idx *Index) Add(id string, record interface{}) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	return idx.bIndex.Index(id, record)
}
