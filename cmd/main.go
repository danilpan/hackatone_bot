package main

import (
	finbot "bot"
	"bot/model"
	"context"
	"database/sql"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
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

			if update.Message.IsCommand() {
				cmdText := update.Message.Command()
				if cmdText == "start" {
					_, errCUD := CheckUserDb(*db, update.Message.Chat.ID)
					if errCUD != nil {
						if errCUD.Error() == "unf" {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для использования бота зарегистрируйтесь. \nПривяжите номер телефона")
							msg.ReplyMarkup = mainMenu
							bot.Send(msg)
						} else if errCUD.Error() == "dbe" {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для использования бота зарегистрируйтесь. \nПривяжите номер телефона")
							msg.ReplyMarkup = mainMenu
							bot.Send(msg)
						}
					}
					courseSignMap[update.Message.From.ID] = new(finbot.CourseSign)
					courseSignMap[update.Message.From.ID].State = finbot.StateRegistered
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите функцию")
					msg.ReplyMarkup = courseMenu
					bot.Send(msg)
				}

			} else {

				if update.Message.Text == mainMenu.Keyboard[0][0].Text {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Главное меню")
					msg.ReplyMarkup = mainMenu
					bot.Send(msg)

				} else if update.Message.Text == mainMenu.Keyboard[0][1].Text {

					courseSignMap[update.Message.From.ID] = new(finbot.CourseSign)
					courseSignMap[update.Message.From.ID].State = finbot.StateTel

					fmt.Printf(
						"message: %s\n",
						update.Message.Text)

					msgConfig := tgbotapi.NewMessage(
						update.Message.Chat.ID,
						"Введите номет телфона в формате +77771234567:")
					msgConfig.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					bot.Send(msgConfig)

				} else if update.Message.Text == courseMenu.Keyboard[0][0].Text {

					courseSignMap[update.Message.From.ID] = new(finbot.CourseSign)
					courseSignMap[update.Message.From.ID].State = finbot.StateGuestAdd

					fmt.Printf(
						"message: %s\n",
						update.Message.Text)

					msgConfig := tgbotapi.NewMessage(
						update.Message.Chat.ID,
						"Введите номет автомобиля в формате 001AAA01")
					msgConfig.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					bot.Send(msgConfig)

				} else {
					cs, ok := courseSignMap[update.Message.From.ID]
					if ok {
						if cs.State == finbot.StateTel {
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
							cs.State = finbot.StateRegistered
							msgConfig := tgbotapi.NewMessage(
								update.Message.Chat.ID,
								"Пользователь зарегистрирован.")
							msgConfig.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
							bot.Send(msgConfig)
						} else if cs.State == finbot.StateRegistered {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите команду")
							msg.ReplyMarkup = courseMenu
							bot.Send(msg)
						} else if cs.State == finbot.StateGuestAdd {

							id, errCUD := CheckUserDb(*db, update.Message.Chat.ID)
							if errCUD != nil {
								if errCUD.Error() == "unf" {
									msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для использования бота зарегистрируйтесь. \nПривяжите номер телефона")
									msg.ReplyMarkup = mainMenu
									bot.Send(msg)
								} else if errCUD.Error() == "dbe" {
									msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Для использования бота зарегистрируйтесь. \nПривяжите номер телефона")
									msg.ReplyMarkup = mainMenu
									bot.Send(msg)
								}
							}
							userAccList, errGAUGAL := GetActiveUserGuestAccessList(*db, id)
							if errGAUGAL != nil {
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка получения гостевого списка: %s", errGAUGAL.Error()))
								msg.ReplyMarkup = mainMenu
								bot.Send(msg)
							}
							if len(*userAccList) > 0 {
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Превышен лимит гостевых доступов")
								msg.ReplyMarkup = mainMenu
								bot.Send(msg)
							}
							errAUGA := AddUserGuestAccess(*db, &model.WhiteList{
								PlateNumber: update.Message.Text,
								BuildingID:  0,
								UserID:      int(update.Message.Chat.ID),
							})
							if errAUGA != nil {
								msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка получения гостевого списка: %s", errAUGA.Error()))
								msg.ReplyMarkup = courseMenu
								bot.Send(msg)
							}
						}

					}
					fmt.Printf("state: %+v\n", cs)

				}
			}
		} else {
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

func CheckNumberDb(db sqlx.DB, phone string) error {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE phone_number=$1`
	err := db.Get(&count, query, phone)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	} else {
		return fmt.Errorf("user with such num not found")
	}
}

func AddTgIdToUser(db sqlx.DB, id int64, phone string) error {
	query := `UPDATE users SET tg_id=$1 WHERE phone_number=$2`
	res, err := db.Exec(query, id, phone)
	if err != nil {
		return err
	}
	a, _ := res.RowsAffected()
	if a == 0 {
		return fmt.Errorf("create parking batch row affected 0 -> %w", err)
	}
	return nil
}

func CheckUserDb(db sqlx.DB, id int64) (int, error) {
	var userId int
	query := `SELECT id FROM users WHERE tg_id=$1`
	err := db.Get(&userId, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return userId, fmt.Errorf("unf")
		}
		return userId, fmt.Errorf("dbe")
	}
	return userId, nil
}

func GetActiveUserGuestAccessList(db sqlx.DB, userID int) (*[]model.WhiteList, error) {
	data := new([]model.WhiteList)
	query := `SELECT id, plate_number, in_black_list, created_at,
       			expires_at, building_id, user_id, is_guest, is_tg_guest 
					FROM white_list 
						WHERE user_id = $1 and (is_guest = true or is_tg_guest = true)
							and expires_at > now()
								order by id DESC;`
	rows, err := db.Queryx(query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return data, nil
		}
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var row model.WhiteList
		if err = rows.Scan(&row.ID, &row.PlateNumber, &row.InBlackList, &row.CreatedAt, &row.ExpiresAt, &row.BuildingID, &row.UserID, &row.IsGuest); err != nil {
			return nil, err
		}
		*data = append(*data, row)
	}

	return data, nil
}

func AddUserGuestAccess(db sqlx.DB, wl *model.WhiteList) error {
	if wl.ExpiresAt == "" {
		wl.ExpiresAt = time.Now().AddDate(0, 0, 0).Add(time.Hour * 1).String()
	}
	_, err := db.Exec(
		"INSERT INTO white_list (plate_number, expires_at, building_id, user_id, is_guest, is_tg_guest) VALUES($1, $2, $3, $4, $5, $6)",
		strings.ToUpper(wl.PlateNumber), wl.ExpiresAt, wl.BuildingID, wl.UserID, true, true)
	if err != nil {
		return err
	}

	return nil
}
