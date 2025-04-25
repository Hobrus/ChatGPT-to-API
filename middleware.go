package main

import (
    "bytes"
    "bufio"     // <-- обратите внимание на импорт "bufio"
    "io"
    "log"
    "os"
    "strings"
    "time"
    "encoding/json"
    gin "github.com/gin-gonic/gin"
)

var ADMIN_PASSWORD string
var API_KEYS map[string]bool

func init() {
    ADMIN_PASSWORD = os.Getenv("ADMIN_PASSWORD")
    if ADMIN_PASSWORD == "" {
        ADMIN_PASSWORD = "TotallySecurePassword"
    }
}

func adminCheck(c *gin.Context) {
    password := c.Request.Header.Get("Authorization")
    if password != ADMIN_PASSWORD {
        c.String(401, "Unauthorized")
        c.Abort()
        return
    }
    c.Next()
}

func cors(c *gin.Context) {
    c.Header("Access-Control-Allow-Origin", "*")
    c.Header("Access-Control-Allow-Methods", "*")
    c.Header("Access-Control-Allow-Headers", "*")
    c.Next()
}

func Authorization(c *gin.Context) {
    if API_KEYS == nil {
        API_KEYS = make(map[string]bool)
        if _, err := os.Stat("api_keys.txt"); err == nil {
            file, _ := os.Open("api_keys.txt")
            defer file.Close()

            // Вместо NewScanner используем bufio.NewScanner
            scanner := bufio.NewScanner(file)
            for scanner.Scan() {
                key := scanner.Text()
                if key != "" {
                    API_KEYS["Bearer "+key] = true
                }
            }
        }
    }
    auth := c.Request.Header.Get("Authorization")
    if len(API_KEYS) != 0 && !API_KEYS[auth] {
        if auth == "" {
            c.JSON(401, gin.H{"error": "No API key provided. Please provide an API key as part of the Authorization header."})
        } else if strings.HasPrefix(auth, "Bearer sk-") {
            c.JSON(401, gin.H{"error": "You tried to use the official API key which is not supported."})
        } else if strings.HasPrefix(auth, "Bearer eyJhbGciOiJSUzI1NiI") {
            // Разрешаем JWT на Bearer
            return
        } else {
            c.JSON(401, gin.H{"error": "Invalid API key."})
        }
        c.Abort()
        return
    }
    c.Next()
}

// Новое middleware: логирование запроса и ответа
func requestResponseLogger(c *gin.Context) {
    start := time.Now()

    // 1) Считываем тело входящего запроса
    var reqBodyBytes []byte
    if c.Request.Body != nil {
        reqBodyBytes, _ = io.ReadAll(c.Request.Body)
    }
    // Восстанавливаем тело, чтобы Gin мог прочитать его повторно
    c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))

    // Логируем запрос с правильной обработкой JSON
    log.Printf("---- REQUEST BEGIN ----")
    log.Printf("%s %s", c.Request.Method, c.Request.URL.String())
    log.Printf("HEADERS: %v", c.Request.Header)

    // Проверяем, есть ли Content-Type: application/json
    contentType := c.Request.Header.Get("Content-Type")
    if strings.Contains(contentType, "application/json") {
        log.Printf("BODY: %s", string(reqBodyBytes))

        // Проверка валидности JSON для диагностики
        var js json.RawMessage
        if err := json.Unmarshal(reqBodyBytes, &js); err != nil {
            log.Printf("WARNING: Invalid JSON in request: %v", err)
        }
    } else {
        log.Printf("BODY: %s", string(reqBodyBytes))
    }

    log.Printf("---- REQUEST END ----")

    // 2) Создаём обёртку для записи ответа в буфер
    blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
    c.Writer = blw

    // Выполняем запрос
    c.Next()

    // После обработки запроса забираем тело ответа
    responseBody := blw.body.String()
    statusCode := c.Writer.Status()
    latency := time.Since(start)

    // Логируем ответ
    log.Printf("---- RESPONSE BEGIN ----")
    log.Printf("STATUS: %d | DURATION: %v", statusCode, latency)
    log.Printf("RESPONSE BODY: %s", responseBody)
    log.Printf("---- RESPONSE END ----")
}

// Обёртка, которая перехватывает запись в тело ответа
type bodyLogWriter struct {
    gin.ResponseWriter
    body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
    // Дополнительно пишем данные в буфер
    w.body.Write(b)
    // И отправляем их клиенту
    return w.ResponseWriter.Write(b)
}
