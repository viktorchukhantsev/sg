package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

const rubygemsRoot = "http://rubygems.org"
const rubygemsSearch = "http://rubygems.org/search?utf8=%E2%9C%93&query="

type Gem struct {
	name        string
	url         string
	version     string
	description string
	position    int
}

// sort by position
type byPosition []Gem

func (a byPosition) Len() int      { return len(a) }
func (a byPosition) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byPosition) Less(i, j int) bool {
	if a[i].position < a[j].position {
		return true
	}
	if a[i].position > a[j].position {
		return false
	}
	return a[i].position < a[j].position
}

func itHasClass(t html.Token, class string) (hasClass bool) {
	for _, linkAttr := range t.Attr {
		if linkAttr.Key == "class" && linkAttr.Val == class {
			hasClass = true
		}
	}
	return
}

func isGem(t html.Token) bool {
	return itHasClass(t, "gems__gem")
}

func isGemAnchor(t html.Token) bool {
	return t.Data == "a" && isGem(t)
}

func isGemName(t html.Token) bool {
	return itHasClass(t, "gems__gem__name")
}

func isGemDesc(t html.Token) bool {
	return itHasClass(t, "gems__gem__desc")
}

func isGemVersion(t html.Token) bool {
	return itHasClass(t, "gems__gem__version")
}

func getName(t html.Token) (name string) {
	for _, h2 := range t.Attr {
		name = h2.Val
	}

	return
}

func getVersion(t html.Token) (name string) {
	for _, span := range t.Attr {
		name = span.Val
	}

	return
}

func getDescription(t html.Token) (desc string) {
	for _, span := range t.Attr {
		desc = span.Val
	}

	return
}

func getHref(t html.Token) (href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
		}
	}

	return
}

func searchGems(url string) map[string]*Gem {
	gems := make(map[string]*Gem)

	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("ERROR: Failed to reach RubyGems")
		return gems
	}

	b := resp.Body
	defer b.Close() // close Body when the function returns

	z := html.NewTokenizer(b)

	pos := 1
	q := new(Gem)

	saveName := false
	saveVer := false
	saveDesc := false

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return gems
		case tt == html.StartTagToken:
			t := z.Token()

			if isGem(t) {
				if len(q.name) > 0 {
					q.position = pos
					gems[q.name] = q
					q = new(Gem)
					pos = pos + 1
				}
			}

			if isGemAnchor(t) {
				href := getHref(t)
				q.url = rubygemsRoot + href
			}

			saveName = isGemName(t)
			saveVer = isGemVersion(t)
			saveDesc = isGemDesc(t)
		case tt == html.TextToken:
			t := z.Token()
			if saveName {
				q.name = strings.TrimSpace(t.Data)
				saveName = false
			}
			if saveVer {
				q.version = strings.TrimSpace(t.Data)
				saveVer = false
			}
			if saveDesc {
				q.description = strings.TrimSpace(t.Data)
				saveDesc = false
			}
		}
	}
}

func searchCommand(searchString string) {
	gemsMap := searchGems(rubygemsSearch + searchString)
	gems := make([]Gem, 0, len(gemsMap))
	for _, v := range gemsMap {
		gems = append(gems, *v)
	}

	fmt.Println("Found", len(gems), "gems:")
	sort.Sort(byPosition(gems))

	for _, g := range gems {
		fmt.Printf("%s %s %s \n", g.name, g.version, g.url)
	}

	os.Exit(0)
}

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("    sg search gem\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	if flag.Arg(0) == "search" {
		searchString := strings.Join(flag.Args()[1:], "+")
		searchCommand(searchString)
	}
}
