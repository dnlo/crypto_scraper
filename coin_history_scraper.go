package main

import (
	"encoding/csv"
	"log"
	"os"
	"io/ioutil"
	"strings"
	"regexp"
	"fmt"
	
	"github.com/gocolly/colly"
	"github.com/martinlindhe/notify"
)

type Conf struct {
	StartDate string
	EndDate string
	FileName string
}

func confValue(param string) string {
	sp := strings.Split(param, ":")
	val := sp[1]
	return val
}

func readConf(filePath string) (*Conf, error) {
	conf := &Conf{}
	f, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fToString := string(f)
	sp := strings.Split(fToString, "\n")

	rStartDate := regexp.MustCompile("^start_date:")
	rEndDate := regexp.MustCompile("^end_date:")
	rFileName := regexp.MustCompile("^file_name:")
	
	for _, v := range sp {
		if rStartDate.MatchString(v) {
			conf.StartDate = strings.TrimSpace(confValue(v))
		}
		if rEndDate.MatchString(v) {
			conf.EndDate = strings.TrimSpace(confValue(v))
		}
		if rFileName.MatchString(v) {
			conf.FileName = strings.TrimSpace(confValue(v))
		}
	}
	return conf, nil
}

func convertScrapeDate(date string) string {
	date = strings.Replace(date, "Jan ", "01-", 1)
	date = strings.Replace(date, "Feb ", "02-", 1)
	date = strings.Replace(date, "Mar ", "03-", 1)
	date = strings.Replace(date, "Apr ", "04-", 1)
	date = strings.Replace(date, "May ", "05-", 1)
	date = strings.Replace(date, "Jun ", "06-", 1)
	date = strings.Replace(date, "Jul ", "07-", 1)
	date = strings.Replace(date, "Aug ", "08-", 1)
	date = strings.Replace(date, "Sep ", "09-", 1)
	date = strings.Replace(date, "Oct ", "10-", 1)
	date = strings.Replace(date, "Nov ", "11-", 1)
	date = strings.Replace(date, "Dec ", "12-", 1)
	date = strings.Replace(date, ", ", "-", 1)
	
	sp := strings.Split(date, "-")
	if len(sp) > 1 {
		date = sp[2]+"-"+sp[0]+"-"+sp[1]
		return date
	} 
	return date
}


func cleanConfDate(date string) string {
	s := strings.Split(date, "-")
	date = strings.Join(s, "")
	return date
}

func scrape(conf *Conf) {
	startDate := cleanConfDate(conf.StartDate)
	endDate := cleanConfDate(conf.EndDate)
	
	fName := conf.FileName
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()


	// Write CSV header
	writer.Write([]string{"Name", "Symbol", "Open", "High", "Low", "Close", "Volume", "Market Cap"})

	// Instantiate default collector
	c := colly.NewCollector()
	
	// Gather and visit individual coin pages
	c.OnHTML(".currency-name-container", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		c.Visit(e.Request.AbsoluteURL(link+"historical-data/?start="+startDate+"&end="+endDate))
	})
	
	// Parse coin pages and write to file
	var name, symbol string
	c.OnHTML(".text-large", func(e *colly.HTMLElement) {
		name = e.ChildAttr(".currency-logo-32x32", "alt")
		symbol = strings.Trim(e.ChildText("small.bold.hidden-xs"), "()")
	})
	c.OnHTML("tbody .text-right", func(e *colly.HTMLElement) {
		
		hasData := regexp.MustCompile("20")
		if hasData.MatchString(e.ChildText("td:nth-child(1)")) {
			writer.Write([]string{
				name,
				symbol,
				convertScrapeDate(e.ChildText("td:nth-child(1)")),
				e.ChildText("td:nth-child(2)"),
				e.ChildText("td:nth-child(3)"),
				e.ChildText("td:nth-child(4)"),
				e.ChildText("td:nth-child(5)"),
				e.ChildText("td:nth-child(6)"),
				e.ChildText("td:nth-child(7)"),
			})
		}
	})

	c.Visit("https://coinmarketcap.com/all/views/all/")

	log.Printf("Scraping finished, check file %q for results\n", fName)
	
	notify.Notify("Coinmarketcap scraper", "Finished scraping", "Finished scraping", "path/to/icon.png")
}

func main() {
	conf, err := readConf("config.txt")
	if err != nil {
		fmt.Println("Error reading config file", err)
	}
	scrape(conf)
}