package chatgpt

import (
	chatgpt_types "freechatgpt/internal/chatgpt"
	"freechatgpt/internal/tokens"
	official_types "freechatgpt/typings/official"
	"regexp"
	"strings"
)

var gptsRegexp = regexp.MustCompile(`-gizmo-g-(\w+)`)

func ConvertAPIRequest(api_request official_types.APIRequest, account string, secret *tokens.Secret, deviceId string, proxy string) chatgpt_types.ChatGPTRequest {
    chatgpt_request := chatgpt_types.NewChatGPTRequest()

    if strings.HasPrefix(api_request.Model, "gpt-4o-mini") || strings.HasPrefix(api_request.Model, "gpt-3.5") {
        chatgpt_request.Model = "gpt-4o-mini"
    } else if strings.HasPrefix(api_request.Model, "gpt-4o") {
        chatgpt_request.Model = "gpt-4o"
    } else if strings.HasPrefix(api_request.Model, "gpt-4") {
        chatgpt_request.Model = "gpt-4"
    // >>> Новый блок <<<
    } else if strings.HasPrefix(api_request.Model, "o1-pro") {
        chatgpt_request.Model = "o1-pro"
    // >>> конец вставки <<<
    } else if strings.HasPrefix(api_request.Model, "o1-mini") {
        chatgpt_request.Model = "o1-mini"
    } else if strings.HasPrefix(api_request.Model, "o1") {
        chatgpt_request.Model = "o1"
    } else if strings.HasPrefix(api_request.Model, "o3") {
        chatgpt_request.Model = "o3"
    } else if strings.HasPrefix(api_request.Model, "o4-mini-high") {
        chatgpt_request.Model = "o4-mini-high"
    }

    // Ниже может идти проверка на gizmo...
    // (если она вам нужна, оставляете как есть)
    matches := gptsRegexp.FindStringSubmatch(api_request.Model)
    if len(matches) == 2 {
        chatgpt_request.ConversationMode.Kind = "gizmo_interaction"
        chatgpt_request.ConversationMode.GizmoId = "g-" + matches[1]
    }

    // Если у вас PLUS аккаунт, то проверка на secret.TeamUserID
    // идёт при отправке. Здесь важна именно модель.

    // Пробегаемся по всем сообщением user/system
    // и добавляем в ChatGPTRequest
    ifMultimodel := secret.Token != ""
    for _, api_message := range api_request.Messages {
        if api_message.Role == "system" {
            api_message.Role = "critic"
        }
        chatgpt_request.AddMessage(api_message.Role, api_message.Content, ifMultimodel, account, secret, deviceId, proxy)
    }
    return chatgpt_request
}


func ConvertTTSAPIRequest(input string) chatgpt_types.ChatGPTRequest {
	chatgpt_request := chatgpt_types.NewChatGPTRequest()
	chatgpt_request.HistoryAndTrainingDisabled = false
	chatgpt_request.AddAssistantMessage(input)
	return chatgpt_request
}