package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"personal-assitant-project/config"
	"personal-assitant-project/personal-assitant-server/storage/elastic"
	"strconv"
)

var redisClient *redis.Client

func init() {
	redisConfig := config.LoadRedisConfig()
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisConfig.Addr,
		DB:   redisConfig.DB,
	})
}

func SaveRedisData(userID int, token string, ctx context.Context) error {
	err := redisClient.HSet(ctx, "user_tokens", token, userID).Err()
	if err != nil {
		log.Printf("Error saving userID-token mapping to Redis: %v", err)
		return err
	}
	return nil
}
func GetUserIDByToken(token string, ctx context.Context) (int, error) {
	userIDStr, err := redisClient.HGet(ctx, "user_tokens", token).Result()
	if err != nil {
		if errors.Is(redis.Nil, err) {
			//TODO можно сделать логи, где будет видно пользователя, который делает запрос к боту
			elastic.LogToElasticsearch(fmt.Sprintf("Token not found in Redis: %s", token))
			log.Printf("Token not found in Redis: %s", token)
			return 0, fmt.Errorf("Token not found")
		}
		elastic.LogToElasticsearch(fmt.Sprintf("Error retrieving userID from Redis: %v", err))
		log.Printf("Error retrieving userID from Redis: %v", err)
		return 0, err
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Error converting userID to integer: %v", err)
		return 0, err
	}

	return userID, nil
}

//func generateSessionToken() string {
//	// Генерация уникального токена (можно использовать как случайную строку)
//	b := make([]byte, 32)
//	_, err := rand.Read(b)
//	if err != nil {
//		log.Fatalf("Error generating session token: %v", err)
//	}
//	return base64.StdEncoding.EncodeToString(b)
//}
