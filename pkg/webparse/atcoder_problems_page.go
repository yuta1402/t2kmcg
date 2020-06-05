package webparse

import (
	"errors"
	"fmt"
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
		"--window-size=1200,800",
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

	{
		elementValues := []ContestOptionElementValue{
			{"#root > div > div.my-5.container > div:nth-child(2) > div > input", options.ContestTitle},
			{"#root > div > div.my-5.container > div:nth-child(3) > div > input", options.Description},
			{"#root > div > div.my-5.container > div:nth-child(5) > div > div > input", startDate},
			{"#root > div > div.my-5.container > div:nth-child(6) > div > div > input", endDate},
		}

		for _, ev := range elementValues {
			e := p.Find(ev.Selector)
			if err := e.Fill(ev.Value); err != nil {
				return nil, err
			}
		}
	}

	{
		elementValues := []ContestOptionElementValue{
			{"#root > div > div.my-5.container > div:nth-child(5) > div > div > select:nth-child(2)", startHour},
			{"#root > div > div.my-5.container > div:nth-child(5) > div > div > select:nth-child(3)", startMinute},
			{"#root > div > div.my-5.container > div:nth-child(6) > div > div > select:nth-child(2)", endHour},
			{"#root > div > div.my-5.container > div:nth-child(6) > div > div > select:nth-child(3)", endMinute},
		}

		for _, ev := range elementValues {
			e := p.Find(ev.Selector)
			if err := e.Select(ev.Value); err != nil {
				return nil, err
			}
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

	createdContest := &CreatedContest{
		Options: options,
		URL:     url,
	}

	return createdContest, nil
}
