package index

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	mapping2 "github.com/blevesearch/bleve/v2/mapping"
	index "github.com/blevesearch/bleve_index_api"
	"log"
	"os"
	"searchengine/internal/config"
	"searchengine/internal/validate"
	"strings"
	"sync"
)

type Index struct {
	cfg *config.Config

	iCfg *config.IndexConfig

	bIndex bleve.Index

	mu *sync.RWMutex

	lastIndex uint64

	isBuilded bool
}

func New(cfg *config.Config) *Index {
	bleveIndex, err := bleve.Open(fmt.Sprintf("%s%s", cfg.IndexPath, cfg.IndexCfg.IndexName))
	if err != nil {
		log.Println("[INDEX][ERROR] error while opening:", err)

		indexMapping := bleve.NewIndexMapping()
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

		indexMapping.AddDocumentMapping("document", docMapping)

		bleveIndex, err = bleve.New(fmt.Sprintf("%s%s", cfg.IndexPath, cfg.IndexCfg.IndexName), indexMapping)
		if err != nil {
			log.Fatalln("[INDEX][ERROR] error while creating:", err)
		}
	}

	return &Index{
		cfg:       cfg,
		bIndex:    bleveIndex,
		iCfg:      cfg.IndexCfg,
		mu:        new(sync.RWMutex),
		isBuilded: true,
	}
}

func (idx *Index) Add(id string, record interface{}) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	return idx.bIndex.Index(id, record)
}

// AddDocument добавляет документ в индекс после валидации
func (i *Index) AddDocument(docID string, document map[string]interface{}) error {
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

	log.Printf("Документ с ID '%s' успешно добавлен в индекс.\n", docID)
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
		return fmt.Errorf("ошибка обновления документа в индекс: %v", err)
	}

	log.Printf("Документ с ID '%s' успешно обновлен в индексе.\n", docID)
	return nil
}

func (i *Index) Search(req *bleve.SearchRequest) (*bleve.SearchResult, error) {
	return i.bIndex.Search(req)
}

