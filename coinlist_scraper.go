package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly"
)


func main() {
	fName := "coinlist.csv"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"Name", "Symbol", "Circulating Suply", "Volume (USD)"})

	// Instantiate collectors
	cMain := colly.NewCollector()
	cPage := colly.NewCollector()
	
	// to store data if the name is truncated
	var record []string
	
	cMain.OnHTML("#currencies-all tbody tr", func(e *colly.HTMLElement) {
		name := e.ChildText(".currency-name-container")
		symbol := e.ChildText(".col-symbol")
		supply := e.ChildAttr("td.circulating-supply a", "data-supply")
		volume := e.ChildAttr("a.volume", "data-usd")
		// Check if the currency name is truncated
		if strings.HasSuffix(name, "...") {
			record = []string{}
			record = append(record, symbol, supply, volume)
			cPage.Visit(e.Request.AbsoluteURL(e.ChildAttr(".currency-name-container", "href")))	
		} else {
			writer.Write([]string{
				name,
				symbol,
				supply,
				volume,
			})
		}
	})
	
	cPage.OnHTML(".text-large", func(e *colly.HTMLElement) {
		name := e.ChildAttr(".currency-logo-32x32", "alt")
		record = append([]string{name}, record...)
		writer.Write(record)
	})
	cMain.Visit("https://coinmarketcap.com/all/views/all/")

	log.Printf("Scraping finished, check file %q for results\n", fName)
}