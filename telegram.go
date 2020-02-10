package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func notification(msg string, output io.Writer) error {

	output.Write([]byte("Sending telegram notification..\n"))
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", os.Getenv("TELEGRAM_BOT_KEY"))

	j, err := json.Marshal(map[string]string{"chat_id": os.Getenv("TELEGRAM_CHAT_ID"),
		"text": msg})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		output.Write([]byte("Failure sending telegram notification..\n"))
		output.Write([]byte(err.Error()))
		return err
	}
	defer resp.Body.Close()

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	//_, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("response Body:", string(body))
	output.Write([]byte("Succesfully sent telegram notification..\n"))
	return nil
}
