package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"personal-assitant-project/config"
)

var redisClient *redis.Client

func init() {
	redisConfig := config.LoadRedisConfig()
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisConfig.Addr,
		DB:   redisConfig.DB,
	})
}

func SaveRedisData(data interface{}, ctx context.Context, token string) error {
	err := redisClient.Set(ctx, token, data, 0).Err()
	if err != nil {
		log.Printf("Error saving session token to Redis: %v", err)
		return err
	}
	return nil
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
