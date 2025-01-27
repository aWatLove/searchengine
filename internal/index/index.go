package index

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"log"
	"searchengine/internal/config"
	"searchengine/internal/validate"
	"sync"
)

type Index struct {
	cfg *config.Config

	iCfg *config.IndexConfig

	bIndex bleve.Index

	mu *sync.RWMutex
}

func New(cfg *config.Config, idxCfg *config.IndexConfig) *Index {

	bleveIndex, err := bleve.Open(fmt.Sprintf("%s%s", cfg.IndexPath, cfg.IndexName))
	if err != nil {
		log.Println("[INDEX][ERROR] error while opening:", err)

		mapping := bleve.NewIndexMapping()
		docMapping := bleve.NewDocumentMapping()

		// Создаем поля на основе конфигурации
		for _, field := range idxCfg.Fields {
			fieldMapping := bleve.NewTextFieldMapping()
			fieldMapping.Index = field.Searchable

			if field.Filterable {
				fieldMapping.Store = true
			}

			if field.Sortable {
				fieldMapping.DocValues = true
			}

			docMapping.AddFieldMappingsAt(field.Name, fieldMapping)
		}

		mapping.AddDocumentMapping("document", docMapping)

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

// AddDocument добавляет документ в индекс после валидации
// todo сделать инкрементальный docID
func (i *Index) AddDocument(docID string, document map[string]interface{}) error { //todo
	// Валидация документа
	err := validate.ValidateDocument(i.iCfg, document)
	if err != nil {
		return fmt.Errorf("документ не прошел валидацию: %v", err)
	}

	// Добавляем документ в индекс
	err = i.bIndex.Index(docID, document)
	if err != nil {
		return fmt.Errorf("ошибка добавления документа в индекс: %v", err)
	}

	fmt.Printf("Документ с ID '%s' успешно добавлен в индекс.\n", docID)
	return nil
}

func (i *Index) Search(req *bleve.SearchRequest) (*bleve.SearchResult, error) {
	return i.bIndex.Search(req)
}
