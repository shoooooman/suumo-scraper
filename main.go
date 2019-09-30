package main

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocarina/gocsv"
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

// Property は物件を表す
// CSVとして出力するため配列を用いない
type Property struct {
	Name     string `csv:"name"`
	Age      string `csv:"age"`
	Height   string `csv:"height"`
	Distance string `csv:"distrance"`
	Price    string `csv:"price"`
	Area     string `csv:"area"`
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

func scrapePage(baseURL string, page int) ([]Building, []Property) {
	pageStr := strconv.Itoa(page)
	url := baseURL + "?page=" + pageStr + "&rn=0020"

	doc, err := goquery.NewDocument(url)
	if err != nil {
		panic(err)
	}

	// 全ページにおける建物，物件情報
	buildings := []Building{}
	properties := []Property{}

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

		// 1ページ内の物件情報の配列
		items := []Property{}

		// 部屋を取得
		rselection := s.Find("div.cassetteitem-item > table > tbody")
		rselection.Each(func(_ int, sc *goquery.Selection) {
			room := Room{}
			// 駅からの距離は一番上に書いてあるものを採用
			property := Property{Name: name, Distance: dist[0], Age: age, Height: height}

			price := sc.Find("span.cassetteitem_price--rent").Text()
			room.Price = price
			property.Price = price

			area := sc.Find("span.cassetteitem_menseki").Text()
			room.Area = area
			property.Area = area

			building.Rooms = append(building.Rooms, room)
			items = append(items, property)
		})

		buildings = append(buildings, building)
		properties = append(properties, items...)
	})

	return buildings, properties
}

// JSONファイルを出力
func outputJSON(buildings []Building) {
	jsonBytes, err := json.Marshal(&buildings)
	if err != nil {
		panic(err)
	}

	jsonFile, err := os.Create(`output.json`)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	jsonFile.Write(jsonBytes)

	// 標準出力
	// fmt.Println(string(jsonBytes))
}

// CSVファイルを出力
func outputCSV(properties []Property) {
	csvFile, err := os.Create(`output.csv`)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	// SetCSVWriterでオーバーライド
	// 日本語対応
	// gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
	// 	writer := csv.NewWriter(transform.NewWriter(out, japanese.ShiftJIS.NewEncoder()))
	// 	return gocsv.NewSafeCSVWriter(writer)
	// })

	gocsv.MarshalFile(&properties, csvFile)
	if err != nil {
		panic(err)
	}

	// 標準出力
	// csvStr, err := gocsv.MarshalString(&properties)
	// fmt.Println(csvStr)
}

func main() {
	baseURL := "https://suumo.jp/chintai/tokyo/ek_27580/"

	pages := getPages(baseURL)

	buildings := []Building{}
	properties := []Property{}
	// ページ毎の情報を取得し連結
	for i := 1; i <= pages; i++ {
		buil, pro := scrapePage(baseURL, i)
		buildings = append(buildings, buil...)
		properties = append(properties, pro...)
	}

	outputJSON(buildings)
	outputCSV(properties)

}
