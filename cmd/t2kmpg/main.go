package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yuta1402/t2kmpg/pkg/slack"
	"github.com/yuta1402/t2kmpg/pkg/webparse"
)

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

	flag.VisitAll(func(f *flag.Flag) {
		n := "T2KMPG_" + strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		if s := os.Getenv(n); s != "" {
			f.Value.Set(s)
		}
	})

	flag.Parse()

	ilog := log.New(os.Stdout, "[info] ", log.LstdFlags|log.LUTC)
	elog := log.New(os.Stderr, "[error] ", log.LstdFlags|log.LUTC)

	acpPage, err := webparse.NewAtCoderProblemsPage()
	if err != nil {
		elog.Fatal(err)
	}
	defer acpPage.Close()

	if !isDry {
		ilog.Println("Trying to login to AtCoder Problems...")

		if err := acpPage.Login(id, password); err != nil {
			elog.Fatal(err)
		}

		ilog.Println("Login succeeded.")
	}

	ilog.Println("Creating contet...")

	startTime, err := makeStartTime(startWeekday, startTimeStr)
	if err != nil {
		elog.Fatal(err)
	}

	endTime := startTime.Add(time.Duration(durationMin) * time.Minute)

	options := webparse.ContestOptions{
		ContestTitle: titlePrefix,
		Description:  description,
		StartTime:    startTime,
		EndTime:      endTime,
	}

	createdContest := &webparse.CreatedContest{
		Options: options,
		URL:     "",
	}

	if !isDry {
		createdContest, err = acpPage.CreateContest(options)
		if err != nil {
			log.Fatal(err)
		}
	}

	ilog.Println("Created contest:")
	ilog.Println("  ContestTitle: " + createdContest.Options.ContestTitle)
	ilog.Println("  Description: " + createdContest.Options.Description)
	ilog.Println("  StartTime: " + createdContest.Options.StartTime.Format("2006/01/02 15:04"))
	ilog.Println("  EndTime: " + createdContest.Options.EndTime.Format("2006/01/02 15:04"))
	ilog.Println("  URL: " + createdContest.URL)

	ilog.Println("Posting to slack...")

	if !isDry {
		res, err := slack.PostContestAnnouncement(createdContest, apiURL)
		if err != nil {
			elog.Fatal(err)
		}

		ilog.Println("Status :" + res.Status)
	}
}
