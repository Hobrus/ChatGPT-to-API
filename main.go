package main

import (
	"bufio"
	"freechatgpt/internal/tokens"
	"os"
	"strings"

	chatgpt_types "freechatgpt/internal/chatgpt"

	"github.com/acheong08/endless"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Объявляем глобальные переменные
var HOST string
var PORT string
var ACCESS_TOKENS tokens.AccessToken
var proxies []string

func checkProxy() {
	// first check for proxies.txt
	proxies = []string{}
	if _, err := os.Stat("proxies.txt"); err == nil {
		// Each line is a proxy, put in proxies array
		file, _ := os.Open("proxies.txt")
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			proxy := scanner.Text()
			proxy_parts := strings.Split(proxy, ":")
			if len(proxy_parts) > 1 {
				proxies = append(proxies, proxy)
			} else {
				continue
			}
		}
	}
	// if no proxies, then check env http_proxy
	if len(proxies) == 0 {
		proxy := os.Getenv("http_proxy")
		if proxy != "" {
			proxies = append(proxies, proxy)
		}
	}
}

func init() {
	_ = godotenv.Load(".env")

	HOST = os.Getenv("SERVER_HOST")
	PORT = os.Getenv("SERVER_PORT")
	if HOST == "" {
		HOST = "127.0.0.1"
	}
	if PORT == "" {
		PORT = "8080"
	}
	checkProxy()
	readAccounts()
	scheduleTokenPUID()
}

func main() {
	// Сохраняем хэш-файлы при завершении
	defer chatgpt_types.SaveFileHash()

	// Инициализируем Gin
	router := gin.Default()

	// Подключаем наши middleware:
	// 1) requestResponseLogger - чтобы логировать запрос/ответ
	// 2) cors - ваше уже существующее middleware
	router.Use(requestResponseLogger)
	router.Use(cors)

	// Роут для тестирования
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Группа админских роутов
	admin_routes := router.Group("/admin")
	admin_routes.Use(adminCheck)
	{
		admin_routes.PATCH("/password", passwordHandler)
		admin_routes.PATCH("/tokens", tokensHandler)
	}

	// Публичные роуты
	router.OPTIONS("/v1/chat/completions", optionsHandler)
	router.POST("/v1/chat/completions", Authorization, nightmare)

	router.OPTIONS("/v1/audio/speech", optionsHandler)
	router.POST("/v1/audio/speech", Authorization, tts)

	router.OPTIONS("/v1/audio/transcriptions", optionsHandler)
	router.POST("/v1/audio/transcriptions", Authorization, stt)

	router.OPTIONS("/v1/models", optionsHandler)
	router.GET("/v1/models", Authorization, simulateModel)

	// Запускаем сервер
	endless.ListenAndServe(HOST+":"+PORT, router)
}

