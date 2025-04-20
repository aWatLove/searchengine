package filter

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"log"
	"searchengine/internal/common/request"
	"searchengine/internal/config"
	"strconv"
	"strings"
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
func (fc *FilterClient) ApplyFilters(filters *request.FilterRequest) (*query.BooleanQuery, error) {
	if filters == nil {
		return nil, nil
	}
	if len(filters.Range) == 0 && len(filters.MultiSelect) == 0 && len(filters.OneSelect) == 0 &&
		len(filters.BoolSelect) == 0 && len(filters.Category) == 0 {
		return nil, nil
	}

	combinedFilter := bleve.NewBooleanQuery()

	// category filter
	if filters.Category != "" {
		filterCategory := bleve.NewTermQuery(strings.ToLower(filters.Category))
		filterCategory.SetField("category")

		combinedFilter.AddMust(filterCategory)
	}

	if len(filters.Range) != 0 {
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
				filterDate := bleve.NewDateRangeQuery(fromDate, toDate) //todo
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
		filterRates := bleve.NewConjunctionQuery(rangeFilters...)

		combinedFilter.AddMust(filterRates)
	}

	if len(filters.MultiSelect) != 0 {
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
	}

	if len(filters.OneSelect) != 0 {
		// one-select filters
		osFilters := make([]query.Query, 0, len(filters.OneSelect))
		for _, o := range filters.OneSelect {
			osf := bleve.NewTermQuery(o.Value)
			osf.SetField(o.Name)
			osFilters = append(osFilters, osf)
		}
		filtersOneSelect := bleve.NewConjunctionQuery(osFilters...)

		combinedFilter.AddMust(filtersOneSelect)
	}

	if len(filters.BoolSelect) != 0 {
		// bool-select filters
		boolFilters := make([]query.Query, 0, len(filters.BoolSelect))
		for _, bs := range filters.BoolSelect {
			bsf := bleve.NewBoolFieldQuery(bs.Value)
			bsf.SetField(bs.Name)
			boolFilters = append(boolFilters, bsf)
		}
		filtersBool := bleve.NewConjunctionQuery(boolFilters...)

		combinedFilter.AddMust(filtersBool)
	}

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

func (fc *FilterClient) GetByCategory(category string) (config.FilterConfig, bool) {
	f, ok := fc.filters[category]
	return f, ok
}

func (fc *FilterClient) GetAllCategories() []string {
	categories := make([]string, 0, len(fc.filters))
	for category := range fc.filters {
		categories = append(categories, category)
	}
	return categories
}
