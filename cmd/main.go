package main

import (
	finbot "bot"
	"bot/model"
	"context"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
	"time"
)

const tgbotapiKey = "6262794329:AAEi3ttuwyNueHDhS40WWrGhX5l8Xs0gAOg"

var mainMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🏠 Главная"),
		tgbotapi.NewKeyboardButton("🗒 Привязать номер телефона"),
	),
)

var courseMenu = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Добавить гостя"),
		tgbotapi.NewKeyboardButton("Мои гости"),
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

	db, dbErr := initDB(context.Background(), "user=postuser password=Aiparking_2022! dbname=postgres sslmode=disable host=195.49.212.96 port=5432")
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
					_, errCUD := CheckUserDb(*db, update.Message.Chat.ID)
					if errCUD != nil {
						if errCUD.Error() == "unf" {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для использования бота зарегистрируйтесь. \nПривяжите номер телефона")
							msg.ReplyMarkup = mainMenu
							bot.Send(msg)
							continue
						} else if errCUD.Error() == "dbe" {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для использования бота зарегистрируйтесь. \nПривяжите номер телефона")
							msg.ReplyMarkup = mainMenu
							bot.Send(msg)
							continue
						}
					}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите функцию")
					msg.ReplyMarkup = courseMenu
					bot.Send(msg)
				}
			} else {
				if update.Message.Text == mainMenu.Keyboard[0][0].Text {
					//Если пользователь нажимает главное меню, возвращаем главное меню
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Главное меню")
					msg.ReplyMarkup = mainMenu
					bot.Send(msg)

				} else if update.Message.Text == mainMenu.Keyboard[0][1].Text {
					//Если пользователь нажимает привязать телефон , переводим в стейт ожидания номера телефона и отправляем сообщение
					courseSignMap[update.Message.From.ID] = new(finbot.CourseSign)
					courseSignMap[update.Message.From.ID].State = finbot.StateTel
					msgConfig := tgbotapi.NewMessage(
						update.Message.Chat.ID,
						"Введите номет телфона в формате +77771234567:")
					msgConfig.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					bot.Send(msgConfig)

				} else if update.Message.Text == courseMenu.Keyboard[0][0].Text {

					fmt.Printf(
						"message: %s\n",
						update.Message.Text)

					//msgConfig := tgbotapi.NewMessage(
					//	update.Message.Chat.ID,
					//	"Выберите доступный объект")
					//msgConfig.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					//bot.Send(msgConfig)
					userId, errCUD := CheckUserDb(*db, update.Message.Chat.ID)
					if errCUD != nil {
						msgConfig := tgbotapi.NewMessage(
							update.Message.Chat.ID,
							"Пользователь не найден.")
						bot.Send(msgConfig)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пользователь не найден.")
						msg.ReplyMarkup = courseMenu
						bot.Send(msg)
						continue
					}
					buildings, errGUB := GetUserBuildings(*db, userId)
					if errGUB != nil {
						msgConfig := tgbotapi.NewMessage(
							update.Message.Chat.ID,
							"Нет дотсупных объектов.")
						bot.Send(msgConfig)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нет доступных объектов.")
						msg.ReplyMarkup = courseMenu
						bot.Send(msg)
						continue
					}
					if len(buildings) < 1 {
						msgConfig := tgbotapi.NewMessage(
							update.Message.Chat.ID,
							"Нет дотсупных объектов.")
						bot.Send(msgConfig)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нет доступных объектов.")
						msg.ReplyMarkup = courseMenu
						bot.Send(msg)
						continue
					}

					var buildingsButtons []tgbotapi.InlineKeyboardButton
					for _, b := range buildings {
						callback := fmt.Sprintf("building_%v", b.Id)
						buildingsButtons = append(buildingsButtons, tgbotapi.InlineKeyboardButton{
							Text:                         b.Name,
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
				} else {
					cs, ok := courseSignMap[update.Message.From.ID]
					if ok {
						if cs.State == finbot.StateTel {
							//Проверяем мапу, если ожидается номер телефона проверяем сообщение
							errCN := CheckNum(update.Message.Text)
							if errCN != nil {
								msgConfig := tgbotapi.NewMessage(
									update.Message.Chat.ID,
									"Номер телефона введен неверно.")
								bot.Send(msgConfig)
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для использования бота зарегистрируйтесь. \nПривяжите номер телефона")
								msg.ReplyMarkup = mainMenu
								bot.Send(msg)
								continue
							}
							errCND := CheckNumberDb(*db, update.Message.Text)
							if errCND != nil {
								msgConfig := tgbotapi.NewMessage(
									update.Message.Chat.ID,
									"Номер телефона не найден. \nВы не являетесь пользователем сервиса.")
								bot.Send(msgConfig)
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для использования бота зарегистрируйтесь. \nПривяжите номер телефона")
								msg.ReplyMarkup = mainMenu
								bot.Send(msg)
								continue
							}
							errATITU := AddTgIdToUser(*db, update.Message.Chat.ID, update.Message.Text)
							if errATITU != nil {
								msgConfig := tgbotapi.NewMessage(
									update.Message.Chat.ID,
									"Пользователь не зарегистрирован")
								bot.Send(msgConfig)
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для использования бота зарегистрируйтесь. \nПривяжите номер телефона")
								msg.ReplyMarkup = mainMenu
								bot.Send(msg)
								continue
							}
							//Меняем стейт на зарегистрирован и добавляем в базу
							UpdateUserState(*db, update.Message.Chat.ID, 1)
							cs.State = finbot.StateRegistered
							msgConfig := tgbotapi.NewMessage(
								update.Message.Chat.ID,
								"Пользователь зарегистрирован.")
							msgConfig.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
							bot.Send(msgConfig)
							msg2 := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите команду")
							msg2.ReplyMarkup = courseMenu
							bot.Send(msg2)
						} else if cs.State == finbot.StateRegistered {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите команду")
							msg.ReplyMarkup = courseMenu
							bot.Send(msg)
						} else if cs.State == finbot.StateBuilding {
							userId, errCUD := CheckUserDb(*db, update.Message.Chat.ID)
							if errCUD != nil {
								msgConfig := tgbotapi.NewMessage(
									update.Message.Chat.ID,
									"Пользователь не найден.")
								bot.Send(msgConfig)
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пользователь не найден.")
								msg.ReplyMarkup = courseMenu
								bot.Send(msg)
								continue
							}
							buildings, errGUB := GetUserBuildings(*db, userId)
							if errGUB != nil {
								msgConfig := tgbotapi.NewMessage(
									update.Message.Chat.ID,
									"Нет дотсупных объектов.")
								bot.Send(msgConfig)
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нет доступных объектов.")
								msg.ReplyMarkup = courseMenu
								bot.Send(msg)
								continue
							}
							if len(buildings) < 1 {
								msgConfig := tgbotapi.NewMessage(
									update.Message.Chat.ID,
									"Нет дотсупных объектов.")
								bot.Send(msgConfig)
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нет доступных объектов.")
								msg.ReplyMarkup = courseMenu
								bot.Send(msg)
								continue
							}
							var buildingMenu = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow())
							for _, b := range buildings {
								var but = tgbotapi.NewKeyboardButton(fmt.Sprintf("%v", b))
								buildingMenu.Keyboard = append(buildingMenu.Keyboard, tgbotapi.NewKeyboardButtonRow(but))
							}
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите объект ")
							msg.ReplyMarkup = buildingMenu
							bot.Send(msg)
						} else if cs.State == finbot.StateGuestAdd {

							id, errCUD := CheckUserDb(*db, update.Message.Chat.ID)
							if errCUD != nil {
								if errCUD.Error() == "unf" {
									msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для использования бота зарегистрируйтесь. \nПривяжите номер телефона")
									msg.ReplyMarkup = mainMenu
									bot.Send(msg)
									continue
								} else if errCUD.Error() == "dbe" {
									msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для использования бота зарегистрируйтесь. \nПривяжите номер телефона")
									msg.ReplyMarkup = mainMenu
									bot.Send(msg)
									continue
								}
							}
							userAccList, errGAUGAL := GetActiveUserGuestAccessList(*db, id)
							if errGAUGAL != nil {
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка получения гостевого списка: %s", errGAUGAL.Error()))
								msg.ReplyMarkup = courseMenu
								cs.State = finbot.StateBuilding
								bot.Send(msg)
								continue
							}
							if len(*userAccList) > 0 {
								for _, i := range *userAccList {
									if i.PlateNumber == update.Message.Text {
										msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Данный номер уже находится гостевом списке")
										msg.ReplyMarkup = courseMenu
										cs.State = finbot.StateBuilding
										bot.Send(msg)
									}
								}
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Превышен лимит гостевых доступов")
								msg.ReplyMarkup = courseMenu
								cs.State = finbot.StateBuilding
								bot.Send(msg)
								continue
							}
							errAUGA := AddUserGuestAccess(*db, &model.WhiteList{
								PlateNumber: update.Message.Text,
								BuildingID:  cs.Building,
								UserID:      int(id),
							})
							if errAUGA != nil {
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка добавления в гостевой список: %s", errAUGA.Error()))
								msg.ReplyMarkup = courseMenu
								cs.State = finbot.StateBuilding
								bot.Send(msg)
								continue
							}
							cs.State = finbot.StateBuilding
							msg4 := tgbotapi.NewMessage(int64(update.Message.Chat.ID), fmt.Sprintf("Гостевой доступ для %v предоставлен на 1 час.", update.Message.Text))
							msg4.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
							if _, errS := bot.Send(msg4); errS != nil {
								fmt.Printf(errS.Error())
							}
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите команду")
							msg.ReplyMarkup = courseMenu
							bot.Send(msg)
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
				if arr[0] == "building" {
					if arr[1] != "" {
						cs, ok := courseSignMap[update.CallbackQuery.From.ID]
						if ok {
							intVar, errAtoi := strconv.Atoi(arr[1])
							if errAtoi != nil {
								fmt.Printf("Error atoi")
							}
							cs.Building = intVar
							msg := tgbotapi.NewMessage(int64(update.CallbackQuery.From.ID), fmt.Sprintf("Введите гос номер гостя: "))
							msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
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

// Функция для проверки строки : является ли она номером телефона
func CheckNum(phone string) error {
	if len(phone) != 12 {
		return fmt.Errorf("Неправильно введен номер телефона")
	} else {
		phoneArr := strings.Split(phone, "")
		if phoneArr[0] != "+" {
			return fmt.Errorf("Неправильно введен номер телефона")
		} else {
			if phoneArr[1] != "7" {
				return fmt.Errorf("Неправильно введен номер телефона")
			} else {
				return nil
			}
		}
	}
}
