package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type Notification struct {
	msg     string
	chat_id int
}

func sendNotification(n Notification) {
	if n.chat_id == 0 {
		Log.Println("Not sending notification with ChatId 0")
		return
	}

	Log.Println("Sending telegram notification..")
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", os.Getenv("TELEGRAM_BOT_KEY"))

	j, err := json.Marshal(map[string]string{"chat_id": strconv.Itoa(n.chat_id), "text": n.msg})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		Log.Println("Failure sending telegram notification..")
		Log.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	Log.Println("Succesfully sent telegram notification..")
	return
}

func processNotifications() {
	for msg := range notifications {
		sendNotification(msg)
	}
}
