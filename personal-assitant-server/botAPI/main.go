package main

import (
	tb "gopkg.in/telebot.v3"
	"log"
	"personal-assitant-project/personal-assitant-server/botAPI/config"
	"time"
)

var (
	// Universal markup builders.
	menu     = &tb.ReplyMarkup{ResizeKeyboard: true}
	selector = &tb.ReplyMarkup{}

	btnHelp     = menu.Text("ℹ Help")
	btnSettings = menu.Text("⚙ Settings")
	//TODO Для каждого дела выводить пометку сделано оно или нет, а также добавить возможность завершить дело из телеграмма
	btnBusinessForToday    = menu.Text("Дела на сегодня")
	btnNonFinishedBusiness = menu.Text("Незаврешенные дела")
	btnPrev                = selector.Data("⬅", "prev", "")
)

func main() {

	menu.Reply(
		menu.Row(btnHelp),
		menu.Row(btnSettings, btnBusinessForToday),
	)
	selector.Inline(
		selector.Row(btnPrev),
	)
	b, err := tb.NewBot(tb.Settings{
		Token:  config.TOKEN,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/start", func(c tb.Context) error {
		return c.Send("Здравствуйте, я рад привествовать вас. Пожалуйста, введите /start для начала работы😏", menu)
	})

	b.Handle(&btnHelp, func(c tb.Context) error {
		return c.Send("В работе😏", selector)
	})
	b.Handle(&btnBusinessForToday, getBusinessForTodayFromDB)

	//TODO найти как отправлять без текста
	//b.Handle(&btnPrev, func(c tb.Context) error {
	//	return c.Send("t", menu)
	//})

	b.Start()
}
func getBusinessForTodayFromDB(c tb.Context) error {
	return c.Send("В работе😏", selector)
}
