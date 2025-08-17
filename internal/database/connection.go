package database

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect conecta ao banco de dados PostgreSQL
func Connect(databaseURL string) (*gorm.DB, error) {
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, err
	}

	// Configurar pool de conexões
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Conectado ao banco de dados PostgreSQL")
	return db, nil
}

// ConnectRedis conecta ao Redis (opcional)
func ConnectRedis(redisURL string) *redis.Client {
	if redisURL == "" {
		log.Println("Redis URL não configurada, continuando sem Redis")
		return nil
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("Erro ao parsear URL do Redis: %v, continuando sem Redis", err)
		return nil
	}

	client := redis.NewClient(opt)

	// Testar conexão
	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Erro ao conectar com Redis: %v, continuando sem Redis", err)
		return nil
	}

	log.Println("Conectado ao Redis")
	return client
}
