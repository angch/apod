package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("Not ready yet ", resp.Status)
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
	posts := make(map[string]ApodPost)
	f, err := os.Open("posts.js")
	if err == nil {
		b, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(b, &posts)
		if err != nil {
			log.Fatal(err)
		}
		f.Close()
	}

	today := time.Now() // Yes, I know. timezone.
	d := today.Day()
	m := int(today.Month())
	y := today.Year()
	key := fmt.Sprintf("%04d%02d%02d", y, m, d)
	_, exists := posts[key]
	if exists {
		log.Println("Done for today")
		return
	}

	p := doYMD(y, m, d)
	if p != nil {
		posts[key] = *p
		f, err := os.OpenFile("posts.js", os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			log.Fatal(err)
		}
		b, _ := json.Marshal(posts)
		_, err = f.Write(b)
		if err != nil {
			log.Fatal(err)
		}
		f.Close()

		log.Printf("%+v\n", p)

		if true {
			postSlack(p)
		}
	}
}
