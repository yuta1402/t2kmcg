package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sclevine/agouti"
)

const (
	AtCoderProblemsEndpoint = "https://kenkoooo.com/atcoder/#"
	SleepInterval           = 100 * time.Millisecond
)

type AtCoderProblemsPage struct {
	driver *agouti.WebDriver
	page   *agouti.Page
}

type ContestOptions struct {
	ContestTitle string
	Description  string
	StartTime    time.Time
	EndTime      time.Time
}

type ContestOptionElementValue struct {
	Selector string
	Value    string
}

type CreatedContest struct {
	Options ContestOptions
	URL     string
}

func NewAtCoderProblemsPage() (*AtCoderProblemsPage, error) {
	options := agouti.ChromeOptions("args", []string{
		"--headless",
		// "--window-size=1200,800",
		"--window-size=1200,2000",
		"--blink-settings=imagesEnabled=false", // don't load images
		"--disable-gpu",                        // ref: https://developers.google.com/web/updates/2017/04/headless-chrome#cli
		"no-sandbox",                           // ref: https://github.com/theintern/intern/issues/878
		"disable-dev-shm-usage",                // ref: https://qiita.com/yoshi10321/items/8b7e6ed2c2c15c3344c6
	})

	driver := agouti.ChromeDriver(options)

	if err := driver.Start(); err != nil {
		return nil, err
	}

	page, err := driver.NewPage()
	if err != nil {
		return nil, err
	}

	p := &AtCoderProblemsPage{
		driver: driver,
		page:   page,
	}

	return p, nil
}

func (acpPage *AtCoderProblemsPage) Close() {
	acpPage.driver.Stop()
}

func navigateWithPath(page *agouti.Page, urlPath string) error {
	// TODO: join path
	if err := page.Navigate(AtCoderProblemsEndpoint + urlPath); err != nil {
		return err
	}

	return nil
}

func (acpPage *AtCoderProblemsPage) Login(id string, password string) error {
	p := acpPage.page
	if err := navigateWithPath(p, ""); err != nil {
		return err
	}

	time.Sleep(SleepInterval)

	{
		e := p.FindByLink("Login")
		if err := e.Click(); err != nil {
			return err
		}
	}

	time.Sleep(SleepInterval)

	{
		e := p.FindByID("login_field")
		if err := e.Fill(id); err != nil {
			return err
		}
	}

	{
		e := p.FindByID("password")
		if err := e.Fill(password); err != nil {
			return err
		}
	}

	{
		e := p.FindByName("commit")
		if err := e.Submit(); err != nil {
			return err
		}

		time.Sleep(SleepInterval)

		url, err := p.URL()
		if err != nil {
			return err
		}

		// githubのsessionページに戻されてしまった場合はログイン失敗
		if url == "https://github.com/session" {
			return errors.New("failed to login")
		}
	}

	return nil
}

func correctTime(t time.Time) time.Time {
	t = t.Add(time.Duration(4-((4+t.Minute())%5)) * time.Minute)
	return t
}

func makeDateStr(t time.Time) string {
	y, m, d := t.Date()
	return fmt.Sprintf("%02d/%02d/%04d/", m, d, y)
}

func makeDateHourMinute(t time.Time) (string, string, string) {
	d := makeDateStr(t)
	h := strconv.Itoa(t.Hour())
	m := strconv.Itoa(t.Minute())
	return d, h, m
}

func (acpPage *AtCoderProblemsPage) CreateContest(options ContestOptions) (*CreatedContest, error) {
	p := acpPage.page
	if err := navigateWithPath(p, "/contest/create"); err != nil {
		return nil, err
	}

	time.Sleep(SleepInterval)

	startDate, startHour, startMinute := makeDateHourMinute(options.StartTime)
	endDate, endHour, endMinute := makeDateHourMinute(options.EndTime)

	elementValues := []ContestOptionElementValue{
		{"#root > div > div.my-5.container > div:nth-child(2) > div > input", options.ContestTitle},
		{"#root > div > div.my-5.container > div:nth-child(3) > div > input", options.Description},
		{"#root > div > div.my-5.container > div:nth-child(5) > div > div > input", startDate},
		{"#root > div > div.my-5.container > div:nth-child(5) > div > div > select:nth-child(2)", startHour},
		{"#root > div > div.my-5.container > div:nth-child(5) > div > div > select:nth-child(3)", startMinute},
		{"#root > div > div.my-5.container > div:nth-child(6) > div > div > input", endDate},
		{"#root > div > div.my-5.container > div:nth-child(6) > div > div > select:nth-child(2)", endHour},
		{"#root > div > div.my-5.container > div:nth-child(6) > div > div > select:nth-child(3)", endMinute},
	}

	for _, ev := range elementValues {
		e := p.Find(ev.Selector)
		if err := e.Fill(ev.Value); err != nil {
			return nil, err
		}
	}

	{
		e := p.FindByButton("Add")
		if err := e.Click(); err != nil {
			return nil, err
		}
	}

	time.Sleep(5 * SleepInterval)

	{
		e := p.FindByButton("Create Contest")
		if err := e.Click(); err != nil {
			return nil, err
		}
	}

	time.Sleep(10 * SleepInterval)

	url, err := p.URL()
	if err != nil {
		return nil, err
	}

	// /contest/showに飛んでいなかったらコンテスト作成失敗
	if !strings.Contains(url, AtCoderProblemsEndpoint+"/contest/show") {
		return nil, errors.New("failed to create contest")
	}

	return nil, nil
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
		namePrefix   string
		description  string
		startWeekday int
		startTimeStr string
		durationMin  int
		apiURL       string
		isDry        bool
	)

	flag.StringVar(&id, "id", "", "id of atcoder virtual contest")
	flag.StringVar(&password, "password", "", "password of atcoder virtual contest")
	flag.StringVar(&namePrefix, "name-prefix", "tmp contest", "prefix of contest name")
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
		ContestTitle: "test contest",
		Description:  "this is test contest.",
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
}
