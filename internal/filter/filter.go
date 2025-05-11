package filter

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
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

func (fc *FilterClient) RebuildFilters(filtersConfig []config.FilterConfig) error {
	filters := make(map[string]config.FilterConfig)
	for _, f := range filtersConfig {
		filters[f.Category] = f
	}

	fc.filters = filters
	fc.FiltersConfig = filtersConfig
	return nil
}

// ApplyFilters добавляет фильтры в запрос Bleve
func (fc *FilterClient) ApplyFilters(filters *request.FilterRequest) (query.Query, error) {
	if filters == nil {
		return nil, nil
	}
	if len(filters.Range) == 0 && len(filters.MultiSelect) == 0 && len(filters.OneSelect) == 0 &&
		len(filters.BoolSelect) == 0 && len(filters.Category) == 0 {
		return nil, nil
	}

	combinedFilter := bleve.NewBooleanQuery()

	// Category filter
	if filters.Category != "" {
		categoryFilter := bleve.NewTermQuery(strings.ToLower(filters.Category))
		categoryFilter.SetField("category")
		combinedFilter.AddMust(categoryFilter)
	}

	// Range filters (numbers and dates)
	if len(filters.Range) > 0 {
		rangeQueries := bleve.NewBooleanQuery()
		for _, r := range filters.Range {
			var q query.Query
			var err error

			switch r.Type {
			case "timestamp":
				q, err = fc.buildDateRangeQuery(r)
			case "number":
				q, err = fc.buildNumericRangeQuery(r)
			default:
				return nil, fmt.Errorf("unsupported range type: %s", r.Type)
			}

			if err != nil {
				return nil, fmt.Errorf("range filter error (%s): %v", r.Name, err)
			}

			rangeQueries.AddShould(q) // OR между разными полями
		}
		combinedFilter.AddMust(rangeQueries)
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

func (fc *FilterClient) buildDateRangeQuery(r config.RangeFilter) (query.Query, error) {
	// Парсим даты без получения указателя
	fromDate, err := time.Parse(time.RFC3339, r.FromValue)
	if err != nil {
		return nil, fmt.Errorf("invalid from date: %s", r.FromValue)
	}

	toDate, err := time.Parse(time.RFC3339, r.ToValue)
	if err != nil {
		return nil, fmt.Errorf("invalid to date: %s", r.ToValue)
	}

	// Передаем значения времени напрямую, не используя указатели
	dateQuery := bleve.NewDateRangeQuery(fromDate, toDate)
	dateQuery.SetField(r.Name)
	return dateQuery, nil
}

func (fc *FilterClient) buildNumericRangeQuery(r config.RangeFilter) (query.Query, error) {
	// Для числовых значений оставляем указатели
	min, err := strconv.ParseFloat(r.FromValue, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid min value: %s", r.FromValue)
	}

	max, err := strconv.ParseFloat(r.ToValue, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid max value: %s", r.ToValue)
	}

	// Для числовых диапазонов передаем указатели
	numQuery := bleve.NewNumericRangeQuery(&min, &max)
	numQuery.SetField(r.Name)
	return numQuery, nil
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
