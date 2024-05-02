package main

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	tb "gopkg.in/telebot.v3"
	"log"
	"personal-assitant-project/config"
	botConfig "personal-assitant-project/personal-assitant-server/botAPI/config"
	"personal-assitant-project/personal-assitant-server/storage/elastic"
	"personal-assitant-project/personal-assitant-server/storage/postgre"
	"strconv"
	"strings"
	"time"
)

var ctx = context.Background()

var redisClient *redis.Client

func init() {
	redisConfig := config.LoadRedisConfig()
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisConfig.Addr,
		DB:   redisConfig.DBTelegramEvents,
	})
}

// TODO добавить в конце дня похвалу, вы молодец! вы выполнили столько-то планов
var (
	menu                   = &tb.ReplyMarkup{}
	selector               = &tb.ReplyMarkup{}
	tasks                  = &tb.ReplyMarkup{}
	btnHelp                = menu.Data("ℹ Help", "help")
	btnSettings            = menu.Data("⚙ Settings", "settings")
	btnBusinessForToday    = menu.Data("Дела на сегодня", "today")
	btnNonFinishedBusiness = menu.Data("Незавершенные дела на сегодня", "unfinished_today")
	btnBusinessForTomorrow = menu.Data("Дела на завтра", "tomorrow")
	btnAddNewTask          = menu.Data("Добавить новое событие", "addTask", "")
	btnPrev                = selector.Data("⬅", "prev", "")
	btnFinishTask          = tasks.Data("Завершить событие", "finish")
	btnUnFinishTask        = tasks.Data("Вернуть событие", "unFinish", "")
)

func main() {
	menu.Inline(
		menu.Row(btnHelp, btnSettings),
		menu.Row(btnBusinessForToday, btnBusinessForTomorrow),
		menu.Row(btnNonFinishedBusiness),
	)
	selector.Inline(
		selector.Row(btnPrev),
	)
	tasks.Inline(
		tasks.Row(btnPrev),
		tasks.Row(btnFinishTask),
	)
	b, err := tb.NewBot(tb.Settings{
		Token:  botConfig.TOKEN,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/start", func(c tb.Context) error {
		return c.Send("Здравствуйте, я рад приветствовать вас.😏", menu)
	})
	b.Handle(&btnHelp, func(c tb.Context) error {
		return c.Send("В работе😏", selector)
	})
	b.Handle(&btnSettings, func(c tb.Context) error {
		return c.Send("В работе😏", selector)
	})
	b.Handle(&btnBusinessForToday, getBusinessForTodayFromDB)
	b.Handle(&btnNonFinishedBusiness, getBusinessNotFinishedForTodayFromDB)
	b.Handle(&btnBusinessForTomorrow, getBusinessForTomorrowFromDB)
	b.Handle(&btnFinishTask, finishTask)
	b.Handle(&btnUnFinishTask, unFinishTask)
	b.Handle(&btnPrev, func(c tb.Context) error {
		return c.Send("Возвращаемся назад", menu)
	})
	go sendEventReminders(b)
	go sendNotifyEveryMoring(b)
	go sendEveryEveningForUsers(b)
	b.Start()
}
func sendNotifyEveryMoring(b *tb.Bot) {
	allUsers, err := postgre.GetAllTGUsersID()
	if err != nil {
		elastic.LogToElasticsearch(fmt.Sprintf("Error getting all users:", err))
		log.Println("Error getting all users:", err)
		return
	}
	for _, user := range allUsers {
		settings, err := postgre.GetUserSettings(user.UserID)
		if err != nil {
			elastic.LogToElasticsearch(fmt.Sprintf("Error getting settings:", err))
			log.Println("Error getting settings:", err)
			return
		}
		now := time.Now()
		targetTime := time.Date(now.Year(), now.Month(), now.Day(), int(settings.NotifyMorningHours), 0, 0, 0, now.Location())
		if now.After(targetTime) {
			targetTime = targetTime.Add(24 * time.Hour)
		}
		time.Sleep(targetTime.Sub(now))
		sendMorningTaskListToAllUsers(b, user)
	}
}
func sendEveryEveningForUsers(b *tb.Bot) {
	allUsers, err := postgre.GetAllTGUsersID()
	if err != nil {
		elastic.LogToElasticsearch(fmt.Sprintf("Error getting all users:", err))
		log.Println("Error getting all users:", err)
		return
	}
	for _, user := range allUsers {
		settings, err := postgre.GetUserSettings(user.UserID)
		if err != nil {
			elastic.LogToElasticsearch(fmt.Sprintf("Error getting settings:", err))
			log.Println("Error getting settings:", err)
			return
		}
		now := time.Now()
		targetTime := time.Date(now.Year(), now.Month(), now.Day(), int(settings.NotifyEveryEveningHours), 0, 0, 0, now.Location())
		if now.After(targetTime) {
			targetTime = targetTime.Add(24 * time.Hour)
		}
		time.Sleep(targetTime.Sub(now))
		allUsers, err = postgre.GetAllTGUsersID()
		if err != nil {
			elastic.LogToElasticsearch(fmt.Sprintf("Error getting all users:", err))
			log.Println("Error getting all users:", err)
			return
		}
		sendEveningTaskListToAllUsers(b, user)
	}
}
func sendEventReminders(b *tb.Bot) {
	ticker := time.NewTicker(1 * time.Minute) // Проверяем события каждую минуту
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		allUsers, err := postgre.GetAllTGUsersID()
		if err != nil {
			elastic.LogToElasticsearch(fmt.Sprintf("Error getting all users:", err))
			log.Println("Error getting all users:", err)
			continue
		}

		for _, user := range allUsers {
			settings, err := postgre.GetUserSettings(user.UserID)
			if err != nil {
				elastic.LogToElasticsearch(fmt.Sprintf("Error getting settings:", err))
				log.Println("Error getting settings:", err)
				return
			}
			events, err := postgre.FindTasksByUser(user.UserID)
			if err != nil {
				elastic.LogToElasticsearch(fmt.Sprintf("Error getting events for user:", err))
				log.Println("Error getting events for user:", err)
				continue
			}

			for _, event := range events {
				if !event.IsFinished {
					reminderTime := event.PlannedDate.Add(-time.Duration(settings.NotifyBeforeEventHours) * time.Minute)
					if now.After(reminderTime) && now.Before(event.PlannedDate) {
						sendReminderForEvent(b, user.TGUserID, event)
					}
				}
			}
		}
	}
}
func sendReminderForEvent(b *tb.Bot, userID int64, event postgre.EventData) {
	formattedTime := event.PlannedDate.Format("02.01.2006 15:04")
	message := fmt.Sprintf("Напоминание: %s\nВремя: %s", event.Title, formattedTime)
	_, err := b.Send(&tb.User{ID: userID}, message)
	if err != nil {
		log.Printf("Error sending reminder to user %d: %v\n", userID, err)
	}
}

