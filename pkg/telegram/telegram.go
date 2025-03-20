package telegram

import (
	"bytes"
	"encoding/json"
	"exapp-go/config"
	"net/http"
)

func SendMsg(msg string) (err error) {
	botToken := config.Conf().TelegramBot.Token
	chatID := config.Conf().TelegramBot.ChatID
	if botToken == "" || chatID == "" {
		return
	}
	apiURL := "https://api.telegram.org/bot" + botToken + "/sendMessage"

	requestBody, err := json.Marshal(map[string]string{
		"chat_id": chatID,
		"text":    msg,
	})
	if err != nil {
		return
	}
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return

}
