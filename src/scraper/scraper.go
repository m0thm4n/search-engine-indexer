package scraper

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	titles = make([]string, 0)
)

// Scraper for each website
type Scraper struct {
	url string
	doc *goquery.Document
}

// NewScraper builds a new scraper for the website
func NewScraper(u string) *Scraper {
	if !strings.HasPrefix(u, "http") {
		return nil
	}

	response, err := http.Get(u)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer response.Body.Close()

	d, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return &Scraper{
		url: u,
		doc: d,
	}
}

// Body returns a string with the body of the page
func (s *Scraper) Body() string {
	body := s.doc.Find("body").Text()
	// Remove leading/ending white spaces
	body = strings.TrimSpace(body)

	return body
}

func (s *Scraper) buildLinks(href string) string {
	var link string

	if strings.HasPrefix(href, "/") {
		link = strings.Join([]string{s.url, href}, "")
	} else {
		link = href
	}

	link = strings.TrimRight(link, "/")
	link = strings.TrimRight(link, ":")

	return link
}

// Links returns an array with all the links from the website
func (s *Scraper) Links() ([]string, []string) {
	links := make([]string, 0)

	var link string

	s.doc.Find("body a").Each(func(index int, item *goquery.Selection) {
		link = ""

		linkTag := item
		href, _ := linkTag.Attr("href")
		title := linkTag.Text()

		if strings.Contains(href, "/cooking/recipe-ideas") || strings.Contains(href, "/everyday-cooking/quick-and-easy/") || strings.Contains(href, "/search?q=Chicken") || strings.Contains(href, "/search?q=Tacos") {
			if !strings.HasPrefix(href, "#") && !strings.HasPrefix(href, "javascript") {
				link = s.buildLinks(href)
				if link != "" && !slices.Contains(titles, title) && !slices.Contains(links, link) {
					links = append(links, link)
					titles = append(titles, title)
				}
			}
		}
	})

	return links, titles
}

// MetaDataInformation returns the title and description from the page
func (s *Scraper) MetaDataInformation() (string, string) {
	var t string
	var d string

	t = s.doc.Find("title").Contents().Text()

	s.doc.Find("meta").Each(func(index int, item *goquery.Selection) {
		if item.AttrOr("name", "") == "description" || item.AttrOr("property", "") == "og:description" {
			d = item.AttrOr("content", "")
		}
	})

	return t, d
}
