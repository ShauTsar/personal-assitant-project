package handlers

import (
	"context"
	"log"
	"personal-assitant-project/personal-assitant-server/grpc/proto"
	"personal-assitant-project/personal-assitant-server/storage/postgre"
)

type UserServiceHandler struct {
}

func NewUserServiceHandler() *UserServiceHandler {
	return &UserServiceHandler{}
}

func (s *UserServiceHandler) RegisterUser(ctx context.Context, request *proto.RegisterRequest) (*proto.RegisterResponse, error) {
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
	err = postgre.SaveUserData(request) // Реализуйте эту функцию
	if err != nil {
		log.Printf("Error saving user data: %v", err)
		return nil, err
	}

	// Сохраните данные пользователя в Redis (например, токен сессии)

	// Верните успешный ответ
	return &proto.RegisterResponse{Success: true}, nil
}
