package main

import (
	finbot "bot"
	"context"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
	"time"
)

const tgbotapiKey = "6599935805:AAFGjCj-2jVrw7_EP-xCDlfsT0A3ID0hRhY"

var mainMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🏠 Главная"),
		tgbotapi.NewKeyboardButton("🗒 Привязать номер телефона"),
	),
)

var signUpMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🗒 Регистрация"),
	),
)

var courseMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Список бронирований"),
	),
)

var stateMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Подтвердить"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Отменить"),
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
	loc, err := time.LoadLocation("Asia/Almaty")
	// handle err
	time.Local = loc
	botUser, err = bot.GetMe()
	if err != nil {
		panic("bot getme error: " + err.Error())
	}

	db, dbErr := initDB(context.Background(), "user=wikizryuatvdce password=b9cf7e3712cf581144fe69a31844d0628e0bb4abec2143759bdbcea5f02b5e73 dbname=d7pa2050gj777r sslmode=require host=ec2-54-73-22-169.eu-west-1.compute.amazonaws.com port=5432")
	if dbErr != nil {
		panic("db error: " + dbErr.Error())
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
			//Если команда, пока обрабатываем только команду /start
			if update.Message.IsCommand() {
				cmdText := update.Message.Command()
				if cmdText == "start" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите функцию")
					msg.ReplyMarkup = courseMenu
					bot.Send(msg)
				}
			} else {
				if update.Message.Text == courseMenu.Keyboard[0][0].Text {
					reservs, _ := GetReservations(*db)
					var buildingsButtons []tgbotapi.InlineKeyboardButton
					for _, b := range reservs {
						callback := fmt.Sprintf("reserv_%v", b.Id)
						buildingsButtons = append(buildingsButtons, tgbotapi.InlineKeyboardButton{
							Text:                         fmt.Sprintf("%v", b.Id),
							URL:                          nil,
							CallbackData:                 &callback,
							SwitchInlineQuery:            nil,
							SwitchInlineQueryCurrentChat: nil,
							CallbackGame:                 nil,
							Pay:                          false,
						})
					}
					courseSignMap[update.Message.From.ID] = new(finbot.CourseSign)
					courseSignMap[update.Message.From.ID].State = finbot.StateTel
					buildingMenu := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buildingsButtons...))
					msg4 := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите объект.")
					msg4.ReplyMarkup = buildingMenu
					if _, errS := bot.Send(msg4); errS != nil {
						fmt.Printf(errS.Error())
					}
				} else if update.Message.Text == stateMenu.Keyboard[0][0].Text {
					cs, ok := courseSignMap[update.Message.From.ID]
					if ok {
						errPUGA := Accept(*db, cs.NumberId)
						if errPUGA != nil {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка продления гостевого доступа"))
							msg.ReplyMarkup = stateMenu
							cs.State = finbot.StateNumberChangeState
							bot.Send(msg)
							continue
						}
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Бронь подтверждена"))
						msg.ReplyMarkup = courseMenu
						cs.State = finbot.StateRegistered
						bot.Send(msg)
					}
				} else if update.Message.Text == stateMenu.Keyboard[1][0].Text {
					cs, ok := courseSignMap[update.Message.From.ID]
					if ok {
						errPUGA := Cancel(*db, cs.NumberId)
						if errPUGA != nil {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка продления гостевого доступа"))
							msg.ReplyMarkup = stateMenu
							cs.State = finbot.StateNumberChangeState
							bot.Send(msg)
							continue
						}
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Бронирование отменено"))
						msg.ReplyMarkup = courseMenu
						cs.State = finbot.StateRegistered
						bot.Send(msg)
					}
				} else {
					cs, ok := courseSignMap[update.Message.From.ID]
					if ok {
						if cs.State == finbot.StateRegistered {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите команду")
							msg.ReplyMarkup = courseMenu
							bot.Send(msg)
						} else if cs.State == finbot.StateRegistrationLastname {

						}

					}
					fmt.Printf("state: %+v\n", cs)

				}
			}
		} else if update.CallbackQuery != nil {
			if update.CallbackQuery.Data != "" {
				arr := strings.Split(update.CallbackQuery.Data, "_")
				if len(arr) != 2 {
					msg4 := tgbotapi.NewMessage(int64(update.CallbackQuery.From.ID), "Выберите объект.")
					msg4.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					if _, errS := bot.Send(msg4); errS != nil {
						fmt.Printf(errS.Error())
					}
				}
				if arr[0] == "reserv" {
					if arr[1] != "" {
						cs, ok := courseSignMap[update.CallbackQuery.From.ID]
						if ok {
							intVar, errAtoi := strconv.Atoi(arr[1])
							if errAtoi != nil {
								fmt.Printf("Error atoi")
							}
							cs.Building = intVar
							msg := tgbotapi.NewMessage(
								int64(update.CallbackQuery.From.ID),
								fmt.Sprintf("Выберите действие"))
							msg.ReplyMarkup = stateMenu
							cs.State = finbot.StateGuestAdd
							bot.Send(msg)

							continue
						}
					}
				}

			} else {
				continue
			}
		} else { //Если не сообщение и не колбэкквери
			fmt.Printf("not message: %+v\n", update)
		}
	}

	bot.StopReceivingUpdates()
}
