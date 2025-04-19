package main

import (
	"encoding/json"
	"io"
	chatgpt_request_converter "freechatgpt/conversion/requests/chatgpt"
	chatgpt "freechatgpt/internal/chatgpt"
	"freechatgpt/internal/tokens"
	official_types "freechatgpt/typings/official"
	"os"

	http "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)
var (
	uuidNamespace = uuid.MustParse("12345678-1234-5678-1234-567812345678")
)

func passwordHandler(c *gin.Context) {
	// Get the password from the request (json) and update the password
	type password_struct struct {
		Password string `json:"password"`
	}
	var password password_struct
	err := c.BindJSON(&password)
	if err != nil {
		c.String(400, "password not provided")
		return
	}
	ADMIN_PASSWORD = password.Password
	// Set environment variable
	os.Setenv("ADMIN_PASSWORD", ADMIN_PASSWORD)
	c.String(200, "password updated")
}

func tokensHandler(c *gin.Context) {
	// Get the request_tokens from the request (json) and update the request_tokens
	var request_tokens map[string]tokens.Secret
	err := c.BindJSON(&request_tokens)
	if err != nil {
		c.String(400, "tokens not provided")
		return
	}
	ACCESS_TOKENS = tokens.NewAccessToken(request_tokens)
	ACCESS_TOKENS.Save()
	validAccounts = ACCESS_TOKENS.GetKeys()
	c.String(200, "tokens updated")
}
func optionsHandler(c *gin.Context) {
	// Set headers for CORS
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST")
	c.Header("Access-Control-Allow-Headers", "*")
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func simulateModel(c *gin.Context) {
    c.JSON(200, gin.H{
        "object": "list",
        "data": []gin.H{
            {
                "id":       "gpt-3.5-turbo",
                "object":   "model",
                "created":  1688888888,
                "owned_by": "chatgpt-to-api",
            },
            {
                "id":       "gpt-4",
                "object":   "model",
                "created":  1688888888,
                "owned_by": "chatgpt-to-api",
            },
            {
                "id":       "gpt-4o",
                "object":   "model",
                "created":  1688888888,
                "owned_by": "chatgpt-to-api",
            },
            {
                "id":       "gpt-4o-mini",
                "object":   "model",
                "created":  1688888888,
                "owned_by": "chatgpt-to-api",
            },
            {
                "id":       "o1",
                "object":   "model",
                "created":  1688888888,
                "owned_by": "chatgpt-to-api",
            },
            {
                "id":       "o1-mini",
                "object":   "model",
                "created":  1688888888,
                "owned_by": "chatgpt-to-api",
            },

            // >>> Добавляем новую модель <<<
            {
                "id":       "o1-pro",
                "object":   "model",
                "created":  1688888888,
                "owned_by": "chatgpt-to-api",
            },
        },
    })
}


func generateUUID(name string) string {
	return uuid.NewSHA1(uuidNamespace, []byte(name)).String()
}
func nightmare(c *gin.Context) {
	var original_request official_types.APIRequest
	err := c.BindJSON(&original_request)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{
			"message": "Request must be proper JSON",
			"type":    "invalid_request_error",
			"param":   nil,
			"code":    err.Error(),
		}})
		return
	}

	account, secret := getSecret()
	var proxy_url string
	if len(proxies) == 0 {
		proxy_url = ""
	} else {
		proxy_url = proxies[0]
		// Push used proxy to the back of the list
		proxies = append(proxies[1:], proxies[0])
	}
	uid := uuid.NewString()
	var deviceId string
	if account == "" {
		deviceId = uid
		chatgpt.SetOAICookie(deviceId)
	} else {
		deviceId = generateUUID(account)
		chatgpt.SetOAICookie(deviceId)
	}
	chat_require, p := chatgpt.CheckRequire(&secret, deviceId, proxy_url)
	if chat_require == nil {
		c.JSON(500, gin.H{"error": "unable to check chat requirement"})
		return
	}
	var proofToken string
	if chat_require.Proof.Required {
		proofToken = chatgpt.CalcProofToken(chat_require, proxy_url)
	}
	var turnstileToken string
	if chat_require.Turnstile.Required {
		turnstileToken = chatgpt.ProcessTurnstile(chat_require.Turnstile.DX, p)
	}
	// Convert the chat request to a ChatGPT request
	translated_request := chatgpt_request_converter.ConvertAPIRequest(original_request, account, &secret, deviceId, proxy_url)

	// Установка заголовков перед отправкой данных
	if original_request.Stream {
		c.Header("Content-Type", "text/event-stream")
	} else {
		c.Header("Content-Type", "application/json")
	}

	response, err := chatgpt.POSTconversation(translated_request, &secret, deviceId, chat_require.Token, proofToken, turnstileToken, proxy_url)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "error sending request",
		})
		return
	}
	defer response.Body.Close()

	// Проверка ошибки в ответе без вызова Handle_request_error
	if response.StatusCode != 200 {
		// Читаем тело ответа
		var error_response map[string]interface{}
		err := json.NewDecoder(response.Body).Decode(&error_response)
		if err != nil {
			body, _ := io.ReadAll(response.Body)
			c.JSON(500, gin.H{"error": gin.H{
				"message": "Unknown error",
				"type":    "internal_server_error",
				"param":   nil,
				"code":    "500",
				"details": string(body),
			}})
			return
		}
		c.JSON(response.StatusCode, gin.H{"error": gin.H{
			"message": error_response["detail"],
			"type":    response.Status,
			"param":   nil,
			"code":    "error",
		}})
		return
	}

	var full_response string
	var responses []*http.Response
	responses = append(responses, response)

	// Чтение первого ответа
	var continue_info *chatgpt.ContinueInfo
	response_part, continue_info := chatgpt.Handler(c, response, &secret, proxy_url, deviceId, uid, original_request.Stream)
	full_response += response_part

	// Если нужно продолжение беседы
	for i := 0; i < 2 && continue_info != nil; i++ { // максимум 2 дополнительных запроса
		println("Continuing conversation")
		translated_request.Messages = nil
		translated_request.Action = "continue"
		translated_request.ConversationID = continue_info.ConversationID
		translated_request.ParentMessageID = continue_info.ParentID

		chat_require, _ = chatgpt.CheckRequire(&secret, deviceId, proxy_url)
		if chat_require.Proof.Required {
			proofToken = chatgpt.CalcProofToken(chat_require, proxy_url)
		}
		if chat_require.Turnstile.Required {
			turnstileToken = chatgpt.ProcessTurnstile(chat_require.Turnstile.DX, p)
		}

		next_response, err := chatgpt.POSTconversation(translated_request, &secret, deviceId, chat_require.Token, proofToken, turnstileToken, proxy_url)
		if err != nil {
			// Для stream мы уже начали отправку, просто завершаем
			if !original_request.Stream {
				c.JSON(500, gin.H{"error": "error sending continuation request"})
			}
			break
		}

		responses = append(responses, next_response)

		// Проверка ошибки без установки статуса
		if next_response.StatusCode != 200 {
			next_response.Body.Close()
			break
		}

		// Чтение следующей части ответа
		response_part, continue_info = chatgpt.Handler(c, next_response, &secret, proxy_url, deviceId, uid, original_request.Stream)
		full_response += response_part
	}

	// Закрываем все ответы кроме первого (который закрывается через defer)
	for i := 1; i < len(responses); i++ {
		responses[i].Body.Close()
	}

	// Финальный ответ для не-стримингового режима
	if !original_request.Stream {
		c.JSON(200, official_types.NewChatCompletion(full_response))
	} else if original_request.Stream {
		// Для стриминга отправляем завершающее сообщение
		c.String(200, "data: [DONE]\n\n")
	}
}

