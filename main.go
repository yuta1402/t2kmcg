package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sclevine/agouti"
)

const (
	AtCoderProblemsEndpoint = "https://kenkoooo.com/atcoder/#"
)

type AtCoderProblemsPage struct {
	driver *agouti.WebDriver
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

	p := &AtCoderProblemsPage{
		driver: driver,
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

func (acpPage *AtCoderProblemsPage) NewPage(urlPath string) (*agouti.Page, error) {
	p, err := acpPage.driver.NewPage()
	if err != nil {
		return nil, err
	}

	if err := navigateWithPath(p, urlPath); err != nil {
		return nil, err
	}

	return p, nil
}

func main() {
	acpPage, err := NewAtCoderProblemsPage()
	if err != nil {
		log.Fatal(err)
	}
	defer acpPage.Close()

	p, err := acpPage.NewPage("")

	time.Sleep(1 * time.Second)

	{
		e := p.FindByLink("Login")
		if err := e.Click(); err != nil {
			log.Fatal(err)
		}
	}

	time.Sleep(1 * time.Second)

	{
		id := os.Getenv("T2KMPG_ID")

		e := p.FindByID("login_field")
		if err := e.Fill(id); err != nil {
			log.Fatal(err)
		}

		fmt.Println(id)
	}

	{
		password := os.Getenv("T2KMPG_PASSWORD")

		e := p.FindByID("password")
		if err := e.Fill(password); err != nil {
			log.Fatal(err)
		}

		fmt.Println(password)
	}

	{
		e := p.FindByName("commit")
		if err := e.Submit(); err != nil {
			log.Fatal(err)
		}
	}

	time.Sleep(1 * time.Second)

	if err := navigateWithPath(p, "/contest/create"); err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	{
		e := p.Find("#root > div > div.my-5.container > div:nth-child(2) > div > input")
		e.Fill("test contest")
	}

	{
		e := p.Find("#root > div > div.my-5.container > div:nth-child(3) > div > input")
		e.Fill("this is test contest.")
	}

	{
		e := p.FindByButton("Add")
		e.Click()
	}

	time.Sleep(100 * time.Millisecond)

	p.Screenshot("test.png")

	fmt.Println(p.HTML())
}
