package handlers

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"personal-assitant-project/personal-assitant-server/auth"
	gen "personal-assitant-project/personal-assitant-server/grpc/proto/gen"
	"personal-assitant-project/personal-assitant-server/storage/postgre"
	"personal-assitant-project/personal-assitant-server/storage/redis"
	"time"
)

type UserServiceServer struct {
	gen.UnimplementedUserServiceServer
}

func (s *UserServiceServer) Register(ctx context.Context, request *gen.RegisterRequest) (*gen.RegisterResponse, error) {
	userExists, err := postgre.CheckUserExists(request.Username)
	if err != nil {
		log.Printf("Error checking user: %v", err)
		return nil, err
	}

	if userExists {
		return nil, err
	}

	// Если пользователь не существует, можно продолжить регистрацию
	// Сгенерируйте код подтверждения и отправьте его на почту (например, через SMTP)
	// Сохраните данные пользователя в PostgreSQL
	err = postgre.SaveUserData(request)
	if err != nil {
		log.Printf("Error saving user data: %v", err)
		return nil, err
	}
	//_________________________
	token, err := auth.GenerateJWTToken(request.Username)
	if err != nil {
		return nil, err
	}
	usernameId, err := postgre.GetUserID(request.Username)
	if err != nil {
		return nil, err
	}
	err = redis.SaveRedisData(usernameId, token, ctx)
	if err != nil {
		log.Printf("Error saving user cache: %v", err)
	}
	//______________________________________

	return &gen.RegisterResponse{Success: true}, nil
}

// TODO добавить время истечения токена, например, 5 часов
func (s *UserServiceServer) Login(ctx context.Context, request *gen.LoginRequest) (*gen.LoginResponse, error) {
	var message string
	success, err := postgre.CheckCredentials(request.Username, request.Password)
	if err != nil {
		return nil, errors.New("Db error")
	}
	if !success {
		message = "User doesn't exist"
		return &gen.LoginResponse{Success: false, Message: message}, nil
	}

	token, err := auth.GenerateJWTToken(request.Username)
	if err != nil {
		return nil, err
	}
	usernameId, err := postgre.GetUserID(request.Username)
	if err != nil {
		return nil, err
	}
	err = redis.SaveRedisData(usernameId, token, ctx)

	return &gen.LoginResponse{Token: token, Success: true, Message: message}, nil
}
func (s *UserServiceServer) UpdateTelegramUserID(ctx context.Context, request *gen.UpdateTelegramUserIDRequest) (*gen.UpdateTelegramUserIDResponse, error) {
	userID, err := redis.GetUserIDByToken(request.Token, ctx)
	if err != nil {
		log.Printf("Error getting userID from Redis: %v", err)
		return &gen.UpdateTelegramUserIDResponse{Success: false, Message: "Unknown error"}, err
	}

	success, err := postgre.UpdateTelegramUser(int(request.UserTelegramID), userID)
	if err != nil {
		log.Printf("Error updating Telegram user ID: %v", err)
		return &gen.UpdateTelegramUserIDResponse{Success: false, Message: "Unknown error"}, err
	}

	log.Printf("Success: UserID %d, TelegramUserID %d", userID, request.UserTelegramID)
	return &gen.UpdateTelegramUserIDResponse{Success: success, Message: "Registration success"}, nil
}
func (s *UserServiceServer) AddEventData(ctx context.Context, request *gen.AddEventDataRequest) (*gen.AddEventDataResponse, error) {
	userID, err := redis.GetUserIDByToken(request.EventData.Token, ctx)
	if err != nil {
		log.Printf("Error getting userID from Redis: %v", err)
		return &gen.AddEventDataResponse{Success: false, Message: "Unknown error"}, err
	}
	startDate, err := timeParsing(request.EventData.StartDate)
	if err != nil {
		log.Printf("Error parsing start date %v", err)
		return &gen.AddEventDataResponse{Success: false, Message: "Unknown error"}, err
	}
	var finishedDate time.Time
	if request.EventData.IsFinished {
		finishedDate, err = timeParsing(request.EventData.FinishedDate)
		if err != nil {
			log.Printf("Error parsing finished date %v", err)
			return &gen.AddEventDataResponse{Success: false, Message: "Unknown error"}, err
		}
	}
	plannedDate, err := timeParsing(request.EventData.PlannedDate)
	if err != nil {
		log.Printf("Error parsing planned date %v", err)
		return &gen.AddEventDataResponse{Success: false, Message: "Unknown error"}, err
	}
	err = postgre.AddTask(userID, startDate, plannedDate, finishedDate, request.EventData.Description, request.EventData.IsFinished, request.EventData.Attachment, request.EventData.Title)
	if err != nil {
		log.Printf("Error parsing planned date %v", err)
		return &gen.AddEventDataResponse{Success: false, Message: "Unknown error"}, err
	}
	return &gen.AddEventDataResponse{Success: true, Message: "Success"}, nil
}
func timeParsing(date string) (time.Time, error) {
	layout := time.RFC3339
	formatDate, err := time.Parse(layout, date)
	if err != nil {
		log.Printf("Error parsing date %v", err)
		return time.Now(), err
	}
	return formatDate, nil
}
func (s *UserServiceServer) GetAllEvents(ctx context.Context, request *gen.GetAllEventsRequest) (*gen.GetAllEventsResponse, error) {
	userID, err := redis.GetUserIDByToken(request.Token, ctx)
	if err != nil {
		log.Printf("Error getting userID from Redis: %v", err)
		return &gen.GetAllEventsResponse{}, err
	}
	events, _ := postgre.ShowTasksByUserID(userID)
	return &gen.GetAllEventsResponse{Events: events}, nil

}
func (s *UserServiceServer) FinishEvent(ctx context.Context, request *gen.FinishEventRequest) (*gen.FinishEventResponse, error) {

	err := postgre.FinishTask(int(request.TaskID), request.Finish)
	if err != nil {
		log.Printf("Error updating task: %v", err)
		return &gen.FinishEventResponse{Success: false, Message: "Unknown error"}, err
	}

	return &gen.FinishEventResponse{Success: true, Message: "Task updated successfully"}, nil
}
func (s *UserServiceServer) UpdateUserSettings(ctx context.Context, request *gen.UpdateUserSettingsRequest) (*gen.UpdateUserSettingsResponse, error) {
	settings := request.Settings
	userID, err := redis.GetUserIDByToken(settings.Token, ctx)
	if err != nil {

	}
	avatarURL := ""
	if len(settings.AvatarUrl) != 0 {
		url, err := postgre.GetAvatarUrl(userID)
		if err != nil {

		}
		if url != "" {
			os.Remove(url)
		}
		avatarURL, err = s.saveAvatar(settings.GetAvatarUrl())
		if err != nil {
			return nil, err
		}
	}
	postgre.UpdateUserSettings(settings, userID, avatarURL)

	return &gen.UpdateUserSettingsResponse{Success: true, Message: "Settings updated successfully"}, nil
}

