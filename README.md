# suumo-scraper
[SUUMO](https://suumo.jp/)から沿線ごとの物件情報をスクレイピングで取得するGoプログラム

## Config
main関数内の`prefectures`の要素に該当する都道府県内の全ての沿線にある物件情報を取得する．

## Output
`outputJSON()`, `outputCSV()`でそれぞれJSON形式，CSV形式で結果を出力できる．

## 注意
大規模なスクレイピングはSUUMOのサーバーに負担をかけることになります．
実行は各自の責任で行ってください．
