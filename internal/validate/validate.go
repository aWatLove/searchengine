package validate

import (
	"fmt"
	"searchengine/internal/config"
)

// ValidateDocument проверяет соответствие данных конфигурации полей
func ValidateDocument(config *config.IndexConfig, document map[string]interface{}) error { // todo не правильная проверка
	for _, field := range config.Fields {
		value, exists := document[field.Name]
		if !exists {
			return fmt.Errorf("поле '%s' отсутствует в документе", field.Name)
		}

		// Проверяем тип поля
		if !validateFieldType(field.Type, value) {
			return fmt.Errorf("поле '%s' имеет некорректный тип: ожидался '%s', получен '%T'", field.Name, field.Type, value)
		}
	}
	return nil
}

// validateFieldType проверяет соответствие типа значения ожидаемому
func validateFieldType(expectedType string, value interface{}) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "timestamp":
		_, ok := value.(string) // Допустим, timestamp передается как строка
		return ok
	case "number":
		_, ok := value.(float64) // JSON числа парсятся как float64
		return ok
	case "bool":
		_, ok := value.(bool) // JSON числа парсятся как bool
		return ok
	default:
		return false
	}
}

func ValidateSortField(cfg *config.Config, sortField string) bool {
	for _, f := range cfg.IndexCfg.Fields {
		if f.Name == sortField {
			return f.Sortable
		}
	}
	return false
}
