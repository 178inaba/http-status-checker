package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/slack-go/slack"
	"gopkg.in/yaml.v2"
)

func main() {
	var count int
	var timeout time.Duration
	var webhookRawurl string

	flag.IntVar(&count, "count", 1, "Count.")
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "Timeout.")
	flag.StringVar(&webhookRawurl, "webhook", "", "Slack webhook URL.")

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("You must specify a target url.")
	}
	targetRawurl := args[0]

	f, err := os.Open("header.yaml")
	if err != nil {
		log.Fatal(err)
	}

	headers := map[string]string{}
	if err = yaml.NewDecoder(f).Decode(&headers); err != nil {
		log.Fatal(err)
	}

	if _, err := url.ParseRequestURI(targetRawurl); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	req, err := http.NewRequest(http.MethodHead, targetRawurl, nil)
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	var i int
	for {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		resp, err := http.DefaultClient.Do(req.Clone(ctx))
		if err != nil {
			i = 0
			log.Print(err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			i++
			if i < count {
				log.Printf("OK count: %d.", i)
				continue
			}

			cancel()
			break
		}

		i = 0
		log.Print(resp.Status)
	}

	msg := fmt.Sprintf("%s: OK!", targetRawurl)
	log.Print(msg)
	if webhookRawurl != "" {
		if err := slack.PostWebhookContext(ctx, webhookRawurl, &slack.WebhookMessage{Text: msg}); err != nil {
			log.Fatal(err)
		}
	}
}
