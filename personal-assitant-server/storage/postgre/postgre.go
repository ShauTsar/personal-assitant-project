package postgre

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"personal-assitant-project/config"
	userpb "personal-assitant-project/personal-assitant-server/grpc/proto/gen"
	"time"
)

func CheckUserExists(username string) (bool, error) {
	db := LoadDBFromConfig()
	query := "SELECT COUNT(*) FROM users WHERE username = $1"
	var count int
	if err := db.QueryRow(query, username).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}
func GetUserID(username string) (int, error) {
	db := LoadDBFromConfig()
	var id int
	query := "SELECT id FROM users WHERE username = $1"
	if err := db.QueryRow(query, username).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return id, nil
}
func SaveUserData(request *userpb.RegisterRequest) error {
	db := LoadDBFromConfig()
	query := `
        INSERT INTO users (username, password, email, phone, timezone)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := db.Exec(query, request.Username, request.Password, request.Email, request.Phone, request.Timezone)
	if err != nil {
		return err
	}

	return nil
}
func LoadDBFromConfig() *sql.DB {
	dbConfig := config.LoadDatabaseConfig()
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Password, dbConfig.Database, dbConfig.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		// Здесь нужно обработать ошибку, если соединение не установлено.
		panic(err)
	}
	return db
}
func CheckCredentials(username, password string) (bool, error) {
	db := LoadDBFromConfig()
	defer db.Close()

	query := "SELECT COUNT(*) FROM users WHERE username = $1 AND password = $2"
	var count int
	if err := db.QueryRow(query, username, password).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return count > 0, nil
}

// TODO 1 проверять есть ли такой tg_user_id, 2 проверять если ли user_id
func UpdateTelegramUser(userTelegramId int, userID int) (bool, error) {
	db := LoadDBFromConfig()
	defer db.Close()
	isExists, err := TgUserExists(userTelegramId)
	if err != nil {
		return false, err
	}
	if isExists {
		return true, nil
	} else {
		isExists, err = WebUserExistsInBot(userTelegramId)
		if err != nil {
			return false, err
		}
		if isExists {
			query := "UPDATE tbot SET tg_user_id = $1 WHERE user_id = $2"
			_, err := db.Exec(query, userTelegramId, userID)
			if err != nil {
				log.Printf("Error updating tg_user_id in tbot table: %v", err)
				return false, err
			}
		} else {
			query := `
			INSERT INTO tbot (user_id, tg_user_id)
			VALUES ($1, $2)
			`
			_, err := db.Exec(query, userID, userTelegramId)
			if err != nil {
				log.Printf("Error updating tg_user_id in tbot table: %v", err)
				return false, err
			}
		}
	}
	return true, nil
}
func AddTask(userID int, startDate, plannedDate, finishedDate time.Time, description string, isFinished bool, attachment []byte, title string) error {
	db := LoadDBFromConfig()
	defer db.Close()
	query := `
        INSERT INTO tasks (user_id, start_date, planned_date, finished_date, description, is_finished, attachment, title)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err := db.Exec(query, userID, startDate, plannedDate, finishedDate, description, isFinished, attachment, title)
	if err != nil {
		return err
	}

	return nil
}

type EventData struct {
	TaskID       int
	UserID       int
	StartDate    time.Time
	PlannedDate  time.Time
	FinishedDate time.Time
	Description  string
	IsFinished   bool
	Attachment   []byte
	Title        string
}
type AllTBotUsers struct {
	TGUserID int64
	UserID   int
}

