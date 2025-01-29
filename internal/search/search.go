package search

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"searchengine/internal/common/request"
	"searchengine/internal/filter"
	"searchengine/internal/index"
	"searchengine/internal/rank"
)

type SearchClient struct {
	indxCli   *index.Index
	rankCli   *rank.RankingClient
	filterCli *filter.FilterClient
}

func NewSearchClient(indxCli *index.Index, rankCli *rank.RankingClient, filterCli *filter.FilterClient) *SearchClient {
	return &SearchClient{
		indxCli:   indxCli,
		rankCli:   rankCli,
		filterCli: filterCli,
	}
}

// SearchIndex выполняет поиск в индексе по заданным параметрам //todo
func (sc *SearchClient) SearchIndex(queryText string, filters map[string]interface{}, sortFields []string) ([]map[string]interface{}, error) {
	// Создаем базовый запрос для полнотекстового поиска
	query := bleve.NewMatchQuery(queryText)
	searchRequest := bleve.NewSearchRequest(query)

	searchRequest.Fields = []string{"*"} // "*" означает вернуть все поля документа

	// Добавляем фильтры (если есть)
	for field, value := range filters {
		termQuery := bleve.NewTermQuery(fmt.Sprintf("%v", value))
		termQuery.SetField(field)
		searchRequest.Query = bleve.NewConjunctionQuery(searchRequest.Query, termQuery)
	}

	// Указываем сортировку (если есть)
	if len(sortFields) > 0 {
		for _, field := range sortFields {
			searchRequest.SortBy([]string{field})
		}
	}

	// Выполняем поиск
	//searchResult, err := index.Search(searchRequest)
	searchResult, err := sc.indxCli.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения поиска: %v", err)
	}

	// Формируем результаты
	var results []map[string]interface{}
	for _, hit := range searchResult.Hits {
		results = append(results, map[string]interface{}{
			"id":     hit.ID,
			"score":  hit.Score,
			"fields": hit.Fields,
		})
	}
	return results, nil
}

// AdvancedSearch выполняет поиск с фильтрами и ранжированием
func (sc *SearchClient) AdvancedSearch(queryText string, filters *request.FilterRequest, sortFields []string) ([]map[string]interface{}, error) {
	mainQuery := bleve.NewMatchQuery(queryText)

	// Применяем фильтры
	filtersQuery, err := sc.filterCli.ApplyFilters(filters)
	if err != nil {
		return nil, fmt.Errorf("ошибка применения фильтров: %v", err)
	}

	combinedQuery := bleve.NewBooleanQuery()
	combinedQuery.AddMust(mainQuery)
	if filtersQuery != nil {
		combinedQuery.AddMust(filtersQuery)
	}

	searchRequest := bleve.NewSearchRequest(combinedQuery)

	searchRequest.Fields = []string{"*"} // "*" означает вернуть все поля документа

	// Применяем ранжирование
	err = sc.rankCli.ApplyRanking(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("ошибка применения ранжирования: %v", err)
	}

	// Выполняем поиск
	//searchResult, err := index.Search(searchRequest)
	searchResult, err := sc.indxCli.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения поиска: %v", err)
	}

	// Формируем результаты
	var results []map[string]interface{}
	for _, hit := range searchResult.Hits {
		results = append(results, map[string]interface{}{
			"id":     hit.ID,
			"score":  hit.Score,
			"fields": hit.Fields,
		})
	}
	return results, nil
}
