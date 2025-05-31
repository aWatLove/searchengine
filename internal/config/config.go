package config

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"log"
	"os"
)

type Config struct {
	IndexPath string `envconfig:"INDEX_PATH" required:"true"`

	PrivatePort string `envconfig:"PRIVATE_PORT" required:"true"`
	PublicPort  string `envconfig:"PUBLIC_PORT" required:"true"`

	CfgDirPath string `envconfig:"CONFIG_DIR_PATH" required:"true"`

	// Index
	IndexCfg        *IndexConfig
	IndexConfigPath string `envconfig:"INDEX_CONFIG_PATH" required:"true"`

	// filter
	DateLayout       string `envconfig:"DATE_LAYOUT" required:"true"`
	FilterConfigPath string `envconfig:"FILTER_CONFIG_PATH" required:"true"`
	FilterCfg        []FilterConfig

	// ranking
	RankCfg        *RankConfig
	RankConfigPath string `envconfig:"RANK_CONFIG_PATH" required:"true"`

	// logs
	LogsDir string `envconfig:"LOGS_DIR" required:"true"`

	// subscriber
	EnableNatsSubscriber  bool
	NatsURL               string `envconfig:"NATS_URL"`
	NatsSubject           string `envconfig:"NATS_SUBJECT"`
	EnableKafkaSubscriber bool
	KafkaURL              string `envconfig:"KAFKA_URL"`
	KafkaTopic            string `envconfig:"KAFKA_TOPIC"`
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

	cfg.IndexCfg, err = LoadIndexConfig(fmt.Sprintf("%s%s", cfg.CfgDirPath, cfg.IndexConfigPath))
	if err != nil {
		log.Fatalln("[CONFIG][ERROR] error while loading index config:", err)
	}

	cfg.FilterCfg, err = LoadFilterConfig(fmt.Sprintf("%s%s", cfg.CfgDirPath, cfg.FilterConfigPath))
	if err != nil {
		log.Fatalln("[CONFIG][ERROR] error while loading filter config:", err)
	}

	cfg.RankCfg, err = LoadRankConfig(fmt.Sprintf("%s%s", cfg.CfgDirPath, cfg.RankConfigPath))
	if err != nil {
		log.Fatalln("[CONFIG][ERROR] error while loading rank config:", err)
	}

	if cfg.NatsURL != "" && cfg.NatsSubject != "" {
		cfg.EnableNatsSubscriber = true
	}
	if cfg.KafkaURL != "" && cfg.KafkaTopic != "" {
		cfg.EnableKafkaSubscriber = true
	}

	cfg.PrintConfig()

	return &cfg
}

func (c *Config) PrintConfig() {
	log.Println("===================== CONFIG =====================")
	log.Println("CONFIG_DIR_PATH............... ", c.CfgDirPath)
	log.Println("_____________INDEX____________ ")
	log.Println("INDEX_PATH.................... ", c.IndexPath)
	log.Println("INDEX_NAME.................... ", c.IndexCfg.IndexName)
	log.Println("INDEX_CONFIG_PATH............. ", c.IndexConfigPath)
	log.Println("_____________FILTER____________ ")
	log.Println("FILTER_CONFIG_PATH............. ", c.FilterConfigPath)
	log.Println("DATE_LAYOUT.................... ", c.DateLayout)
	log.Println("______________RANK_____________ ")
	log.Println("RANK_CONFIG_PATH............... ", c.RankConfigPath)
	if c.EnableNatsSubscriber || c.EnableKafkaSubscriber {
		log.Println("___________SUBSCRIBER__________ ")
	}
	if c.EnableNatsSubscriber {
		log.Println("NATS_ENABLED................ ", fmt.Sprintf("%v", c.EnableNatsSubscriber))
		log.Println("NATS_URL.................... ", c.NatsURL)
		log.Println("NATS_SUBJECT................ ", c.NatsSubject)
	}
	if c.EnableKafkaSubscriber {
		log.Println("KAFKA_ENABLED............. ", fmt.Sprintf("%v", c.EnableKafkaSubscriber))
		log.Println("KAFKA_URL................. ", c.KafkaURL)
		log.Println("KAFKA_TOPIC................ ", c.KafkaTopic)
	}
	log.Println("_____________SERVER____________ ")
	log.Println("PRIVATE_PORT................... ", c.PrivatePort)
	log.Println("PUBLIC_PORT.................... ", c.PublicPort)

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
	Synonym    bool   `json:"synonym,omitempty"`
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

	return LoadAnyConfigData[*IndexConfig](file)
}

func LoadAnyConfigData[T any](data []byte) (T, error) {
	var config T
	err := json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

// Ranking
type RankConfig struct {
	Boosts []BoostConfig `json:"boosts"`
}

type BoostConfig struct {
	Field     string  `json:"field"`
	Weight    float64 `json:"weight"`
	BoostType string  `json:"boost_type"`
	Formula   string  `json:"formula,omitempty"`
}

// LoadConfig загружает конфигурацию индекса из файла
func LoadRankConfig(filePath string) (*RankConfig, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return LoadAnyConfigData[*RankConfig](file)
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
	Name string `json:"name"`
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

	return LoadAnyConfigData[[]FilterConfig](data)
}