func ShowTasksByUserID(userID int) ([]*userpb.EventDataMessage, error) {
	db := LoadDBFromConfig()
	defer db.Close()

	events := []*userpb.EventDataMessage{}

	query := "SELECT id, user_id, start_date, planned_date, finished_date, description, is_finished, attachment, title FROM tasks WHERE user_id = $1"
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var eventData EventData
		if err := rows.Scan(
			&eventData.TaskID,
			&eventData.UserID,
			&eventData.StartDate,
			&eventData.PlannedDate,
			&eventData.FinishedDate,
			&eventData.Description,
			&eventData.IsFinished,
			&eventData.Attachment,
			&eventData.Title,
		); err != nil {
			return nil, err
		}

		eventDataMessage := EventDataToEventDataMessage(eventData)
		events = append(events, &eventDataMessage)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func EventDataToEventDataMessage(eventData EventData) userpb.EventDataMessage {
	return userpb.EventDataMessage{
		Token:        "",
		TaskId:       int32(eventData.TaskID),
		UserId:       int32(eventData.UserID),
		StartDate:    eventData.StartDate.Format(time.RFC3339),
		PlannedDate:  eventData.PlannedDate.Format(time.RFC3339),
		FinishedDate: eventData.FinishedDate.Format(time.RFC3339),
		Description:  eventData.Description,
		IsFinished:   eventData.IsFinished,
		Attachment:   eventData.Attachment,
		Title:        eventData.Title,
	}
}
func TgUserExists(userTelegramId int) (bool, error) {
	db := LoadDBFromConfig()
	defer db.Close()
	query := "SELECT COUNT(*) FROM tbot WHERE tg_user_id = $1"
	var count int
	if err := db.QueryRow(query, userTelegramId).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}
func WebUserExistsInBot(userID int) (bool, error) {
	db := LoadDBFromConfig()
	defer db.Close()
	query := "SELECT COUNT(*) FROM tbot WHERE user_id = $1"
	var count int
	if err := db.QueryRow(query, userID).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}
func GetUserIDFromTasks(tgUserID int64) (int, error) {
	db := LoadDBFromConfig()
	defer db.Close()
	query := "SELECT user_id FROM tbot WHERE tg_user_id = $1"
	var userID int
	if err := db.QueryRow(query, tgUserID).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userID, nil
		}
		return 0, err
	}
	return userID, nil
}
func FinishTask(taskID int, finish bool) error {
	db := LoadDBFromConfig()
	defer db.Close()

	query := "UPDATE tasks SET is_finished = $1 WHERE id = $2"
	_, err := db.Exec(query, finish, taskID)
	if err != nil {
		log.Printf("Error updating tg id in tasks table: %v", err)
		return err
	}
	return nil
}

