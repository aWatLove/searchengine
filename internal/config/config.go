package config

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"log"
	"os"
)

type Config struct {
	TestEnv   string `envconfig:"TEST_ENV" required:"true"`
	IndexPath string `envconfig:"INDEX_PATH" required:"true"`
	IndexName string `envconfig:"INDEX_NAME" required:"true"`

	PrivatePort string `envconfig:"PRIVATE_PORT" required:"true"`
	PublicPort  string `envconfig:"PUBLIC_PORT" required:"true"`

	// filter
	DateLayout string `envconfig:"DATE_LAYOUT" required:"true"`
}

func LoadConfig() *Config {

	for _, fileName := range []string{".env.local", ".env"} {
		err := godotenv.Load(fileName)
		if err != nil {
			log.Println("[CONFIG][ERROR]:", err)
		}
	}

	cfg := Config{}

	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatalln("[CONFIG][ERROR]:", err)
	}

	cfg.PrintConfig()

	return &cfg
}

func (c *Config) PrintConfig() {
	log.Println("===================== CONFIG =====================")
	log.Println("TEST_ENV...................... ", c.TestEnv)
	log.Println("INDEX_PATH.................... ", c.IndexPath)
	log.Println("INDEX_NAME.................... ", c.IndexName)
	log.Println("==================================================")
}

// Index

// FieldConfig описывает конфигурацию для поля индекса
type FieldConfig struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Searchable bool   `json:"searchable"`
	Filterable bool   `json:"filterable,omitempty"`
	Sortable   bool   `json:"sortable,omitempty"`
}

// IndexConfig описывает конфигурацию индекса
type IndexConfig struct {
	IndexName string        `json:"indexName"`
	Category  []string      `json:"category,omitempty"`
	Fields    []FieldConfig `json:"fields"`
}

// LoadConfig загружает конфигурацию индекса из файла
func LoadIndexConfig(filePath string) (*IndexConfig, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config IndexConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Ranking
type RankConfig struct {
	Boosts []BoostConfig `json:"boosts"`
}

type BoostConfig struct {
	Field     string  `json:"field"`
	Weight    float64 `json:"weight"`
	BoostType string  `json:"boost_type"`
}

// LoadConfig загружает конфигурацию индекса из файла
func LoadRankConfig(filePath string) (*RankConfig, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config RankConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Filters
type FilterConfig struct {
	Category    string              `json:"category"`
	Range       []RangeFilter       `json:"range"`
	MultiSelect []MultiSelectFilter `json:"multi-select"`
	OneSelect   []OneSelectFilter   `json:"one-select"`
	BoolSelect  []BoolSelectFilter  `json:"bool-select"`
}

type BoolSelectFilter struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

type RangeFilter struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	FromValue string `json:"from_value"`
	ToValue   string `json:"to_value"`
}

type MultiSelectFilter struct {
	Name  string   `json:"name"`
	Value []string `json:"value"`
}

type OneSelectFilter struct {
	Name  string   `json:"name"`
	Value []string `json:"value"`
}

func LoadFilterConfig(filePath string) ([]FilterConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var configs []FilterConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}
