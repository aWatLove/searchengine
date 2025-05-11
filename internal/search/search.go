package search

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"searchengine/internal/common/request"
	"searchengine/internal/filter"
	"searchengine/internal/index"
	"searchengine/internal/rank"
	"strings"
)

type SearchClient struct {
	indxCli   *index.Index
	RankCli   *rank.RankingClient
	filterCli *filter.FilterClient
}

func NewSearchClient(indxCli *index.Index, rankCli *rank.RankingClient, filterCli *filter.FilterClient) *SearchClient {
	return &SearchClient{
		indxCli:   indxCli,
		RankCli:   rankCli,
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
//func (sc *SearchClient) AdvancedSearch(queryText string, filters *request.FilterRequest, sortField string, sortOrder string) ([]map[string]interface{}, error) {
//
//	mainQuery := bleve.NewFuzzyQuery(queryText)
//	mainQuery.Fuzziness = 1
//
//	// Применяем фильтры
//	filtersQuery, err := sc.filterCli.ApplyFilters(filters)
//	if err != nil {
//		return nil, fmt.Errorf("ошибка применения фильтров: %v", err)
//	}
//
//	combinedQuery := bleve.NewBooleanQuery()
//	combinedQuery.AddMust(mainQuery)
//	if filtersQuery != nil {
//		combinedQuery.AddMust(filtersQuery)
//	}
//
//	searchRequest := bleve.NewSearchRequest(combinedQuery)
//
//	searchRequest.Fields = []string{"*"} // "*" означает вернуть все поля документа
//
//	// Применяем ранжирование
//	err = sc.RankCli.ApplyRanking(searchRequest, sortField, sortOrder)
//	if err != nil {
//		return nil, fmt.Errorf("ошибка применения ранжирования: %v", err)
//	}
//
//	// Выполняем поиск
//	searchResult, err := sc.indxCli.Search(searchRequest)
//	if err != nil {
//		return nil, fmt.Errorf("ошибка выполнения поиска: %v", err)
//	}
//
//	// Формируем результаты
//	var results []map[string]interface{}
//	for _, hit := range searchResult.Hits {
//		results = append(results, map[string]interface{}{
//			"id":     hit.ID,
//			"score":  hit.Score,
//			"fields": hit.Fields,
//		})
//	}
//	return results, nil
//}

func (sc *SearchClient) AdvancedSearch(queryText string, filters *request.FilterRequest, sortField string, sortOrder string) ([]map[string]interface{}, error) {
	// Разделяем запрос на отдельные термины
	terms := strings.Fields(queryText)
	booleanQuery := bleve.NewBooleanQuery()

	// Добавляем каждый термин как отдельный MatchQuery с Fuzzy
	for _, term := range terms {
		termQuery := bleve.NewMatchQuery(term)
		termQuery.Fuzziness = 1
		booleanQuery.AddShould(termQuery) // Используем Should для логического OR
	}

	// Применяем фильтры
	filtersQuery, err := sc.filterCli.ApplyFilters(filters)
	if err != nil {
		return nil, fmt.Errorf("ошибка применения фильтров: %v", err)
	}

	// Комбинируем поиск и фильтры
	combinedQuery := bleve.NewBooleanQuery()
	if len(terms) > 0 {
		combinedQuery.AddMust(booleanQuery)
	}
	if filtersQuery != nil {
		combinedQuery.AddMust(filtersQuery)
	}

	searchRequest := bleve.NewSearchRequest(combinedQuery)
	searchRequest.Fields = []string{"*"}

	// Применяем сортировку
	if err := sc.RankCli.ApplyRanking(searchRequest, sortField, sortOrder); err != nil {
		return nil, fmt.Errorf("ошибка сортировки: %v", err)
	}

	// Выполняем поиск
	searchResult, err := sc.indxCli.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска: %v", err)
	}

	// Формируем результаты
	results := make([]map[string]interface{}, 0, len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		results = append(results, map[string]interface{}{
			"id":     hit.ID,
			"score":  hit.Score,
			"fields": hit.Fields,
		})
	}
	return results, nil
}
