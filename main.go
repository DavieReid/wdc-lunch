package main

import (
	"fmt"
	"log"
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

func main() {
	weekday := time.Now().Weekday()

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

			fmt.Println(menu)
			// todo: compare todays menu with what jack likes
			// send email or something
		}

	})

	err := c.Visit("https://www.west-dunbarton.gov.uk/schools-and-learning/schools/school-meals/primary-menus/")

	if err != nil {
		log.Fatal(err)
	}
}