func finishTask(c tb.Context) error {
	taskID, err := strconv.Atoi(c.Callback().Data)
	if err != nil {
		return c.Reply("Ошибка изменения статуса!", selector)
	}
	redisKey := fmt.Sprintf("user:%d", c.Sender().ID)
	postgre.FinishTask(taskID, true)
	redisClient.Set(ctx, fmt.Sprintf("%s:last_update", redisKey), time.Now().Add(time.Duration(-5)*time.Minute).Format(time.RFC3339), 0)
	return c.Reply("Событие успешно завершено!", selector)
}
func unFinishTask(c tb.Context) error {
	taskID, err := strconv.Atoi(c.Callback().Data)
	if err != nil {

	}
	redisKey := fmt.Sprintf("user:%d", c.Sender().ID)
	postgre.FinishTask(taskID, false)
	redisClient.Set(ctx, fmt.Sprintf("%s:last_update", redisKey), time.Now().Add(time.Duration(-5)*time.Minute).Format(time.RFC3339), 0)
	return c.Reply("Событие теперь имеет статус не завершено!", selector)
}
func sendEventsToUser(filteredEvents []postgre.EventData, c tb.Context, includeFinished bool) error {
	if len(filteredEvents) == 0 {
		return c.Send("Никаких незавершенных дел", selector)
	}
	for _, event := range filteredEvents {
		formattedTime := event.PlannedDate.Format("02.01.2006 15:04")
		formattedStartDate := event.StartDate.Format("02.01.2006 15:04")
		attachment := ""
		if string(event.Attachment) != "" {
			attachment = "Attachment: " + string(event.Attachment)
		}
		statusEmoji := "❗️"
		keyboard := tasks
		btnFinishTask = keyboard.Data("Завершить событие", "finish", strconv.Itoa(event.TaskID))
		keyboard.Inline(keyboard.Row(btnPrev), keyboard.Row(btnFinishTask))
		if event.IsFinished {
			statusEmoji = "✅"
			keyboard = &tb.ReplyMarkup{}
			btnUnFinishTask = keyboard.Data("Вернуть событие", "unFinish", strconv.Itoa(event.TaskID))
			keyboard.Inline(keyboard.Row(btnPrev), keyboard.Row(btnUnFinishTask))
		}
		message := fmt.Sprintf(
			"Заголовок: %s %s\n"+
				"Описание: %s\n"+
				"Планируемое время: %s\n"+
				"Дата добавления: %s\n"+
				"%s", event.Title, statusEmoji, event.Description, formattedTime, formattedStartDate, attachment)
		if err := c.Send(message, keyboard); err != nil {
			// Обработка ошибок отправки
			return err

		}
	}

	return nil
}
func sendEventsToAllUsers(filteredEvents []postgre.EventData, tgUserID int64, includeFinished bool, b *tb.Bot, messageText string) error {
	if len(filteredEvents) == 0 {
		return nil
	}
	for _, event := range filteredEvents {
		formattedTime := event.PlannedDate.Format("02.01.2006 15:04")
		formattedStartDate := event.StartDate.Format("02.01.2006 15:04")
		attachment := ""
		if string(event.Attachment) != "" {
			attachment = "Attachment: " + string(event.Attachment)
		}
		statusEmoji := "❗️"
		keyboard := tasks
		btnFinishTask = keyboard.Data("Завершить событие", "finish", strconv.Itoa(event.TaskID))
		keyboard.Inline(keyboard.Row(btnPrev), keyboard.Row(btnFinishTask))
		if event.IsFinished {
			statusEmoji = "✅"
			keyboard = &tb.ReplyMarkup{}
			btnUnFinishTask = keyboard.Data("Вернуть событие", "unFinish", strconv.Itoa(event.TaskID))
			keyboard.Inline(keyboard.Row(btnPrev), keyboard.Row(btnUnFinishTask))
		}
		message := fmt.Sprintf(
			messageText+
				"Заголовок: %s %s\n"+
				"Описание: %s\n"+
				"Планируемое время: %s\n"+
				"Дата добавления: %s\n"+
				"%s", event.Title, statusEmoji, event.Description, formattedTime, formattedStartDate, attachment)
		_, err := b.Send(&tb.User{ID: tgUserID}, message, keyboard)
		if err != nil {
			log.Printf("Error sending evening message to user %d: %v\n", tgUserID, err)
		}
	}

	return nil
}
func getBusinessForTomorrowFromDB(c tb.Context) error {
	events, err := getEventsFromCache(c)
	if err != nil {
		return err
	}

	// Отфильтровать события по PlannedDate равному сегодняшней дате
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	var filteredEvents []postgre.EventData
	for _, event := range events {
		eventDate := event.PlannedDate.Format("2006-01-02")
		if eventDate == tomorrow {
			filteredEvents = append(filteredEvents, event)
		}
	}
	return sendEventsToUser(filteredEvents, c, true)
}

