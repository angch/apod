package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ApodPost struct {
	Text     string
	ImageURL string
}

func doYMD(y, m, d int) *ApodPost {
	if y > 2000 {
		// Perlism
		y -= 2000
	}

	url := fmt.Sprintf("https://apod.nasa.gov/apod/ap%02d%02d%02d.html", y, m, d)

	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.Status != "200" {
		log.Fatal("Not ready yet")
	}
	if true {
		log.Fatal("foom")
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	imgUrl := ""
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		img := s.AttrOr("href", "")
		if strings.HasSuffix(img, ".jpg") && strings.HasPrefix(img, "image") {
			imgUrl = "https://apod.nasa.gov/apod/" + img
		}
	})
	title := ""
	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		title = strings.TrimSpace(s.Text())
	})
	if title != "" && imgUrl != "" {
		post := ApodPost{
			Text:     title,
			ImageURL: imgUrl,
		}
		return &post

	}
	return nil
}

func postSlack(post *ApodPost) {
	text := fmt.Sprintf("%s %s", post.Text, post.ImageURL)
	content := bytes.NewBuffer([]byte(fmt.Sprintf("{\"text\":\"%s\"}", text)))
	_, err := http.Post(slackUrl, "Content-type: application/json", content)
	if err != nil {
		log.Fatal(err)
	}
}

var slackUrl string

func main() {
	slackUrl = os.Getenv("SLACKWEBHOOK")
	if slackUrl == "" {
		log.Fatal("Env SLACKWEBHOOK required, e.g. https://hooks.slack.com/services/z/y/x")
	}
	today := time.Now()
	d := today.Day()
	m := int(today.Month())
	y := today.Year()

	p := doYMD(y, m, d)
	log.Printf("%+v\n", p)
	if p != nil {
		postSlack(p)
	}
}
