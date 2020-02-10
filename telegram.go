package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func notification(msg string) error {

	Log.Println("Sending telegram notification..")
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", os.Getenv("TELEGRAM_BOT_KEY"))

	j, err := json.Marshal(map[string]string{"chat_id": os.Getenv("TELEGRAM_CHAT_ID"),
		"text": msg})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		Log.Println("Failure sending telegram notification..")
		Log.Println(err.Error())
		return err
	}
	defer resp.Body.Close()

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	//_, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("response Body:", string(body))
	Log.Println("Succesfully sent telegram notification..")
	return nil
}
