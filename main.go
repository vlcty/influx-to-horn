/*
MIT License

Copyright (c) 2021 Josef 'veloc1ty' Stautner

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

type HornMessage struct {
	Text string `json:"text"`
}

type InfluxMessage struct {
	Message string `json:"_message"`
}

func main() {
	log.Info("Started influx-to-horn")

	if len(os.Getenv("HOOKID")) == 0 {
		log.Fatal("No environment Variable HOOKID found")
	}

	http.HandleFunc("/send", func(response http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()

		if request.Method != http.MethodPost {
			http.Error(response, "Only post is allowed", http.StatusForbidden)
		} else {
			message := &InfluxMessage{}
			decodeErr := json.NewDecoder(request.Body).Decode(message)

			if decodeErr != nil {
				log.WithField("Error", decodeErr.Error()).Error("Request decode error")
				return
			}

			log.Info("Message: " + message.Message)

			newMessage := &bytes.Buffer{}
			json.NewEncoder(newMessage).Encode(&HornMessage{Text: message.Message})

			response, err := http.Post("https://integram.org/webhook/"+os.Getenv("HOOKID"), "application/json", newMessage)

			if err != nil {
				log.WithField("Error", err.Error()).Error("Error posting to integram")
			} else if response.StatusCode != http.StatusOK {
				log.WithField("Status", response.Status).Error("Did not receive a HTTP 200")
			} else {
				log.WithField("Status", response.Status).Info("Delivered message")
			}
		}
	})

	log.Fatal(http.ListenAndServe(":2456", nil))
}
