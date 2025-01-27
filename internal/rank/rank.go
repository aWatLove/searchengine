package rank

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"searchengine/internal/config"
)

const (
	catboostV2  = "catboostV2"
	logarithmic = "logarithmic"
)

type RankingClient struct {
	cfg config.RankConfig
}

func NewRankingClient(cfg config.RankConfig) *RankingClient {
	return &RankingClient{cfg: cfg}
}

// ApplyRanking добавляет настройки ранжирования в запрос Bleve
func (rc *RankingClient) ApplyRanking(searchRequest *bleve.SearchRequest) error {

	var sortOrder []string
	for _, boost := range rc.cfg.Boosts {
		field := boost.Field
		weight := boost.Weight
		boostType := boost.BoostType

		// Пример: Вы можете кастомизировать обработку boost_type здесь
		switch boostType {
		case catboostV2:
			sortOrder = append(sortOrder, fmt.Sprintf("%s^%.2f", field, weight))
		case logarithmic:
			sortOrder = append(sortOrder, fmt.Sprintf("%s^log", field))
		default:
			sortOrder = append(sortOrder, field)
		}
	}

	searchRequest.SortBy(sortOrder)
	return nil
}