//TODO добавить новую таблицу в postgre tg_messages, где будет id из tg bot id сообщения и таска
//TODO либо как-то добавить в eventData task_id и при нажатии на завершить вытягивать из контекста этот task id

func getBusinessNotFinishedForTodayFromDB(c tb.Context) error {
	events, err := getEventsFromCache(c)
	if err != nil {
		return err
	}

	// Отфильтровать события по PlannedDate равному сегодняшней дате
	today := time.Now().Format("2006-01-02")
	var filteredEvents []postgre.EventData
	for _, event := range events {
		eventDate := event.PlannedDate.Format("2006-01-02")
		if eventDate == today && !event.IsFinished {
			filteredEvents = append(filteredEvents, event)
		}
	}
	return sendEventsToUser(filteredEvents, c, false)
}

// TODO Есть два варианта: либо для каждой кнопки сделать отдельный запрос в postgre, либо один раз вывести все события  и раз в 5 минут обновлять данные в redis
func getBusinessForTodayFromDB(c tb.Context) error {
	events, err := getEventsFromCache(c)
	if err != nil {
		// Обработка ошибок
		return err
	}

	// Отфильтровать события по PlannedDate равному сегодняшней дате
	today := time.Now().Format("2006-01-02")
	var filteredEvents []postgre.EventData
	for _, event := range events {
		eventDate := event.PlannedDate.Format("2006-01-02")
		if eventDate == today {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return sendEventsToUser(filteredEvents, c, true)
}
func getEventsFromCache(c tb.Context) ([]postgre.EventData, error) {
	var cachedEvents []postgre.EventData
	// Создаем уникальный ключ для пользователя в Redis
	redisKey := fmt.Sprintf("user:%d", c.Sender().ID)

	// Проверяем, есть ли дата последнего обновления в Redis
	lastUpdate, err := redisClient.Get(ctx, redisKey+":last_update").Result()
	if err != nil || lastUpdate == "" {
		// Если дата отсутствует, либо произошла ошибка, обновляем кэш и устанавливаем новую дату
		cachedEvents, err = updateCache(c, redisKey)
		if err != nil {
			// Обработка ошибок
			return nil, err
		}
		return cachedEvents, nil
	}

	// Парсим дату последнего обновления
	lastUpdateDate, err := time.Parse(time.RFC3339, lastUpdate)
	if err != nil {
		// Обработка ошибок
		return nil, err
	}

	// Проверяем, прошло ли более 1 минут с последнего обновления
	if time.Since(lastUpdateDate) > 1*time.Minute {
		// Если прошло более 1 минут, обновляем кэш и устанавливаем новую дату
		cachedEvents, err = updateCache(c, redisKey)
		if err != nil {
			// Обработка ошибок
			return nil, err
		}
	}
	cachedData, err := redisClient.Get(ctx, redisKey).Result()
	if err != nil {
		// Обработка ошибок
		return nil, err
	}
	err = json.Unmarshal([]byte(cachedData), &cachedEvents)
	if err != nil {
		// Обработка ошибок
		return nil, err
	}

	// Отправляем кэшированные данные пользователю
	return cachedEvents, nil
}

// Функция обновления кэша
func updateCache(c tb.Context, redisKey string) ([]postgre.EventData, error) {
	// Получаем данные из Postgres
	userID, err := postgre.GetUserIDFromTasks(c.Sender().ID)
	if err != nil {
		// Обработка ошибок
		return nil, err
	}
	events, err := postgre.FindTasksByUser(userID)
	if err != nil {
		// Обработка ошибок
		return nil, err
	}

	// Преобразуем события в JSON
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		// Обработка ошибок
		return nil, err
	}

	// Сохраняем данные в Redis
	redisClient.Set(ctx, redisKey, eventsJSON, 0) // 0 означает, что запись не истекает
	redisClient.Set(ctx, fmt.Sprintf("%s:last_update", redisKey), time.Now().Format(time.RFC3339), 0)

	return events, nil
}

func sendMorningTaskListToAllUsers(b *tb.Bot, user postgre.AllTBotUsers) {
	today := time.Now().Format("2006-01-02")
	taskListMessage, err := postgre.FindTasksByUser(user.UserID)
	var filteredEvents []postgre.EventData
	for _, event := range taskListMessage {
		eventDate := event.PlannedDate.Format("2006-01-02")
		if eventDate == today {
			filteredEvents = append(filteredEvents, event)
		}
	}
	if err != nil {
		log.Printf("Error getting morning task list for user %d: %v\n", user.UserID, err)
	}
	messageText := "Доброе утро и так дела на сегодня:"
	sendEventsToAllUsers(filteredEvents, user.TGUserID, true, b, messageText)

}

func sendEveningTaskListToAllUsers(b *tb.Bot, user postgre.AllTBotUsers) {
	today := time.Now().Format("2006-01-02")
	taskListMessage, err := postgre.FindTasksByUser(user.UserID)
	var filteredEvents []postgre.EventData
	for _, event := range taskListMessage {
		eventDate := event.PlannedDate.Format("2006-01-02")
		if eventDate == today || event.IsFinished {
			filteredEvents = append(filteredEvents, event)
		}
	}
	if err != nil {
		log.Printf("Error getting morning task list for user %d: %v\n", user.UserID, err)
	}
	messageText := "Добрый вечер и так дела, которые вы успешно завершили за сегодня:"
	sendEveningEventsToAllUsers(filteredEvents, user.TGUserID, true, b, messageText)

}
func sendEveningEventsToAllUsers(filteredEvents []postgre.EventData, tgUserID int64, includeFinished bool, b *tb.Bot, messageText string) error {
	if len(filteredEvents) == 0 {
		return nil
	}
	var result []string
	for _, event := range filteredEvents {
		if event.FinishedDate.Format("02.01.2006") == time.Now().Format("02.01.2006") {
			formattedTime := event.PlannedDate.Format("02.01.2006 15:04")
			formattedStartDate := event.StartDate.Format("02.01.2006 15:04")
			attachment := ""
			if string(event.Attachment) != "" {
				attachment = "Attachment: " + string(event.Attachment)
			}
			statusEmoji := "❗️"
			if event.IsFinished {
				statusEmoji = "✅"
				message := fmt.Sprintf(
					messageText+
						"Заголовок: %s %s\n"+
						"Описание: %s\n"+
						"Планируемое время: %s\n"+
						"Дата добавления: %s\n"+
						"%s", event.Title, statusEmoji, event.Description, formattedTime, formattedStartDate, attachment)
				result = append(result, message)
			}
		} else {
			continue
		}

	}
	resultMessage := strings.Join(result, "\n")
	_, err := b.Send(&tb.User{ID: tgUserID}, "За сегодня вы выполнили: "+strconv.Itoa(len(result))+resultMessage)
	if err != nil {
		log.Printf("Error sending evening message to user %d: %v\n", tgUserID, err)
	}
	return nil
}
