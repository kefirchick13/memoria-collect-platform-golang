package main

import (
	"log"
	"os"

	"github.com/joho/godotenv" // чтение env файлов
	"github.com/kefirchick13/memoria-collect-platform-golang"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/handler"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/repository"
	"github.com/kefirchick13/memoria-collect-platform-golang/internal/service"
	_ "github.com/lib/pq"
	"github.com/spf13/viper" // чтение конфиг файлов разных  форматов
	"go.uber.org/zap"        // самый быстрый логгер для go от uber
)

func main() {
	// Logger init
	customLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Logger cant't init")
	}
	log := customLogger.Sugar()

	// Init app config
	if err := initAppConfig(); err != nil {
		log.Fatal(err)
	}

	// Init env variables
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(".env not found; continuing without it, ", err)
	}

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString(("db.port")),
		Username: viper.GetString("db.username"),
		DBName:   viper.GetString("db.name"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: os.Getenv("DB_PASSWORD"),
	})

	if err != nil {
		log.Fatal(err)
	}

	repos := repository.NewRepository(db, log)
	services := service.NewService(repos, log)
	handlers := handler.NewHandler(services, log)

	server := memoria.Server{}

	if err := server.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
		log.Fatalf("Error during pushing: %s", err.Error())
	}
}

func initAppConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("configs")
	return viper.ReadInConfig()
}
