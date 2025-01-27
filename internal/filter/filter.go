package filter

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"log"
	"searchengine/internal/common/request"
	"searchengine/internal/config"
	"strconv"
	"time"
)

type FilterClient struct {
	cfg *config.Config

	FiltersConfig []config.FilterConfig `json:"filters"`

	filters map[string]config.FilterConfig
}

func New(cfg *config.Config) *FilterClient {
	filters := make(map[string]config.FilterConfig)
	for _, f := range cfg.FilterCfg {
		filters[f.Category] = f
	}

	return &FilterClient{cfg: cfg, FiltersConfig: cfg.FilterCfg, filters: filters}
}

// ApplyFilters добавляет фильтры в запрос Bleve
func (fc *FilterClient) ApplyFilters(filters request.FilterRequest) (*query.BooleanQuery, error) {
	combinedFilter := bleve.NewBooleanQuery()

	// category filter
	filterCategory := bleve.NewTermQuery(filters.Category)
	filterCategory.SetField("category")

	combinedFilter.AddMust(filterCategory)

	// range filters
	rangeFilters := make([]query.Query, 0, len(filters.Range))
	for _, r := range filters.Range {
		if r.Type == "date" {
			// Преобразуем FromValue и ToValue в time.Time
			fromDate, errFrom := fc.parseDate(r.FromValue)
			toDate, errTo := fc.parseDate(r.ToValue)
			if errFrom != nil || errTo != nil {
				log.Println("[FILTER][ERROR] Error parsing date range")
				continue
			}

			// Создаем DateRangeQuery
			filterDate := bleve.NewDateRangeQuery(fromDate, toDate)
			filterDate.SetField(r.Name)

			// Добавляем в список фильтров
			rangeFilters = append(rangeFilters, filterDate)
		} else {
			minV := parseNumeric(r.FromValue)
			maxV := parseNumeric(r.ToValue)
			rf := bleve.NewNumericRangeQuery(minV, maxV)
			rf.SetField(r.Name)
			rangeFilters = append(rangeFilters, rf)
		}

	}
	filterRates := bleve.NewDisjunctionQuery(rangeFilters...)

	combinedFilter.AddMust(filterRates)

	// multi-select filters
	msFilters := make([]query.Query, 0, len(filters.MultiSelect))
	for _, ms := range filters.MultiSelect {
		msValues := make([]query.Query, 0, len(ms.Value))
		for _, val := range ms.Value {
			msf := bleve.NewTermQuery(val)
			msf.SetField(ms.Name)
			msValues = append(msValues, msf)
		}
		filterMultiSelect := bleve.NewDisjunctionQuery(msValues...)

		msFilters = append(msFilters, filterMultiSelect)

	}
	filtersMultiSelect := bleve.NewConjunctionQuery(msFilters...)

	combinedFilter.AddMust(filtersMultiSelect)

	// one-select filters
	osFilters := make([]query.Query, 0, len(filters.OneSelect))
	for _, o := range filters.OneSelect {
		osf := bleve.NewTermQuery(o.Value)
		osf.SetField(o.Name)
		osFilters = append(osFilters, osf)
	}
	filtersOneSelect := bleve.NewConjunctionQuery(osFilters...)

	combinedFilter.AddMust(filtersOneSelect)

	// bool-select filters
	boolFilters := make([]query.Query, 0, len(filters.BoolSelect))
	for _, bs := range filters.BoolSelect {
		bsf := bleve.NewTermQuery(fmt.Sprintf("%v", bs.Value))
		bsf.SetField(bs.Name)
		boolFilters = append(boolFilters, bsf)
	}
	filtersBool := bleve.NewConjunctionQuery(boolFilters...)

	combinedFilter.AddMust(filtersBool)

	return combinedFilter, nil
}

// Вспомогательные функции
func parseNumeric(value string) *float64 {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	return &val
}

// parseDate преобразует строку в time.Time
func (fc *FilterClient) parseDate(dateStr string) (time.Time, error) {
	layout := fc.cfg.DateLayout
	return time.Parse(layout, dateStr)
}

func toStringSlice(input []interface{}) []string {
	var result []string
	for _, v := range input {
		result = append(result, v.(string))
	}
	return result
}
