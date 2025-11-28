package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config структура для конфигурации
type Config struct {
	Bot struct {
		Token   string `yaml:"token"`
		Debug   bool   `yaml:"debug"`
		Timeout int    `yaml:"timeout"`
	} `yaml:"bot"`

	Messages struct {
		Welcome string `yaml:"welcome"`
		Help    string `yaml:"help"`
		About   string `yaml:"about"`
	} `yaml:"messages"`

	Images struct {
		Path          string `yaml:"path"`
		DailyReminder bool   `yaml:"daily_reminder"`
	} `yaml:"images"`
}

// LoadConfig загружает конфигурацию из YAML файла
func LoadConfig(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	// Проверяем переменную окружения для токена (приоритет над config.yaml)
	if envToken := os.Getenv("BOT_TOKEN"); envToken != "" {
		config.Bot.Token = envToken
	}

	// Валидация конфигурации
	if config.Bot.Token == "" || config.Bot.Token == "YOUR_BOT_TOKEN_HERE" {
		log.Fatal("Bot token is not set. Please set BOT_TOKEN environment variable or configure it in config.yaml")
	}

	if config.Images.Path == "" {
		config.Images.Path = "images"
	}

	if config.Bot.Timeout == 0 {
		config.Bot.Timeout = 60
	}

	log.Printf("Configuration loaded successfully from %s", filename)
	return nil
}