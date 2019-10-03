package main

import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocarina/gocsv"
)

// Building は建物を表す
type Building struct {
	LineName    string   `json:"line"`
	StationName string   `json:"station"`
	Name        string   `json:"name"`
	Age         string   `json:"age"`
	Height      string   `json:"height"`
	Distance    []string `json:"distance"`
	Rooms       []Room   `json:"rooms"`
}

// Room は部屋を表す
type Room struct {
	Price      string `json:"price"`
	AdminPrice string `json:"admin"`
	Area       string `json:"area"`
}

// Property は物件を表す
// CSVとして出力するため配列を用いない
type Property struct {
	LineName    string `csv:"line"`
	StationName string `csv:"station"`
	Name        string `csv:"name"`
	Age         string `csv:"age"`
	Height      string `csv:"height"`
	Distance    string `csv:"distrance"`
	Price       string `csv:"price"`
	AdminPrice  string `csv:"admin"`
	Area        string `csv:"area"`
}

// Line は路線を表す
type Line struct {
	Name     string    `json:"name"`
	URL      string    `json:"url"`
	Stations []Station `json:"stations"`
}

// Station は駅を表す
type Station struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// 1つの駅の物件のページ数を取得
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

// スライスの最小値とそのインデックスを返す
func minSlice(slice []int) (m int, index int) {
	if len(slice) == 0 {
		panic("slice is empty.")
	}

	for i, val := range slice {
		if i == 0 || val < m {
			m = val
			index = i
		}
	}
	return
}

// 複数の駅からもっとも近い駅を取得
// "ＪＲ山手線/東京駅 歩6分"のような形式から数字を抜き出して比較
// 同じ時間の場合はインデックスが小さい方が選ばれる
func getMinDist(distances []string) string {
	times := []int{}
	for _, distance := range distances {
		rex := regexp.MustCompile(`\d+`)
		time, err := strconv.Atoi(rex.FindString(distance))
		if err != nil {
			// 駅からの距離が3つ無い場合("-"が入っている)
		} else {
			times = append(times, time)
		}
	}
	_, minIndex := minSlice(times)

	return distances[minIndex]
}

func scrapeStationPage(baseURL string, page int, lineName string, stationName string) ([]Building, []Property) {
	pageStr := strconv.Itoa(page)
	// baseURL(stationURL)は
	// https://suumo.jp/chintai/tokyo/ek_25620/?md=01&rn=0005&ts=1
	// のような形式で格納されている
	// パラメータ部分を削除しても良い
	// url := baseURL + "?page=" + pageStr + "&rn=0020"
	url := baseURL + "&page=" + pageStr

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
		building := Building{LineName: lineName, StationName: stationName}

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
			property := Property{LineName: lineName, StationName: stationName, Name: name, Distance: getMinDist(dist), Age: age, Height: height}

			price := sc.Find("span.cassetteitem_price--rent").Text()
			room.Price = price
			property.Price = price

			admin := sc.Find("span.cassetteitem_price--administration").Text()
			room.AdminPrice = admin
			property.AdminPrice = admin

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

// 特定の路線の駅のURLを取得
func scrapeStationURL(lineURL string) []Station {
	baseURL := "https://suumo.jp"
	url := baseURL + lineURL

	doc, err := goquery.NewDocument(url)
	if err != nil {
		panic(err)
	}

	stations := []Station{}

	items := doc.Find("tr.js-graph-data > td > a")
	items.Each(func(_ int, s *goquery.Selection) {
		name := s.Text()
		link, _ := s.Attr("href")
		station := Station{Name: name, URL: link}
		stations = append(stations, station)
	})

	return stations
}

// 沿線のURLを路線毎に取得
func scrapeLineURL(prefecture string) []Line {
	ensenURL := "https://suumo.jp/chintai/" + prefecture + "/ensen/"

	doc, err := goquery.NewDocument(ensenURL)
	if err != nil {
		panic(err)
	}

	lines := []Line{}

	labels := doc.Find("ul.searchitem-list > li > label > a")
	labels.Each(func(_ int, s *goquery.Selection) {
		name := s.Text()
		link, _ := s.Attr("href")
		line := Line{Name: name, URL: link}
		lines = append(lines, line)
	})

	for key, line := range lines {
		lines[key].Stations = scrapeStationURL(line.URL)
	}

	// test用
	// for i := 0; i < 2; i++ {
	// 	line := lines[i]
	// 	link := line.URL
	// 	lines[i].Stations = scrapeStationURL(link)
	// }

	return lines
}

func main() {
	lines := []Line{}

	prefectures := [...]string{"tokyo", "kanagawa"}
	for _, prefecture := range prefectures {
		prefectureLines := scrapeLineURL(prefecture)
		lines = append(lines, prefectureLines...)
	}

	baseURL := "https://suumo.jp"

	buildings := []Building{}
	properties := []Property{}
	for _, line := range lines {
		for _, station := range line.Stations {
			stationURL := baseURL + station.URL
			pages := getPages(stationURL)
			// ページ毎の情報を取得し連結
			for i := 1; i <= pages; i++ {
				building, property := scrapeStationPage(stationURL, i, line.Name, station.Name)
				buildings = append(buildings, building...)
				properties = append(properties, property...)
			}
		}
	}

	// test用
	// for i := 0; i < 2; i++ {
	// 	line := lines[i]
	// 	for j := 0; j < 2; j++ {
	// 		station := line.Stations[j]
	// 		stationURL := baseURL + station.URL
	// 		getPages(stationURL)
	// 		// ページ毎の情報を取得し連結
	// 		for k := 1; k <= 2; k++ {
	// 			building, property := scrapeStationPage(stationURL, k, line.Name, station.Name)
	// 			buildings = append(buildings, building...)
	// 			properties = append(properties, property...)
	// 		}
	// 	}
	// }

	outputJSON(buildings)
	outputCSV(properties)

}
