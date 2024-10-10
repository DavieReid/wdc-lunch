package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type MenuOption struct {
	Heading string
	Options []string
}

func (menuOption *MenuOption) AddItem(food string) []string {
	menuOption.Options = append(menuOption.Options, food)
	return menuOption.Options
}

type Menu struct {
	Meals []MenuOption
}

func (menu *Menu) AddMeal(menuOption MenuOption) []MenuOption {
	menu.Meals = append(menu.Meals, menuOption)
	return menu.Meals
}

func favesFromFile(file string) ([]string, error) {
	defer log.Printf("---------")
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := make([]string, 0)
	err = json.NewDecoder(f).Decode(&tok)

	log.Println("Read favourite foods: ", file)

	return tok, err
}

func main() {
	weekday := time.Now().Add(time.Hour).Weekday()

	if weekday == time.Saturday || weekday == time.Sunday {
		fmt.Println("It's the weekend, no packed lunch needed!")
	}

	fmt.Println("Getting Lunch Menu for: ", weekday)
	c := colly.NewCollector(colly.AllowedDomains("www.west-dunbarton.gov.uk"))

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	querySelector := fmt.Sprintf("table[summary='menu for %v']", weekday)

	c.OnHTML(querySelector, func(e *colly.HTMLElement) {

		menu := Menu{}
		tableBody := e.DOM.Find("tbody")

		if len(tableBody.Nodes) == 0 {
			fmt.Println("Unable to find the Lunch Menu for", weekday)
		} else {

			tableBody.Each(func(i int, tableBodies *goquery.Selection) {
				tableBodies.Find("tr").Each(func(i int, tableRows *goquery.Selection) {
					menuOption := MenuOption{}
					menuOption.Heading = tableRows.Find("th").Text()
					tableRows.Find("div [class='menu-item-entry']").Each(func(i int, menuItemEntries *goquery.Selection) {

						// most menu items are links but "water" isn't as everyone knows what that is
						if menuItemEntries.Children().Length() > 1 {
							menuItemEntries.Find("a").Each(func(i int, s4 *goquery.Selection) {
								menuOption.AddItem(s4.Text())
							})
						} else {
							// water
							menuOption.AddItem(menuItemEntries.Text())
						}
					})
					menu.AddMeal(menuOption)
				})
			})

			faves, err := favesFromFile("./favs.json")

			if err != nil {
				fmt.Println("Whoops we don't know what Jack likes")
			}

			needPackedLunch := true
			for _, fave := range faves {
				for i := range menu.Meals {
					options := menu.Meals[i].Options
					for _, option := range options {
						if option == fave {
							fmt.Println("Found", option)
							needPackedLunch = false
							break
						}
					}
				}
			}

			if needPackedLunch {
				fmt.Println("A packed lunch is needed!")
				// send saying we need packed lunch
			} else {
				fmt.Println("No packed lunch is needed!")
				// send saying we don't packed lunch
			}
		}

	})

	err := c.Visit("https://www.west-dunbarton.gov.uk/schools-and-learning/schools/school-meals/primary-menus/")

	if err != nil {
		log.Fatal(err)
	}
}