func FindTasksByUser(userID int) ([]EventData, error) {
	db := LoadDBFromConfig()
	defer db.Close()

	query := "SELECT id, user_id, start_date, planned_date, finished_date, description, is_finished, attachment, title FROM tasks WHERE user_id = $1"
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []EventData
	for rows.Next() {
		var eventData EventData
		if err := rows.Scan(
			&eventData.TaskID,
			&eventData.UserID,
			&eventData.StartDate,
			&eventData.PlannedDate,
			&eventData.FinishedDate,
			&eventData.Description,
			&eventData.IsFinished,
			&eventData.Attachment,
			&eventData.Title,
		); err != nil {
			return nil, err
		}
		result = append(result, eventData)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
func GetAllTGUsersID() ([]AllTBotUsers, error) {
	db := LoadDBFromConfig()
	defer db.Close()

	query := "SELECT tg_user_id, user_id FROM tbot"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []AllTBotUsers
	for rows.Next() {
		var tBotUser AllTBotUsers
		if err := rows.Scan(
			&tBotUser.TGUserID,
			&tBotUser.UserID,
		); err != nil {
			return nil, err
		}
		result = append(result, tBotUser)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
func UpdateUserSettings(settings *userpb.UserSettings, userID int, avatarURL string) error {
	db := LoadDBFromConfig()
	defer db.Close()
	isUserSettingsExist, err := UserSettingsExists(userID)
	if err != nil {

	}
	if isUserSettingsExist {
		query := "UPDATE users_notification_settings SET for_morning = $1, for_evening = $2, for_task = $3 WHERE user_id = $4"
		_, err := db.Exec(query, settings.NotifyMorningHours, settings.NotifyEveryEveningHours, settings.NotifyBeforeEventHours, userID)
		if err != nil {
			log.Printf("Error updating tg id in tasks table: %v", err)
			return err
		}
		if avatarURL != "" {
			query = "UPDATE users SET timezone = $1, avatar_url = $2 WHERE id = $3"
			_, err = db.Exec(query, settings.Timezone, avatarURL, userID)
			if err != nil {
				log.Printf("Error updating tg id in tasks table: %v", err)
				return err
			}
		} else {
			query = "UPDATE users SET timezone = $1 WHERE id = $2"
			_, err = db.Exec(query, settings.Timezone, userID)
			if err != nil {
				log.Printf("Error updating tg id in tasks table: %v", err)
				return err
			}
		}

	} else {
		query := `
        INSERT INTO users_notification_settings (user_id, for_morning, for_evening, for_task)
        VALUES ($1, $2, $3, $4)
    `
		_, err := db.Exec(query, userID, settings.NotifyMorningHours, settings.NotifyEveryEveningHours, settings.NotifyBeforeEventHours)
		if err != nil {
			return err
		}
	}
	return nil
}
func UserSettingsExists(userID int) (bool, error) {
	db := LoadDBFromConfig()
	defer db.Close()
	query := "SELECT COUNT(*) FROM users_notification_settings WHERE user_id = $1"
	var count int
	if err := db.QueryRow(query, userID).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}
func GetUserSettings(userID int) (*userpb.UserSettings, error) {
	db := LoadDBFromConfig()
	defer db.Close()
	settings := &userpb.UserSettings{}
	query := "SELECT for_morning, for_evening, for_task FROM users_notification_settings WHERE user_id = $1"
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(
			&settings.NotifyMorningHours,
			&settings.NotifyEveryEveningHours,
			&settings.NotifyBeforeEventHours,
		); err != nil {
			return nil, err
		}
	}
	settings.Token = ""
	query = "SELECT timezone, avatar_url from users where id = $1"
	if err := db.QueryRow(query, userID).Scan(&settings.Timezone, &settings.AvatarUrl); err != nil {
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return settings, nil
}
func GetAvatarUrl(userID int) (string, error) {
	db := LoadDBFromConfig()
	defer db.Close()
	query := "SELECT avatar_url FROM users WHERE id = $1"
	var avaratUrl string
	if err := db.QueryRow(query, userID).Scan(&avaratUrl); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return avaratUrl, nil
		}
		return "", err
	}
	return avaratUrl, nil
}

//type UserSettings struct {
//	UserID      int32
//	ForMorning  int32
//	ForEvening  int32
//	TimeZone    string
//	IsDarkTheme bool
//}

func GetFinances(userID int) ([]*userpb.Finances, error) {
	db := LoadDBFromConfig()
	defer db.Close()
	finances := []*userpb.Finances{}
	query := "SELECT id, category_id, fin_date, description, price, is_expense FROM finances WHERE user_id = $1"
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		financesRow := &userpb.Finances{}
		categorys := &userpb.Category{}
		var category int
		if err := rows.Scan(
			&financesRow.FinanceId,
			&category,
			&financesRow.Date,
			&financesRow.Description,
			&financesRow.Price,
			&financesRow.IsExpense,
		); err != nil {
			return nil, err
		}
		query = "SELECT name, is_for_all_users from finances_categories where id = $1"
		if err := db.QueryRow(query, category).Scan(&categorys.Name, &categorys.IsForAll); err != nil {
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		financesRow.Category = categorys
		finances = append(finances, financesRow)
	}
	return finances, nil
}

//	func GetFinances(userID int) ([]*userpb.Finances, error) {
//		db := LoadDBFromConfig()
//		defer db.Close()
//		finances := []*userpb.Finances{}
//		query := "SELECT id, category_id, fin_date, description, price, is_expense FROM finances WHERE user_id = $1"
//		rows, err := db.Query(query, userID)
//		if err != nil {
//			return nil, err
//		}
//		defer rows.Close()
//		for rows.Next() {
//			financesRow := &userpb.Finances{}
//			if err := rows.Scan(
//				&financesRow.FinanceId,
//				&financesRow.Category.Id,
//				&financesRow.Date,
//				&financesRow.Description,
//				&financesRow.Price,
//				&financesRow.IsExpense,
//			); err != nil {
//				return nil, err
//			}
//			query = "SELECT name, is_for_all_users from finances_categories where id = $1"
//			if err := db.QueryRow(query, &financesRow.Category.Id).Scan(&financesRow.Category.Name, &financesRow.Category.IsForAll); err != nil {
//			}
//			if err := rows.Err(); err != nil {
//				fmt.Println(err)
//				return nil, err
//			}
//			finances = append(finances, financesRow)
//		}
//		return finances, nil
//	}
func GetFinanceCategory(userID int) ([]*userpb.Category, error) {
	db := LoadDBFromConfig()
	defer db.Close()
	categories := []*userpb.Category{}
	query := "SELECT id, name, is_for_all_users  FROM finances_categories WHERE user_id = $1 or is_for_all_users = true"
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		categoryRow := &userpb.Category{}
		if err := rows.Scan(
			&categoryRow.Id,
			&categoryRow.Name,
			&categoryRow.IsForAll,
		); err != nil {
			return nil, err
		}
		categories = append(categories, categoryRow)
	}
	return categories, nil
}

//func FinanceToFinanceMessage(finance Finance) userpb.Finances {
//	return userpb.Finances{
//		Token:        "",
//
//	}
//}
//type Finance struct {
//	TaskID       int
//	UserID       int
//	StartDate    time.Time
//	PlannedDate  time.Time
//	FinishedDate time.Time
//	Description  string
//	IsFinished   bool
//	Attachment   []byte
//	Title        string
//}

func AddFinance(finances *userpb.Finances, userID int) error {
	db := LoadDBFromConfig()
	defer db.Close()
	query := `
        INSERT INTO finances (category_id, fin_date, description, price, is_expense,user_id)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := db.Exec(query, finances.Category.Id, finances.Date, finances.Description, finances.Price, finances.IsExpense, userID)
	if err != nil {
		return err
	}
	return nil
}

func DeleteFinance(financeID int32) error {
	db := LoadDBFromConfig()
	defer db.Close()
	query := `
        Delete from finances where id = $1
    `
	_, err := db.Exec(query, financeID)
	if err != nil {
		return err
	}
	return nil
}

func AddCategory(category *userpb.Category, userID int) error {
	db := LoadDBFromConfig()
	defer db.Close()
	//TODO Специалльно не добавлял для всех юзеров
	query := `
        INSERT INTO finances_categories (name, user_id) 
        VALUES ($1, $2)
    `
	//Остоновился над этим, нужно подумать когда создается категория
	_, err := db.Exec(query, category.Name, userID)
	if err != nil {
		return err
	}
	return nil
}

func ArchiveEvent(taskID int32) error {
	db := LoadDBFromConfig()
	defer db.Close()
	archivedDate := time.Now()
	query := `
        INSERT INTO archived_tasks (user_id, start_date, planned_date, finished_date, description, attachment, title, archived_date)
        SELECT user_id, start_date, planned_date, finished_date, description, attachment, title, $2 From tasks WHERE id = $1
    `
	_, err := db.Exec(query, taskID, archivedDate)
	if err != nil {
		return err
	}
	query = `
        Delete from tasks where id = $1
    `
	_, err = db.Exec(query, taskID)
	if err != nil {
		return err
	}
	return nil
}

//func RestoreEvent ( ){
//
//}
