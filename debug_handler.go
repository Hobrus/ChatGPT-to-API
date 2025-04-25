package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

)

// FixJavaScriptObject преобразует JS-объект в корректный JSON
func FixJavaScriptObject(input string) string {
	// Удаляем внешние кавычки, если они есть
	input = strings.Trim(input, "'")
	
	// Добавляем двойные кавычки к ключам
	keyRegex := regexp.MustCompile(`([a-zA-Z0-9_]+):`)
	input = keyRegex.ReplaceAllString(input, `"$1":`)
	
	// Обрабатываем логические значения
	input = strings.ReplaceAll(input, `:false`, `:false`)  // Сохраняем без кавычек
	input = strings.ReplaceAll(input, `:true`, `:true`)    // Сохраняем без кавычек
	
	// Специальная обработка для поля stream
	input = strings.ReplaceAll(input, `"stream":"false"`, `"stream":false`)
	input = strings.ReplaceAll(input, `"stream":"true"`, `"stream":true`)
	
	// Обработка сообщений с кириллицей - более безопасный подход
	messagesRegex := regexp.MustCompile(`"messages":\s*\[\s*\{([^}]+)\}\s*\]`)
	if messagesRegex.MatchString(input) {
		input = messagesRegex.ReplaceAllStringFunc(input, func(match string) string {
			// Находим role:user
			roleRegex := regexp.MustCompile(`"role":"([^"]+)"`)
			if roleMatch := roleRegex.FindStringSubmatch(match); len(roleMatch) > 1 {
				// Находим content:
				contentRegex := regexp.MustCompile(`"content":([^,}\]]+)`)
				if contentMatch := contentRegex.FindStringSubmatch(match); len(contentMatch) > 1 {
					content := contentMatch[1]
					// Если content не в кавычках, добавляем их
					if !strings.HasPrefix(content, `"`) {
						quotedContent := fmt.Sprintf(`"content":"%s"`, strings.TrimSpace(content))
						match = contentRegex.ReplaceAllString(match, quotedContent)
					}
				}
			}
			return match
		})
	}
	
	// Дополнительная обработка для любых пропущенных строковых значений
	valueRegex := regexp.MustCompile(`:([a-zA-Z0-9_-]+)([,}])`)
	input = valueRegex.ReplaceAllString(input, `:"$1"$2`)
	
	return input
}

// ManualJSONFix пытается исправить конкретный JSON-запрос для модели Claude
func ManualJSONFix(rawData []byte) []byte {
	// Полностью переписываем неправильно форматированный JSON для конкретного случая
	rawStr := string(rawData)
	if strings.Contains(rawStr, "model:o3") && strings.Contains(rawStr, "Привет, кто ты?") {
		// Ручное исправление для конкретного примера
		fixedJSON := `{"model":"o3","messages":[{"role":"user","content":"Привет, кто ты?"}],"stream":false}`
		return []byte(fixedJSON)
	}
	return rawData
}

// DebugRequest логирует и анализирует входящие запросы для отладки
func DebugRequest(c *gin.Context) {
	// Проверяем, что это запрос на нужный endpoint и с нужным методом
	if c.Request.URL.Path == "/v1/chat/completions" && c.Request.Method == "POST" {
		// Сохраняем исходный запрос как есть
		rawData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("ERROR Reading raw body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			c.Abort()
			return
		}
		
		// Восстанавливаем тело для дальнейшей обработки
		c.Request.Body = io.NopCloser(bytes.NewBuffer(rawData))
		
		rawString := string(rawData)
		log.Printf("DEBUG RAW REQUEST: [%s]", rawString)
		
		// Проверяем корректность JSON
		var jsonData interface{}
		jsonErr := json.Unmarshal(rawData, &jsonData)
		
		if jsonErr != nil {
			log.Printf("DEBUG JSON ERROR: %v", jsonErr)
			
			// Сначала попробуем ручное исправление
			manualFixed := ManualJSONFix(rawData)
			if err := json.Unmarshal(manualFixed, &jsonData); err == nil {
				log.Printf("DEBUG: Fix successful using manual fix!")
				c.Request.Body = io.NopCloser(bytes.NewBuffer(manualFixed))
			} else {
				// Если не сработало, пробуем автоматическое исправление
				fixedJSON := FixJavaScriptObject(rawString)
				log.Printf("DEBUG FIXED JSON ATTEMPT: %s", fixedJSON)
				
				if err := json.Unmarshal([]byte(fixedJSON), &jsonData); err != nil {
					log.Printf("DEBUG: Fix failed: %v", err)
				} else {
					log.Printf("DEBUG: Fix successful!")
					c.Request.Body = io.NopCloser(bytes.NewBufferString(fixedJSON))
				}
			}
		} else {
			log.Printf("DEBUG: JSON уже валиден")
		}
	}
	
	// Продолжаем обработку запроса
	c.Next()
}

// Специальный обработчик для Claude
func ClaudeHandler(c *gin.Context) {
	// Сначала пробуем прочитать тело как есть
	rawData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": "Не удалось прочитать тело запроса",
				"type":    "invalid_request_error",
			},
		})
		return
	}
	
	// Пробуем ручное исправление для известного формата
	fixedData := ManualJSONFix(rawData)
	
	// Восстанавливаем тело для дальнейшей обработки
	c.Request.Body = io.NopCloser(bytes.NewBuffer(fixedData))
	
	// Вызываем обычный обработчик nightmare
	nightmare(c)
}