package index

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	mapping2 "github.com/blevesearch/bleve/v2/mapping"
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

	lastIndex uint64
}

func New(cfg *config.Config) *Index {
	bleveIndex, err := bleve.Open(fmt.Sprintf("%s%s", cfg.IndexPath, cfg.IndexCfg.IndexName))
	if err != nil {
		log.Println("[INDEX][ERROR] error while opening:", err)

		mapping := bleve.NewIndexMapping()
		docMapping := bleve.NewDocumentMapping()

		// Создаем поля на основе конфигурации
		for _, field := range cfg.IndexCfg.Fields {
			var fieldMapping *mapping2.FieldMapping

			if field.Type == "timestamp" {
				fieldMapping = bleve.NewDateTimeFieldMapping()
			} else {
				fieldMapping = bleve.NewTextFieldMapping()
			}
			//fieldMapping := bleve.NewTextFieldMapping()

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

		bleveIndex, err = bleve.New(fmt.Sprintf("%s%s", cfg.IndexPath, cfg.IndexCfg.IndexName), mapping)
		if err != nil {
			log.Fatalln("[INDEX][ERROR] error while creating:", err)
		}
	}

	return &Index{
		cfg:    cfg,
		bIndex: bleveIndex,
		iCfg:   cfg.IndexCfg,
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
	i.mu.Lock()
	i.mu.Unlock()
	err = i.bIndex.Index(docID, document)
	if err != nil {
		return fmt.Errorf("ошибка добавления документа в индекс: %v", err)
	}

	fmt.Printf("Документ с ID '%s' успешно добавлен в индекс.\n", docID)
	return nil
}

func (i *Index) Delete(docID string) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.bIndex.Delete(docID)
}

func (i *Index) Update(docID string, document map[string]interface{}) error {
	// Валидация документа
	err := validate.ValidateDocument(i.iCfg, document)
	if err != nil {
		return fmt.Errorf("документ не прошел валидацию: %v", err)
	}

	err = i.Delete(docID)
	if err != nil {
		log.Printf("[INDEX][ERROR] error while deleting docID: '%s' err:\n", docID, err)
	}
	// Добавляем документ в индекс
	i.mu.Lock()
	i.mu.Unlock()
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

func (i *Index) GetAllDoc() ([]map[string]interface{}, error) {
	// Создаем запрос, который соответствует всем документам
	query := bleve.NewMatchAllQuery()

	// Настраиваем параметры поиска
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = 10000           // Максимальное количество документов на странице
	searchRequest.Fields = []string{"*"} // Запрашиваем все поля

	var results []map[string]interface{}

	for {
		// Выполняем поиск
		searchResult, err := i.bIndex.Search(searchRequest)
		if err != nil {
			return nil, err
		}

		// Собираем результаты
		for _, hit := range searchResult.Hits {
			results = append(results, hit.Fields)
		}

		// Проверяем, есть ли еще документы
		if searchResult.Total <= uint64(searchRequest.From+searchRequest.Size) {
			break
		}
		searchRequest.From += searchRequest.Size
	}

	return results, nil
}
