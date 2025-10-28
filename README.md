# go-AnniversaryDay-SVG

誕生日や記念日の経過日数をSVG画像として生成するWebアプリケーションです。

## 機能概要

- 誕生日からの経過日数を計算してSVG画像化
- 記念日からの経過日数を計算してSVG画像化
- 3種類のレイアウトスタイル
- カスタム背景色対応（3桁/6桁のカラーコード）
- 干支表示（漢字/絵文字切り替え可能）
- WCAGガイドライン準拠のコントラスト比自動調整

## レイアウトタイプ
| type値 | 説明 | 特徴 |
|:--|:--|:--|
| *(空)* / `modern` / `card` | 最新のモダンレイアウト（１） | 左側アクセントバー + 右側大きな数字表示 |
| `simplecard` | シンプルカードモード（２） | シンプルな1行表示 |
| `legacy` | レガシーカードモード（３） |3行構成の従来型カード |

メンテや動作保証としてるのは１だけにする。。

## 使用方法

### ローカル実行

```bash
go run main.go
```

サーバーがポート9999で起動します。

### 基本的な使い方

以下のURLパターンでアクセスできます：

```
http://localhost:9999/api?birthday=19830506.svg
http://localhost:9999/api?birthday=20230506.svg
http://localhost:9999/api?anniversaryday=19050807.svg
```

### スタイル指定例

```
# モダンレイアウト（デフォルト）
http://localhost:9999/api?color=3CF&birthday=20010105.svg
http://localhost:9999/api?color=222&anniversaryday=20010105.svg

http://localhost:9999/api?color=ECE038&birthday=20010105.svg
http://localhost:9999/api?color=222&birthday=20040105.svg
http://localhost:9999/api?dispname=abc&color=CCC&birthday=20040105.svg


# シンプルカード
http://localhost:9999/api?color=3CF&type=simplecard&birthday=20010105.svg

# レガシーカード
http://localhost:9999/api?color=3CF&type=legacy&birthday=20010105.svg
```


http://localhost:9999/api?color=FFFFFF&type=modern&birthday=20010118.svg&appname=%E8%A8%98%E5%BF%B5%E6%97%A5.SVG

### 背景色カスタマイズ例

```
# 6桁カラーコード
http://localhost:9999/api?color=ECE038&type=card&birthday=19731108.svg
http://localhost:9999/api?color=493759&type=card&birthday=19731108.svg

# 3桁カラーコード
http://localhost:9999/api?color=3CF&type=card&birthday=20010105.svg
```

## Web公開版

Vercelでホスティングされているバージョンを利用できます：

```
https://go-anniversary-day-svg.vercel.app/api?birthday=19830506.svg
https://go-anniversary-day-svg.vercel.app/api?anniversaryday=20020918.svg
https://go-anniversary-day-svg.vercel.app/api?color=3CF&type=card&birthday=20010105.svg

https://go-anniversary-day-svg.vercel.app/api?color=222&type=card&birthday=20090105.svg

```

## 技術的特徴

- UTCベースの日付計算による正確な経過日数算出
- WCAGガイドラインに準拠したコントラスト比の自動計算
- レスポンシブSVGデザイン
- 3桁/6桁カラーコードのサポート