const avatarDir = "personal-assistant-web/public/avatars"

func (s *UserServiceServer) saveAvatar(avatarData []byte) (string, error) {
	hash := sha256.Sum256(avatarData)
	fileName := fmt.Sprintf("%x.jpg", hash)
	filePath := filepath.Join(avatarDir, fileName)
	err := os.MkdirAll(avatarDir, 0755)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(filePath, avatarData, 0644)
	if err != nil {
		return "", err
	}

	avatarURL := fmt.Sprintf("personal-assistant-web/public/avatars/%s", fileName)
	return avatarURL, nil
}
func (s *UserServiceServer) GetUserSettings(ctx context.Context, request *gen.UserSettingsRequest) (*gen.UserSettingsResponse, error) {
	userID, err := redis.GetUserIDByToken(request.Token, ctx)
	if err != nil {
		return nil, err
	}
	settings, err := postgre.GetUserSettings(userID)
	if err != nil {
		return nil, err
	}
	return &gen.UserSettingsResponse{Settings: settings}, nil
}
func (s *UserServiceServer) GetFinances(ctx context.Context, request *gen.GetFinancesRequest) (*gen.GetFinancesResponse, error) {
	userID, err := redis.GetUserIDByToken(request.Token, ctx)
	if err != nil {
		return nil, err
	}
	finances, err := postgre.GetFinances(userID)
	if err != nil {
		return nil, err
	}
	return &gen.GetFinancesResponse{Finances: finances}, nil
}
func (s *UserServiceServer) GetCategories(ctx context.Context, request *gen.GetCategoriesRequest) (*gen.GetCategoriesResponse, error) {
	userID, err := redis.GetUserIDByToken(request.Token, ctx)
	if err != nil {

	}
	categories, err := postgre.GetFinanceCategory(userID)
	if err != nil {

	}
	//fmt.Println(finances)
	return &gen.GetCategoriesResponse{Category: categories}, nil
}
func (s *UserServiceServer) AddFinance(ctx context.Context, request *gen.AddFinanceRequest) (*gen.AddFinanceResponse, error) {
	userID, err := redis.GetUserIDByToken(request.Finances.Token, ctx)
	if err != nil {
		return nil, err
	}
	err = postgre.AddFinance(request.Finances, userID)
	if err != nil {
		return &gen.AddFinanceResponse{Message: err.Error(), Success: false}, err
	}
	return &gen.AddFinanceResponse{Message: "", Success: true}, nil

}
func (s *UserServiceServer) DeleteFinance(ctx context.Context, request *gen.DeleteFinanceRequest) (*gen.DeleteFinanceResponse, error) {
	err := postgre.DeleteFinance(request.FinanceId)
	if err != nil {
		return &gen.DeleteFinanceResponse{Message: err.Error(), Success: false}, err
	}
	return &gen.DeleteFinanceResponse{Message: "", Success: true}, nil

}
func (s *UserServiceServer) AddCategory(ctx context.Context, request *gen.AddCategoryRequest) (*gen.AddCategoryResponse, error) {
	userID, err := redis.GetUserIDByToken(request.Token, ctx)
	if err != nil {
		return nil, err
	}
	err = postgre.AddCategory(request.Category, userID)
	if err != nil {
		return &gen.AddCategoryResponse{Message: err.Error(), Success: false}, err
	}
	return &gen.AddCategoryResponse{Message: "", Success: true}, nil

}
func (s *UserServiceServer) ArchiveEvent(ctx context.Context, request *gen.ArchiveEventRequest) (*gen.ArchiveEventResponse, error) {
	err := postgre.ArchiveEvent(request.TaskId)
	if err != nil {
		return nil, err
	}
	return &gen.ArchiveEventResponse{Message: "Success", Success: true}, nil

}
