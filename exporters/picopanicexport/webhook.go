package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const payload = `
{
  "username": "Webhook",
  "avatar_url": "https://cdn.pixabay.com/photo/2016/05/15/12/05/panic-1393619_1280.png",
  "content": "Panic occurred!",
  "embeds": [
    {
      "author": {
        "name": "Kids",
        "icon_url": "https://i.fbcd.co/products/resized/resized-750-500/s211206-kids-avatar-mainpreview-575cd7e35d16d467063db0c6288b5a1a0ede3e0d5a1fe02c2a6cf02420f9fd38.jpg"
      },
      "title": "PANICOOOOO",
      "description": "Someone activated the panic button!",
      "color": 15224655,
      "fields": [
        {
          "name": "Time Ago",
          "value": "%s",
          "inline": true
        },
        {
          "name": "Actions",
          "value": "Get in touch. Get back home",
          "inline": true
        }
      ],
      "footer": {
        "text": "No panic!! Can be a false alarm! :smirk:",
        "icon_url": "https://i.imgur.com/fKL31aD.jpg"
      }
    }
  ]
}
`

const webhook = "https://discord.com/api/webhooks/1288571537737256971/nf6I8ssSXOSkYYVGs0YFbd00QSwmOC-opJx5Y88D-QeTZ4CEmAiXpA5NL16DHs2eDD2f"

func triggerWebhook(panic *panicCheck) error {
	timeAgo := time.Since(panic.Timestamp)
	slog.Info("panic occurred", "panic", panic.Panic, "timeAgo", timeAgo)

	formattedPayload := fmt.Sprintf(payload, timeAgo)
	reader := bytes.NewReader([]byte(formattedPayload))
	req, err := http.NewRequest("POST", webhook, reader)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return fmt.Errorf("error creating http request %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing http request %w", err)
	}

	defer res.Body.Close()
	body := new(bytes.Buffer)
	body.ReadFrom(res.Body)

	if res.StatusCode > 299 {
		return fmt.Errorf("webhook failed with status %v and body %s", res.StatusCode, body.String())
	}
	return nil
}
