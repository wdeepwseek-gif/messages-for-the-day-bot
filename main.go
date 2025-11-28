package main

import (
	"log"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Глобальные переменные
var (
	bot           *tgbotapi.BotAPI
	imagesCount   int
	availableNums []int
	imagesPath    string
	userStates    = make(map[int64]string)
	config        Config
)

func main() {
	// Загрузка конфигурации
	err := LoadConfig("config.yaml")
	if err != nil {
		log.Panicf("Error loading config: %v", err)
	}

	// Загрузка информации о картинках
	imagesPath = config.Images.Path
	loadImages()

	// Инициализация бота
	bot, err = tgbotapi.NewBotAPI(config.Bot.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = config.Bot.Debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Установка команд бота для меню
	setBotCommands()

	// Настройка обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = config.Bot.Timeout

	updates := bot.GetUpdatesChan(u)

	// Обработка сообщений
	for update := range updates {
		if update.Message != nil {
			go handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			go handleCallback(update.CallbackQuery)
		}
	}
}

// loadImages загружает список доступных картинок из папки images
func loadImages() {
	files, err := filepath.Glob(filepath.Join(imagesPath, "*.jpg"))
	if err != nil {
		log.Printf("Error reading images directory: %v", err)
		return
	}

	if len(files) == 0 {
		log.Printf("Warning: No images found in %s directory", imagesPath)
		return
	}

	// Извлекаем номера из имен файлов
	availableNums = make([]int, 0, len(files))
	for _, file := range files {
		baseName := filepath.Base(file)
		// Убираем расширение .jpg
		nameWithoutExt := strings.TrimSuffix(baseName, ".jpg")
		// Пытаемся преобразовать в число
		if num, err := strconv.Atoi(nameWithoutExt); err == nil {
			availableNums = append(availableNums, num)
		}
	}

	// Сортируем номера для удобства
	sort.Ints(availableNums)
	imagesCount = len(availableNums)

	if imagesCount == 0 {
		log.Printf("Warning: No valid numbered images found in %s directory", imagesPath)
		return
	}

	if len(availableNums) > 0 {
		log.Printf("Loaded %d images from %s directory (numbers: %d-%d)", 
			imagesCount, imagesPath, availableNums[0], availableNums[len(availableNums)-1])
	} else {
		log.Printf("Loaded %d images from %s directory", imagesCount, imagesPath)
	}
}

// setBotCommands устанавливает команды бота для отображения в меню
func setBotCommands() {
	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "Начать работу с ботом",
		},
		{
			Command:     "card",
			Description: "Получить послание дня",
		},
		{
			Command:     "random",
			Description: "Случайное послание",
		},
		{
			Command:     "help",
			Description: "Помощь и справка",
		},
		{
			Command:     "about",
			Description: "О боте",
		},
	}

	cmdConfig := tgbotapi.NewSetMyCommands(commands...)
	if _, err := bot.Request(cmdConfig); err != nil {
		log.Printf("Error setting bot commands: %v", err)
	} else {
		log.Printf("Bot commands set successfully")
	}
}