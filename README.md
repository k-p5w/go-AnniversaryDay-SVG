# go-AnniversaryDay-SVG



## やれること

/20220221.svg
という名前にすると

令和4年2月21日生まれは、いま0歳（生後18日）です

みたいなSVGイメージを作るだけのツール

### やりたいこと

モバイル端末向けの画像にしたい（気もする）

## 実行方法

main.goがある場所で以下コマンドを実行

$ go run main.go

### テスト方法、使い方

以下、URLを開くだけ

http://localhost:9999/api?birthday=19830506&type=card.svg
http://localhost:9999/api?birthday=19720312.svg
http://localhost:9999/api?anniversaryday=19050807.svg

19830506

## WEBアプリ版

- vercelなどの公開するホスティングできるところで公開したらこんな風に使えます。

https://go-anniversary-day-svg.vercel.app/api?birthday=19830506.svg
https://go-anniversary-day-svg.vercel.app/api?anniversaryday=20020918.svg
