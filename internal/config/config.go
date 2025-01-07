package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"log"
)

type Config struct {
	TestEnv   string `envconfig:"TEST_ENV" required:"true"`
	IndexPath string `envconfig:"INDEX_PATH" required:"true"`
	IndexName string `envconfig:"INDEX_NAME" required:"true"`
}

func LoadConfig() *Config {

	for _, fileName := range []string{".env.local", ".env"} {
		err := godotenv.Load(fileName)
		if err != nil {
			log.Println("[CONFIG][ERROR]: ", err)
		}
	}

	cfg := Config{}

	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatalln(err)
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
