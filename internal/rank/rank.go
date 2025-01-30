package rank

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"searchengine/internal/common/constants"
	"searchengine/internal/config"
)

const (
	customBoost = "custom"
	catboostV2  = "catboostV2"
	logarithmic = "logarithmic"
)

type RankingClient struct {
	cfg *config.RankConfig
}

func New(cfg *config.RankConfig) *RankingClient {
	return &RankingClient{cfg: cfg}
}

// ApplyRanking добавляет настройки ранжирования в запрос Bleve
func (rc *RankingClient) ApplyRanking(searchRequest *bleve.SearchRequest, sortField string, sortOrder string) error {
	var sortOrderList []string

	// Если передана сортировка, добавляем ее
	if sortField != "" && sortOrder != "" {
		if sortOrder == constants.SortOrderDesc {
			// Сортировка по убыванию
			sortOrderList = append(sortOrderList, fmt.Sprintf("-%s", sortField))
		} else if sortOrder == constants.SortOrderAsc {
			// Сортировка по возрастанию
			sortOrderList = append(sortOrderList, sortField)
		} else {
			// Некорректный порядок сортировки
			return fmt.Errorf("invalid sort order: %s. Expected 'asc' or 'desc'", sortOrder)
		}
	} else {
		// Если сортировка не передана, применяем ранжирование
		for _, boost := range rc.cfg.Boosts {
			field := boost.Field
			weight := boost.Weight
			boostType := boost.BoostType

			switch boostType {
			case customBoost:
				sortOrderList = append(sortOrderList, fmt.Sprintf("-%s", field))
			case catboostV2:
				sortOrderList = append(sortOrderList, fmt.Sprintf("%s^%.2f", field, weight))
			case logarithmic:
				sortOrderList = append(sortOrderList, fmt.Sprintf("%s^log", field))
			default:
				sortOrderList = append(sortOrderList, field)
			}
		}
	}

	// Применяем сортировку или ранжирование
	searchRequest.SortBy(sortOrderList)
	return nil
}

//// ValidateFormula проверяет, что формула корректна и содержит шаблоны $F и $W. // todo
//func ValidateFormula(formula string) error {
//	// Проверяем, что формула не пустая
//	if formula == "" {
//		return errors.New("formula is empty")
//	}
//
//	// Проверяем наличие шаблонов $F и $W
//	if !strings.Contains(formula, "$F") {
//		return errors.New("formula must contain $F (field)")
//	}
//	if !strings.Contains(formula, "$W") {
//		return errors.New("formula must contain $W (weight)")
//	}
//
//	// Проверяем, что формула не содержит недопустимых символов
//	// (например, можно добавить дополнительные проверки)
//	invalidChars := []string{"$", "{", "}", "(", ")", "[", "]", "&", "|", "!", "@", "#", "%", "?", "`", "~"}
//	for _, char := range invalidChars {
//		if strings.Contains(formula, char) && char != "$" {
//			return fmt.Errorf("formula contains invalid character: %s", char)
//		}
//	}
//
//	// Проверяем, что формула может быть использована для подстановки
//	testField := "testField"
//	testWeight := 1.0
//	_, err := ApplyFormula(formula, testField, testWeight)
//	if err != nil {
//		return fmt.Errorf("formula is invalid: %v", err)
//	}
//
//	return nil
//}
//
//// ApplyFormula подставляет значения поля и веса в формулу.
//func ApplyFormula(formula, field string, weight float64) (string, error) {
//	result := strings.ReplaceAll(formula, "$F", field)
//	result = strings.ReplaceAll(result, "$W", fmt.Sprintf("%.2f", weight))
//
//	if strings.Contains(result, "$F") || strings.Contains(result, "$W") {
//		return "", errors.New("formula still contains $F or $W after substitution")
//	}
//
//	return result, nil
//}
