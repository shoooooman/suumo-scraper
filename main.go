package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

// Building は建物を表す
type Building struct {
	Name     string   `json:"name"`
	Age      string   `json:"age"`
	Height   string   `json:"height"`
	Distance []string `json:"distance"`
	Rooms    []Room   `json:"rooms"`
}

// Room は部屋を表す
type Room struct {
	Price string `json:"price"`
	Area  string `json:"area"`
}

func getPages(url string) int {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		panic(err)
	}

	// 最後のページを取得
	// titleから取得しても良い
	pagesStr := doc.Find("div.pagination.pagination_set-nav > ol > li").Last().Text()
	pages, _ := strconv.Atoi(pagesStr)

	return pages
}

func scrapePage(baseURL string, page int) []Building {
	pageStr := strconv.Itoa(page)
	url := baseURL + "?page=" + pageStr + "&rn=0020"

	doc, err := goquery.NewDocument(url)
	if err != nil {
		panic(err)
	}

	buildings := []Building{}

	// 建物を全て取得
	selection := doc.Find("div.cassetteitem")
	selection.Each(func(index int, s *goquery.Selection) {
		building := Building{}
		name := s.Find("div.cassetteitem_content-title").Text()
		building.Name = name

		// 駅からの距離
		col2 := s.Find("li.cassetteitem_detail-col2 > div")
		dist := []string{}
		col2.Each(func(_ int, sc *goquery.Selection) {
			dist = append(dist, sc.Text())
		})
		building.Distance = dist

		// 築年数と高さ
		col3 := s.Find("li.cassetteitem_detail-col3 > div")
		age := col3.First().Text()
		height := col3.Last().Text()
		building.Age = age
		building.Height = height

		// 部屋を取得
		rselection := s.Find("div.cassetteitem-item > table > tbody")
		rselection.Each(func(_ int, sc *goquery.Selection) {
			room := Room{}
			price := sc.Find("span.cassetteitem_price--rent").Text()
			room.Price = price

			area := sc.Find("span.cassetteitem_menseki").Text()
			room.Area = area

			building.Rooms = append(building.Rooms, room)
		})

		buildings = append(buildings, building)
	})

	return buildings
}

func main() {
	baseURL := "https://suumo.jp/chintai/tokyo/ek_27580/"

	pages := getPages(baseURL)
	fmt.Println(pages)

	buildings := []Building{}
	// ページ毎の情報を取得し連結
	for i := 1; i <= pages; i++ {
		info := scrapePage(baseURL, i)
		buildings = append(buildings, info...)
	}

	jsonBytes, err := json.Marshal(&buildings)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))

}
