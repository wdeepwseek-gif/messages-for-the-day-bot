package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleMessage обрабатывает текстовые сообщения от пользователей
func handleMessage(message *tgbotapi.Message) {
	userID := message.Chat.ID
	text := message.Text

	log.Printf("Received message from %d: %s", userID, text)

	switch text {
	case "/start":
		sendWelcomeMessage(userID)
	case "/card":
		getDailyCard(userID)
	case "/random":
		getRandomCard(userID)
	case "/help":
		sendHelpMessage(userID)
	case "/about":
		sendAboutMessage(userID)
	default:
		if state, exists := userStates[userID]; exists && state != "" {
			handleUserState(userID, text, state)
		} else {
			sendMainMenu(userID)
		}
	}
}

// handleCallback обрабатывает нажатия inline кнопок
func handleCallback(callback *tgbotapi.CallbackQuery) {
	userID := callback.Message.Chat.ID
	data := callback.Data

	log.Printf("Received callback from %d: %s", userID, data)

	switch data {
	case "get_daily_card":
		getDailyCard(userID)
	case "get_random_card":
		getRandomCard(userID)
	case "main_menu":
		sendMainMenu(userID)
	case "about":
		sendAboutMessage(userID)
	case "help":
		sendHelpMessage(userID)
	}

	// Ответ на callback чтобы убрать "часики" в кнопке
	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	if _, err := bot.Send(callbackConfig); err != nil {
		log.Printf("Error answering callback: %v", err)
	}
}

// sendWelcomeMessage отправляет приветственное сообщение
func sendWelcomeMessage(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, config.Messages.Welcome)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createMainKeyboard()

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending welcome message: %v", err)
	}
}

// sendMainMenu отправляет главное меню
func sendMainMenu(chatID int64) {
	text := "✨ *Главное меню* ✨\n\nВыберите действие:"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createMainKeyboard()

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending main menu: %v", err)
	}
}

// createMainKeyboard создает клавиатуру для главного меню
func createMainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💫 Послание Дня", "get_daily_card"),
			tgbotapi.NewInlineKeyboardButtonData("✨ Случайное Послание", "get_random_card"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📖 О боте", "about"),
			tgbotapi.NewInlineKeyboardButtonData("🆘 Помощь", "help"),
		),
	)
}

// getDailyCard получает и отправляет послание дня
func getDailyCard(chatID int64) {
	if imagesCount == 0 {
		sendErrorMessage(chatID, "Послания не загружены. Попробуйте позже.")
		return
	}

	// Генерируем послание на основе даты (одинаковое для пользователя в течение дня)
	today := time.Now().Format("20060102")
	seed := today + strconv.FormatInt(chatID, 10)

	// Создаем свой источник случайных чисел для предсказуемости
	source := rand.NewSource(createSeed(seed))
	rng := rand.New(source)

	cardIndex := rng.Intn(imagesCount)
	cardNumber := availableNums[cardIndex]

	sendCardImage(chatID, cardNumber, "💫 *ВАШЕ ПОСЛАНИЕ ДНЯ* 💫", "Это послание будет с вами до конца дня. Откройте сердце и примите его энергию.")
}

// getRandomCard получает и отправляет случайное послание
func getRandomCard(chatID int64) {
	if imagesCount == 0 {
		sendErrorMessage(chatID, "Послания не загружены. Попробуйте позже.")
		return
	}

	// Полностью случайное послание
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	cardIndex := rng.Intn(imagesCount)
	cardNumber := availableNums[cardIndex]

	sendCardImage(chatID, cardNumber, "✨ *СЛУЧАЙНОЕ ПОСЛАНИЕ* ✨", "Это послание пришло к вам именно сейчас. Примите его энергию и позвольте ей наполнить вас.")
}

// createSeed создает числовой seed из строки
func createSeed(seed string) int64 {
	var hash int64
	for _, char := range seed {
		hash = hash*31 + int64(char)
	}
	return hash
}

// sendCardImage отправляет картинку с посланием
func sendCardImage(chatID int64, cardNumber int, title string, subtitle string) {
	imagePath := fmt.Sprintf("%s/%d.jpg", imagesPath, cardNumber)
	
	// Проверяем существование файла
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		sendErrorMessage(chatID, fmt.Sprintf("Послание %d не найдено.", cardNumber))
		return
	}

	// Отправляем фото
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imagePath))
	photo.Caption = fmt.Sprintf(`%s

%s

🌟 *Пусть энергия послания наполнит вас* 🌟`, title, subtitle)
	photo.ParseMode = "Markdown"
	photo.ReplyMarkup = createCardKeyboard()

	if _, err := bot.Send(photo); err != nil {
		log.Printf("Error sending card image: %v", err)
	}
}

// createCardKeyboard создает клавиатуру для сообщения с посланием
func createCardKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💫 Послание дня", "get_daily_card"),
			tgbotapi.NewInlineKeyboardButtonData("✨ Случайное", "get_random_card"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)
}

// sendAboutMessage отправляет информацию о боте
func sendAboutMessage(chatID int64) {
	aboutText := fmt.Sprintf("%s\n\n📊 *Статистика коллекции:* \n• Всего посланий: %d",
		config.Messages.About, imagesCount)

	msg := tgbotapi.NewMessage(chatID, aboutText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending about message: %v", err)
	}
}

// sendHelpMessage отправляет справку по использованию бота
func sendHelpMessage(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, config.Messages.Help)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "main_menu"),
		),
	)

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending help message: %v", err)
	}
}

// sendErrorMessage отправляет сообщение об ошибке
func sendErrorMessage(chatID int64, errorMessage string) {
	msg := tgbotapi.NewMessage(chatID, "❌ *Ошибка:* "+errorMessage)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createMainKeyboard()

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending error message: %v", err)
	}
}

// handleUserState обрабатывает сообщения в зависимости от состояния пользователя
func handleUserState(chatID int64, text string, state string) {
	// Здесь можно добавить логику для разных состояний
	// Например, если пользователь в процессе какого-то диалога

	// Пока просто очищаем состояние и возвращаем в главное меню
	delete(userStates, chatID)
	sendMainMenu(chatID)
}

// sendDailyReminder отправляет напоминание о послании дня (можно использовать с cron)
func sendDailyReminder(chatID int64) {
	reminderText := `🌅 *Доброе утро!*

Не забудьте получить ваше послание дня для сегодняшнего вдохновения и поддержки.

Пусть энергия послания наполнит ваш день! ✨`

	msg := tgbotapi.NewMessage(chatID, reminderText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createMainKeyboard()

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending daily reminder: %v", err)
	}
}