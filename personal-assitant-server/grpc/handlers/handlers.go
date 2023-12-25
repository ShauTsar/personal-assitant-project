package handlers

import (
	"context"
	"errors"
	"log"
	"personal-assitant-project/personal-assitant-server/auth"
	"personal-assitant-project/personal-assitant-server/grpc/proto/gen"
	"personal-assitant-project/personal-assitant-server/storage/postgre"
	"personal-assitant-project/personal-assitant-server/storage/redis"
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
	err = postgre.SaveUserData(request) // Реализуйте эту функцию
	if err != nil {
		log.Printf("Error saving user data: %v", err)
		return nil, err
	}
	//_________________________
	token, err := auth.GenerateJWTToken(request.Username)
	if err != nil {
		return nil, err
	}
	err = redis.SaveRedisData(request, ctx, token)
	if err != nil {
		log.Printf("Error saving user cache: %v", err)
	}
	//______________________________________

	return &gen.RegisterResponse{Success: true}, nil
}

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
	redis.SaveRedisData(request, ctx, token)

	return &gen.LoginResponse{Token: token, Success: true, Message: message}, nil
}
