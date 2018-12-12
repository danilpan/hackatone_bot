package main

import (
	"fmt"

	"github.com/finalist736/finalistx-tg-bot"
	"github.com/finalist736/finalistx-tg-bot/post"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const tgbotapiKey = "long-long-tgbot-api-key"

var mainMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🏠 Главная"),
		tgbotapi.NewKeyboardButton("🗒 Запись"),
	),
)

var courseMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Golang"),
		tgbotapi.NewKeyboardButton("Intense golang"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("HighLoad"),
		tgbotapi.NewKeyboardButton("VueJS"),
	),
)

var courseSignMap map[int]*finbot.CourseSign

func init() {
	courseSignMap = make(map[int]*finbot.CourseSign)
}

func main() {
	var (
		bot        *tgbotapi.BotAPI
		err        error
		updChannel tgbotapi.UpdatesChannel
		update     tgbotapi.Update
		updConfig  tgbotapi.UpdateConfig
		botUser    tgbotapi.User
	)
	bot, err = tgbotapi.NewBotAPI(tgbotapiKey)
	if err != nil {
		panic("bot init error: " + err.Error())
	}

	botUser, err = bot.GetMe()
	if err != nil {
		panic("bot getme error: " + err.Error())
	}

	fmt.Printf("auth ok! bot is: %s\n", botUser.FirstName)

	updConfig.Timeout = 60
	updConfig.Limit = 1
	updConfig.Offset = 0

	updChannel, err = bot.GetUpdatesChan(updConfig)
	if err != nil {
		panic("update channel error: " + err.Error())
	}

	for {

		update = <-updChannel

		if update.Message != nil {

			if update.Message.IsCommand() {
				cmdText := update.Message.Command()
				if cmdText == "test" {
					msgConfig := tgbotapi.NewMessage(
						update.Message.Chat.ID,
						"test cmd")
					bot.Send(msgConfig)
				} else if cmdText == "menu" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Главное меню")
					msg.ReplyMarkup = mainMenu
					bot.Send(msg)
				}
			} else {

				if update.Message.Text == mainMenu.Keyboard[0][1].Text {

					courseSignMap[update.Message.From.ID] = new(finbot.CourseSign)
					courseSignMap[update.Message.From.ID].State = finbot.StateEmail

					fmt.Printf(
						"message: %s\n",
						update.Message.Text)

					msgConfig := tgbotapi.NewMessage(
						update.Message.Chat.ID,
						"Введите email:")
					msgConfig.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					bot.Send(msgConfig)
				} else {
					cs, ok := courseSignMap[update.Message.From.ID]
					if ok {
						if cs.State == finbot.StateEmail {
							cs.Email = update.Message.Text
							msgConfig := tgbotapi.NewMessage(
								update.Message.Chat.ID,
								"Введите телефон:")
							bot.Send(msgConfig)
							cs.State = 1
						} else if cs.State == finbot.StateTel {
							cs.Telephone = update.Message.Text
							cs.State = 2
							msgConfig := tgbotapi.NewMessage(
								update.Message.Chat.ID,
								"Введите course:")
							msgConfig.ReplyMarkup = courseMenu
							bot.Send(msgConfig)
						} else if cs.State == finbot.StateCourse {
							cs.Course = update.Message.Text
							msgConfig := tgbotapi.NewMessage(
								update.Message.Chat.ID,
								"ok!")
							msgConfig.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
							bot.Send(msgConfig)
							delete(courseSignMap, update.Message.From.ID)
							//  post to site!
							err = post.SendPost(cs)
							if err != nil {
								fmt.Printf("send post error: %v\n", err)
							}
						}
						fmt.Printf("state: %+v\n", cs)
					} else {
						// other messages
						msgConfig := tgbotapi.NewMessage(
							update.Message.Chat.ID,
							"ok")
						msgConfig.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
						bot.Send(msgConfig)
					}
				}
			}
		} else {
			fmt.Printf("not message: %+v\n", update)
		}
	}

	bot.StopReceivingUpdates()
}