func (i *Index) GetDocId(id string) (index.Document, error) {
	return i.bIndex.Document(id)
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

func (i *Index) ReindexBleve() error {
	tmpIndexPath := fmt.Sprintf("%s%s", i.cfg.IndexPath, i.cfg.IndexCfg.IndexName) + "_tmp"
	_ = os.RemoveAll(tmpIndexPath)

	oldIndex := i.bIndex
	oldMappingIface := oldIndex.Mapping()
	oldMapping, ok := oldMappingIface.(*mapping2.IndexMappingImpl)
	if !ok {
		return fmt.Errorf("failed to cast index mapping to IndexMappingImpl")
	}

	// Копируем маппинг, чтобы не мутировать существующий
	newMapping := *oldMapping

	synonymCollection := "my_synonym_collection"
	synonymSourceName := "my_synonyms"

	// Добавляем synonym source в новый маппинг ДО создания индекса
	err := newMapping.AddSynonymSource(synonymSourceName, map[string]interface{}{
		"collection": synonymCollection,
	})
	if err != nil && !strings.Contains(err.Error(), "already defined") {
		return fmt.Errorf("AddSynonymSource() failed: %v", err)
	}

	// Привязываем SynonymSource к полям по умолчанию
	for _, field := range newMapping.DefaultMapping.Fields {
		field.SynonymSource = synonymSourceName
	}

	// Создаём новый индекс
	newIndex, err := bleve.New(tmpIndexPath, &newMapping)
	if err != nil {
		return fmt.Errorf("failed to create new index with updated mapping: %w", err)
	}

	// Добавляем синонимы в индекс
	synDef := &bleve.SynonymDefinition{
		Synonyms: []string{"кепка", "шапка", "бейсболка", "панама"},
	}

	if synIndex, ok := newIndex.(bleve.SynonymIndex); ok {
		err = synIndex.IndexSynonym("synDoc1", synonymCollection, synDef)
		if err != nil {
			return fmt.Errorf("failed to index synonym: %w", err)
		}
	} else {
		return fmt.Errorf("index does not support synonym indexing")
	}

	// Переносим документы из старого индекса
	query := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(query)
	sessionSize := 10000
	searchRequest.Size = sessionSize
	searchRequest.Fields = []string{"*"}
	searchRequest.From = 0
	count := 0
	for {
		res, err := oldIndex.Search(searchRequest)
		if err != nil {
			return fmt.Errorf("search error: %w", err)
		}
		if len(res.Hits) == 0 {
			break
		}
		for _, hit := range res.Hits {
			id := hit.ID
			doc := hit.Fields
			err = newIndex.Index(id, doc)
			if err != nil {
				log.Printf("failed to reindex doc %s: %v", id, err)
				continue
			}
			count++
			if count%1000 == 0 {
				log.Printf("Reindexed %d documents...", count)
			}
		}
		searchRequest.From += sessionSize
	}

	err = newIndex.Close()
	if err != nil {
		return fmt.Errorf("failed to close new index: %w", err)
	}
	err = oldIndex.Close()
	if err != nil {
		return fmt.Errorf("failed to close old index: %w", err)
	}
	err = os.RemoveAll(fmt.Sprintf("%s%s", i.cfg.IndexPath, i.cfg.IndexCfg.IndexName))
	if err != nil {
		return fmt.Errorf("failed to delete old index: %w", err)
	}
	err = os.Rename(tmpIndexPath, fmt.Sprintf("%s%s", i.cfg.IndexPath, i.cfg.IndexCfg.IndexName))
	if err != nil {
		return fmt.Errorf("failed to rename new index: %w", err)
	}
	i.bIndex, err = bleve.Open(fmt.Sprintf("%s%s", i.cfg.IndexPath, i.cfg.IndexCfg.IndexName))
	if err != nil {
		return fmt.Errorf("failed to reopen new index: %w", err)
	}

	log.Printf("Reindexing complete. Total documents reindexed: %d", count)
	return nil
}

func (i *Index) RebuildIndex() error {
	tmpIndexPath := fmt.Sprintf("%s%s", i.cfg.IndexPath, i.cfg.IndexCfg.IndexName) + "_tmp"
	_ = os.RemoveAll(tmpIndexPath)

	oldIndex := i.bIndex

	// Создаём новый индекс
	indexMapping := bleve.NewIndexMapping()
	docMapping := bleve.NewDocumentMapping()

	// Создаем поля на основе конфигурации
	for _, field := range i.cfg.IndexCfg.Fields {
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

	indexMapping.AddDocumentMapping("document", docMapping)

	newIndex, err := bleve.New(tmpIndexPath, indexMapping)
	if err != nil {
		log.Fatalln("[INDEX][ERROR] error while creating:", err)
	}

	// Переносим документы из старого индекса
	query := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(query)
	sessionSize := 10000
	searchRequest.Size = sessionSize
	searchRequest.Fields = []string{"*"}
	searchRequest.From = 0
	count := 0
	for {
		res, err := oldIndex.Search(searchRequest)
		if err != nil {
			return fmt.Errorf("search error: %w", err)
		}
		if len(res.Hits) == 0 {
			break
		}
		for _, hit := range res.Hits {
			id := hit.ID
			doc := hit.Fields
			err = newIndex.Index(id, doc)
			if err != nil {
				log.Printf("failed to reindex doc %s: %v", id, err)
				continue
			}
			count++
			if count%1000 == 0 {
				log.Printf("Reindexed %d documents...", count)
			}
		}
		searchRequest.From += sessionSize
	}

	err = newIndex.Close()
	if err != nil {
		return fmt.Errorf("failed to close new index: %w", err)
	}
	err = oldIndex.Close()
	if err != nil {
		return fmt.Errorf("failed to close old index: %w", err)
	}
	err = os.RemoveAll(fmt.Sprintf("%s%s", i.cfg.IndexPath, i.cfg.IndexCfg.IndexName))
	if err != nil {
		return fmt.Errorf("failed to delete old index: %w", err)
	}
	err = os.Rename(tmpIndexPath, fmt.Sprintf("%s%s", i.cfg.IndexPath, i.cfg.IndexCfg.IndexName))
	if err != nil {
		return fmt.Errorf("failed to rename new index: %w", err)
	}
	i.bIndex, err = bleve.Open(fmt.Sprintf("%s%s", i.cfg.IndexPath, i.cfg.IndexCfg.IndexName))
	if err != nil {
		return fmt.Errorf("failed to reopen new index: %w", err)
	}

	log.Printf("Reindexing complete. Total documents reindexed: %d\n", count)
	log.Printf("Complete rebuilding index\n")
	i.isBuilded = true
	return nil
}

func (i *Index) SetNeedRebuild() {
	i.isBuilded = false
}

func (i *Index) SetBuilded() {
	i.isBuilded = true
}

func (i Index) IsBuilded() bool {
	return i.isBuilded
}
