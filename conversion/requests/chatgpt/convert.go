package chatgpt

import (
	chatgpt_types "freechatgpt/internal/chatgpt"
	"freechatgpt/internal/tokens"
	official_types "freechatgpt/typings/official"
	"regexp"
	"strings"
)

var gptsRegexp = regexp.MustCompile(`-gizmo-g-(\w+)`)
var imStartRegexp = regexp.MustCompile(`<\|im_start\|>user\s*(.*?)(?:<\|im_end\|>|$)`)
var fimPrefixRegexp = regexp.MustCompile(`<fim_prefix>(.*?)<fim_suffix>`)
var startEditingRegexp = regexp.MustCompile(`<START EDITING HERE>\s*(.*?)<STOP EDITING HERE>`)

func ConvertAPIRequest(api_request official_types.APIRequest, account string, secret *tokens.Secret, deviceId string, proxy string) chatgpt_types.ChatGPTRequest {
    chatgpt_request := chatgpt_types.NewChatGPTRequest()

    // Передача action, conversation_id и parent_message_id, если они указаны
    if api_request.Action != "" {
        chatgpt_request.Action = api_request.Action
    }
    if api_request.ConversationID != "" {
        chatgpt_request.ConversationID = api_request.ConversationID
    }
    if api_request.ParentMessageID != "" {
        chatgpt_request.ParentMessageID = api_request.ParentMessageID
    }

    if strings.HasPrefix(api_request.Model, "gpt-4o-mini") || strings.HasPrefix(api_request.Model, "gpt-3.5") {
        chatgpt_request.Model = "gpt-4o-mini"
    } else if strings.HasPrefix(api_request.Model, "gpt-4o") {
        chatgpt_request.Model = "gpt-4o"
    } else if strings.HasPrefix(api_request.Model, "gpt-4") {
        chatgpt_request.Model = "gpt-4"
    } else if strings.HasPrefix(api_request.Model, "o1-pro") {
        chatgpt_request.Model = "o1-pro"
    } else if strings.HasPrefix(api_request.Model, "o1-mini") {
        chatgpt_request.Model = "o1-mini"
    } else if strings.HasPrefix(api_request.Model, "o1") {
        chatgpt_request.Model = "o1"
    } else if strings.HasPrefix(api_request.Model, "o3") {
        chatgpt_request.Model = "o3"
    }

    // Проверка на gizmo
    matches := gptsRegexp.FindStringSubmatch(api_request.Model)
    if len(matches) == 2 {
        chatgpt_request.ConversationMode.Kind = "gizmo_interaction"
        chatgpt_request.ConversationMode.GizmoId = "g-" + matches[1]
    }

    // Обработка для continue.dev с prompt вместо messages
    if api_request.Prompt != "" {
        // Check for Anthropic Claude format: <|im_start|>user ... <|im_end|>
        imStartMatches := imStartRegexp.FindStringSubmatch(api_request.Prompt)
        if len(imStartMatches) > 1 {
            // Создаем сообщение от пользователя с prompt из Claude формата
            if api_request.Action != "continue" {
                ifMultimodel := secret.Token != ""
                chatgpt_request.AddMessage("user", imStartMatches[1], ifMultimodel, account, secret, deviceId, proxy)
            }
        } else if strings.Contains(api_request.Prompt, "<fim_prefix>") && strings.Contains(api_request.Prompt, "<fim_suffix>") {
            // Handle FIM (Fill In Middle) format from Continue.dev
            if api_request.Action != "continue" {
                // Extract content between fim_prefix and fim_suffix and use as prompt
                fimMatch := fimPrefixRegexp.FindStringSubmatch(api_request.Prompt)
                if len(fimMatch) > 1 {
                    ifMultimodel := secret.Token != ""
                    chatgpt_request.AddMessage("user", fimMatch[1], ifMultimodel, account, secret, deviceId, proxy)
                } else {
                    // Fallback to using the whole prompt
                    ifMultimodel := secret.Token != ""
                    chatgpt_request.AddMessage("user", api_request.Prompt, ifMultimodel, account, secret, deviceId, proxy)
                }
            }
        } else if strings.Contains(api_request.Prompt, "<START EDITING HERE>") && strings.Contains(api_request.Prompt, "<STOP EDITING HERE>") {
            // Handle editing format
            if api_request.Action != "continue" {
                editMatch := startEditingRegexp.FindStringSubmatch(api_request.Prompt)
                if len(editMatch) > 1 {
                    // Extract the content between START EDITING HERE and STOP EDITING HERE
                    ifMultimodel := secret.Token != ""
                    // Use the whole prompt as context is important
                    chatgpt_request.AddMessage("user", api_request.Prompt, ifMultimodel, account, secret, deviceId, proxy)
                } else {
                    // Fallback to using the whole prompt
                    ifMultimodel := secret.Token != ""
                    chatgpt_request.AddMessage("user", api_request.Prompt, ifMultimodel, account, secret, deviceId, proxy)
                }
            }
        } else {
            // Standard prompt handling
            if api_request.Action != "continue" {
                ifMultimodel := secret.Token != ""
                chatgpt_request.AddMessage("user", api_request.Prompt, ifMultimodel, account, secret, deviceId, proxy)
            }
        }
    } else {
        // Пропускаем добавление сообщений для режима continue
        if api_request.Action != "continue" {
            // Пробегаемся по всем сообщением user/system
            // и добавляем в ChatGPTRequest
            ifMultimodel := secret.Token != ""
            for _, api_message := range api_request.Messages {
                if api_message.Role == "system" {
                    api_message.Role = "critic"
                }
                chatgpt_request.AddMessage(api_message.Role, api_message.Content, ifMultimodel, account, secret, deviceId, proxy)
            }
        }
    }

    return chatgpt_request
}

func ConvertTTSAPIRequest(input string) chatgpt_types.ChatGPTRequest {
	chatgpt_request := chatgpt_types.NewChatGPTRequest()
	chatgpt_request.HistoryAndTrainingDisabled = false
	chatgpt_request.AddAssistantMessage(input)
	return chatgpt_request
}