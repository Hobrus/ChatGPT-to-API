package main

import (
	"bytes"
	"encoding/json"
	chatgpt_request_converter "freechatgpt/conversion/requests/chatgpt"
	chatgpt "freechatgpt/internal/chatgpt"
	"freechatgpt/internal/tokens"
	official_types "freechatgpt/typings/official"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	uuidNamespace = uuid.MustParse("12345678-1234-5678-1234-567812345678")
)

// Model structure for the /api/v0/models endpoint
type ModelV0 struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Type              string `json:"type"`
	Publisher         string `json:"publisher"`
	Arch              string `json:"arch"`
	CompatibilityType string `json:"compatibility_type"`
	Quantization      string `json:"quantization"`
	State             string `json:"state"`
	MaxContextLength  int    `json:"max_context_length"`
}

// Model data store
var modelsV0 = []ModelV0{
	{
		ID:                "gpt-4o-mini",
		Object:            "model",
		Type:              "llm",
		Publisher:         "openai",
		Arch:              "gpt",
		CompatibilityType: "openai",
		Quantization:      "none",
		State:             "loaded",
		MaxContextLength:  128000,
	},
	{
		ID:                "gpt-4o",
		Object:            "model",
		Type:              "llm",
		Publisher:         "openai",
		Arch:              "gpt",
		CompatibilityType: "openai",
		Quantization:      "none",
		State:             "loaded",
		MaxContextLength:  128000,
	},
	{
		ID:                "gpt-4",
		Object:            "model",
		Type:              "llm",
		Publisher:         "openai",
		Arch:              "gpt",
		CompatibilityType: "openai",
		Quantization:      "none",
		State:             "loaded",
		MaxContextLength:  128000,
	},
	{
		ID:                "o1-pro",
		Object:            "model",
		Type:              "llm",
		Publisher:         "anthropic",
		Arch:              "claude",
		CompatibilityType: "anthropic",
		Quantization:      "none",
		State:             "loaded",
		MaxContextLength:  128000,
	},
	{
		ID:                "o1-mini",
		Object:            "model",
		Type:              "llm",
		Publisher:         "anthropic",
		Arch:              "claude",
		CompatibilityType: "anthropic",
		Quantization:      "none",
		State:             "loaded",
		MaxContextLength:  128000,
	},
	{
		ID:                "o1",
		Object:            "model",
		Type:              "llm",
		Publisher:         "anthropic",
		Arch:              "claude",
		CompatibilityType: "anthropic",
		Quantization:      "none",
		State:             "loaded",
		MaxContextLength:  128000,
	},
	{
		ID:                "o3",
		Object:            "model",
		Type:              "llm",
		Publisher:         "anthropic",
		Arch:              "claude",
		CompatibilityType: "anthropic",
		Quantization:      "none",
		State:             "loaded",
		MaxContextLength:  128000,
	},
	{
		ID:                "o4-mini-high",
		Object:            "model",
		Type:              "llm",
		Publisher:         "anthropic",
		Arch:              "claude",
		CompatibilityType: "anthropic",
		Quantization:      "none",
		State:             "loaded",
		MaxContextLength:  128000,
	},
}

// Handler for /api/v0/models
func getModelsV0(c *gin.Context) {
	c.JSON(200, gin.H{
		"object": "list",
		"data":   modelsV0,
	})
}

// Handler for /api/v0/models/:model
func getModelV0(c *gin.Context) {
	modelID := c.Param("model")

	for _, model := range modelsV0 {
		if model.ID == modelID {
			c.JSON(200, model)
			return
		}
	}

	c.JSON(404, gin.H{"error": "Model not found"})
}

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
            {
                "id":       "o1-pro",
                "object":   "model",
                "created":  1688888888,
                "owned_by": "chatgpt-to-api",
            },
            {
                "id":       "o3",
                "object":   "model",
                "created":  1688888888,
                "owned_by": "chatgpt-to-api",
            },
            {
                "id":       "o4-mini-high",
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
    // Read the raw request body for better debugging
    jsonData, err := io.ReadAll(c.Request.Body)
    if err != nil {
        log.Printf("Error reading request body: %v", err)
        c.JSON(400, gin.H{"error": gin.H{
            "message": "Could not read request body",
            "type":    "invalid_request_error",
            "param":   nil,
            "code":    err.Error(),
        }})
        return
    }

    // Restore body for further processing
    c.Request.Body = io.NopCloser(bytes.NewBuffer(jsonData))

    // Log raw request for debugging
    log.Printf("Raw JSON request: %s", string(jsonData))

    // Manually parse the JSON to have more control
    var original_request official_types.APIRequest
    decoder := json.NewDecoder(bytes.NewBuffer(jsonData))
    decoder.DisallowUnknownFields() // Строгая проверка полей

    if err := decoder.Decode(&original_request); err != nil {
        log.Printf("Error decoding JSON: %v", err)
        c.JSON(400, gin.H{"error": gin.H{
            "message": "Request must be proper JSON",
            "type":    "invalid_request_error",
            "param":   nil,
            "code":    err.Error(),
        }})
        return
    }

    // Log the parsed request
    log.Printf("Parsed request: %+v", original_request)

    // Special handling for Claude models
    if strings.HasPrefix(original_request.Model, "o1") ||
       strings.HasPrefix(original_request.Model, "o3") ||
       strings.HasPrefix(original_request.Model, "o4") {
        log.Printf("Detected Claude model: %s", original_request.Model)
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

    response, err := chatgpt.POSTconversation(translated_request, &secret, deviceId, chat_require.Token, proofToken, turnstileToken, proxy_url)
    if err != nil {
        log.Printf("Error sending request: %v", err)
        c.JSON(500, gin.H{
            "error": "error sending request",
        })
        return
    }
    defer response.Body.Close()
    if chatgpt.Handle_request_error(c, response) {
        return
    }
    var full_response string
    for i := 3; i > 0; i-- {
        var continue_info *chatgpt.ContinueInfo
        var response_part string
        response_part, continue_info = chatgpt.Handler(c, response, &secret, proxy_url, deviceId, uid, original_request.Stream)
        full_response += response_part
        if continue_info == nil {
            break
        }
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
        response, err = chatgpt.POSTconversation(translated_request, &secret, deviceId, chat_require.Token, proofToken, turnstileToken, proxy_url)
        if err != nil {
            c.JSON(500, gin.H{
                "error": "error sending request",
            })
            return
        }
        defer response.Body.Close()
        if chatgpt.Handle_request_error(c, response) {
            return
        }
    }
    if c.Writer.Status() != 200 {
        return
    }
    if !original_request.Stream {
        c.JSON(200, official_types.NewChatCompletion(full_response))
    } else {
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