var ttsFmtMap = map[string]string{
	"mp3":  "mp3",
	"opus": "opus",
	"aac":  "aac",
	"flac": "aac",
	"wav":  "aac",
	"pcm":  "aac",
}

var ttsTypeMap = map[string]string{
	"mp3":  "audio/mpeg",
	"opus": "audio/ogg",
	"aac":  "audio/aac",
}

var ttsVoiceMap = map[string]string{
	"alloy":   "cove",
	"ash":     "fathom",
	"coral":   "vale",
	"echo":    "ember",
	"fable":   "breeze",
	"onyx":    "orbit",
	"nova":    "maple",
	"sage":    "glimmer",
	"shimmer": "juniper",
}

func tts(c *gin.Context) {
	var original_request official_types.TTSAPIRequest
	err := c.BindJSON(&original_request)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{
			"message": "Request must be proper JSON",
			"type":    "invalid_request_error",
			"param":   nil,
			"code":    err.Error(),
		}})
		return
	}

	account, secret := getSecret()
	var proxy_url string
	if len(proxies) == 0 {
		proxy_url = ""
	} else {
		proxy_url = proxies[0]
		// Push used proxy to the back of the list
		proxies = append(proxies[1:], proxies[0])
	}
	var deviceId = generateUUID(account)
	chatgpt.SetOAICookie(deviceId)
	chat_require, p := chatgpt.CheckRequire(&secret, deviceId, proxy_url)
	if chat_require == nil {
		c.JSON(500, gin.H{"error": "unable to check chat requirement"})
		return
	}
	var proofToken string
	if chat_require.Proof.Required {
		proofToken = chatgpt.CalcProofToken(chat_require, proxy_url)
	}
	var turnstileToken string
	if chat_require.Turnstile.Required {
		turnstileToken = chatgpt.ProcessTurnstile(chat_require.Turnstile.DX, p)
	}
	// Convert the chat request to a ChatGPT request
	translated_request := chatgpt_request_converter.ConvertTTSAPIRequest(original_request.Input)

	response, err := chatgpt.POSTconversation(translated_request, &secret, deviceId, chat_require.Token, proofToken, turnstileToken, proxy_url)
	if err != nil {
		c.JSON(500, gin.H{"error": "error sending request"})
		return
	}
	defer response.Body.Close()
	if chatgpt.Handle_request_error(c, response) {
		return
	}
	msgId, convId := chatgpt.HandlerTTS(response, original_request.Input)
	format := ttsFmtMap[original_request.Format]
	if format == "" {
		format = "aac"
	}
	voice := ttsVoiceMap[original_request.Voice]
	if voice == "" {
		voice = "cove"
	}
	apiUrl := "https://chatgpt.com/backend-api/synthesize?message_id=" + msgId + "&conversation_id=" + convId + "&voice=" + voice + "&format=" + format
	data := chatgpt.GetTTS(&secret, deviceId, apiUrl, proxy_url)
	if data != nil {
		c.Data(200, ttsTypeMap[format], data)
	} else {
		c.JSON(500, gin.H{"error": "synthesize error"})
	}
	chatgpt.RemoveConversation(&secret, deviceId, convId, proxy_url)
}

func stt(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		println(err.Error())
		c.JSON(400, gin.H{"error": gin.H{
			"message": "Request must has proper file",
			"type":    "invalid_request_error",
			"param":   nil,
			"code":    err.Error(),
		}})
		return
	}
	defer file.Close()
	lang := c.Request.FormValue("language")

	account, secret := getSecret()
	if account == "" {
		c.JSON(500, gin.H{"error": "Logined user only"})
		return
	}
	var proxy_url string
	if len(proxies) == 0 {
		proxy_url = ""
	} else {
		proxy_url = proxies[0]
		// Push used proxy to the back of the list
		proxies = append(proxies[1:], proxies[0])
	}
	var deviceId = generateUUID(account)
	chatgpt.SetOAICookie(deviceId)

	data := chatgpt.GetSTT(file, header, lang, &secret, deviceId, proxy_url)
	if data != nil {
		c.Data(200, "application/json", data)
	} else {
		c.JSON(500, gin.H{"error": "transcribe error"})
	}
}
