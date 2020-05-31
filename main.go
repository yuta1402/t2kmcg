package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type RequestData struct {
	Text string `json:"text"`
}

func correctTime(t time.Time) time.Time {
	t = t.Add(time.Duration(4-((4+t.Minute())%5)) * time.Minute)
	return t
}

func makeStartTime(startWeekday int, startTimeStr string) (time.Time, error) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Time{}, err
	}

	now := time.Now().In(jst)

	diffDay := (startWeekday - int(now.Weekday()) + 7) % 7
	startDay := now.Day() + diffDay

	a := strings.Split(startTimeStr, ":")
	if len(a) != 2 {
		return time.Time{}, errors.New("start time parse error")
	}

	startHour, err := strconv.Atoi(a[0])
	if err != nil {
		return time.Time{}, err
	}

	startMin, err := strconv.Atoi(a[1])
	if err != nil {
		return time.Time{}, err
	}

	startTime := time.Date(now.Year(), now.Month(), startDay, startHour, startMin, 0, 0, jst)

	if startTime.Unix() < now.Unix() {
		return startTime.AddDate(0, 0, 7), nil
	}

	return correctTime(startTime), nil
}

func postSlack(createdContest *CreatedContest, apiURL string) (*http.Response, error) {
	startTimeStr := createdContest.Options.StartTime.Format("2006/01/02 15:04")
	endTimeStr := createdContest.Options.EndTime.Format("2006/01/02 15:04")

	text := "*「" + createdContest.Options.ContestTitle + "」開催のお知らせ*\n" +
		"日時: " + startTimeStr + " ~ " + endTimeStr + "\n" +
		"会場: " + createdContest.URL + "\n" +
		"\n参加できる方は:ok: 絵文字、参加できない方は:ng: 絵文字でお知らせください。"

	d := RequestData{
		Text: text,
	}

	json, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	res, err := http.Post(apiURL, "application/json", bytes.NewBuffer([]byte(json)))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func main() {
	var (
		id           string
		password     string
		titlePrefix  string
		description  string
		startWeekday int
		startTimeStr string
		durationMin  int
		apiURL       string
		isDry        bool
	)

	flag.StringVar(&id, "id", "", "id of atcoder virtual contest")
	flag.StringVar(&password, "password", "", "password of atcoder virtual contest")
	flag.StringVar(&titlePrefix, "title-prefix", "tmp contest", "prefix of contest title")
	flag.StringVar(&description, "description", "", "contest description")
	flag.IntVar(&startWeekday, "start-weekday", 0, "start weekday Sun=0, Mon=1, ... (default 0)")
	flag.StringVar(&startTimeStr, "start-time", "21:00", "start time")
	flag.IntVar(&durationMin, "duration", 100, "duration [min]")
	flag.StringVar(&apiURL, "api", "", "API of slack")
	flag.BoolVar(&isDry, "dry-run", false, "enable dry run mode")

	acpPage, err := NewAtCoderProblemsPage()
	if err != nil {
		log.Fatal(err)
	}
	defer acpPage.Close()

	flag.VisitAll(func(f *flag.Flag) {
		n := "T2KMPG_" + strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		if s := os.Getenv(n); s != "" {
			f.Value.Set(s)
		}
	})

	flag.Parse()

	if !isDry {
		if err := acpPage.Login(id, password); err != nil {
			log.Fatal(err)
		}
	}

	startTime, err := makeStartTime(startWeekday, startTimeStr)
	if err != nil {
		log.Fatal(err)
	}

	endTime := startTime.Add(time.Duration(durationMin) * time.Minute)

	options := ContestOptions{
		ContestTitle: titlePrefix,
		Description:  description,
		StartTime:    startTime,
		EndTime:      endTime,
	}

	createdContest := &CreatedContest{
		Options: options,
		URL:     "",
	}

	if !isDry {
		createdContest, err = acpPage.CreateContest(options)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Created contest:")
	fmt.Println("  ContestTitle: " + createdContest.Options.ContestTitle)
	fmt.Println("  Description: " + createdContest.Options.Description)
	fmt.Println("  StartTime: " + createdContest.Options.StartTime.Format("2006/01/02 15:04"))
	fmt.Println("  EndTime: " + createdContest.Options.EndTime.Format("2006/01/02 15:04"))
	fmt.Println("  URL: " + createdContest.URL)

	if !isDry {
		res, err := postSlack(createdContest, apiURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			return
		}

		fmt.Println("Status :" + res.Status)
	}
}
