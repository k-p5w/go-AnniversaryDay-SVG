package anniversary

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"
)

// go mod init github.com/k-p5w/go-AnniversaryDay-SVG
// github.com/k-p5w/go-AnniversaryDay-SVG
// layoutYMD is SVGで描画する年月日
const layoutYMD = "2006年1月2日"

// CANVAS向け定数
const (
	FontSize       = 40
	FrameRoundness = FontSize / 2
	FrameBase      = FontSize * 5
	FrameTextXY    = FontSize * 30
	canvasHeight   = FontSize * 2
	FrameHeight    = FontSize
	TextBaseX      = FontSize
	TextBaseY      = FontSize + (FontSize / 3)
	FrameXY        = FontSize / 2
)

type ColorInfo struct {
	BackgroundColor string
	// 基準の色
	BaseColor string
	// 補色
	ComplementaryColor string
	// 反転色
	InvertColor string
}
type RGB struct {
	R, G, B float64
}

type AgeInfo struct {
	Age                  int
	TotalDate            int
	BaseDate             string
	Text                 string
	MultiText1           string
	MultiText2           string
	MultiText3           string
	SexagenaryCycle      string
	SexagenaryCycleColor string
}

type CanvasInfo struct {
	CanvasHeight   int
	CanvasWidth    int
	FrameWidth     int
	FrameHeight    int
	TextAreaHeight int
	TextAreaUpY    int
	FontSize       int
}

// Handler is /APIから呼ばれる
func Handler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Handler start.")

	// getパラメータの解析
	q := r.URL.Query()

	// どういう状況で閲覧されているかで描画内容を変えるためにロジックをいる

	agent := r.UserAgent()

	svgType := ""
	svgType = q.Get("type")
	svgBGcolor := "#FFF"
	qColor := ""
	qColor = q.Get("color")
	// 3桁か6桁なら色扱いにする（ホントはカラーコードに変換できるかチェックいるんだろうなあ）
	if len(qColor) == 3 {
		svgBGcolor = fmt.Sprintf("#%v", qColor)
	} else if len(qColor) == 6 {
		svgBGcolor = fmt.Sprintf("#%v", qColor)
	}

	textColor := getFontColor(svgBGcolor)
	drawMode := "normal"
	if len(svgType) == 0 {

		if strings.Index(agent, "Windows") > 0 {
			fmt.Println("Windows!")
			drawMode = "wide"
		} else {
			if strings.Index(agent, "Macintosh") > 0 {
				fmt.Println("Macintosh!")
				drawMode = "wide"
			} else {
				fmt.Println("mobile!")

			}

		}
	} else {
		// 　描画タイプが指定されている場合
		drawMode = "normal"
	}
	fmt.Println(svgType)
	svgname := ""
	itemTxt1 := ""
	itemTxt2 := ""
	itemTxt3 := "【%v】\n"
	// 誕生日or記念日モードの判定
	svgname = q.Get("birthday")
	if len(svgname) > 0 {

		itemTxt1 = " %v生まれ"
		itemTxt2 = "%v歳(生後%v日)"
	}
	svgnameAniv := q.Get("anniversaryday")
	if len(svgnameAniv) > 0 {
		itemTxt1 = "%v開始"
		itemTxt2 = " %v周年(%v日目)"
		svgname = svgnameAniv
	}
	svgPage := "<h1>エラーが発生しました.</h1>"

	yyyymmdd := ""
	// SVGで終わっていること
	if strings.HasSuffix(svgname, ".svg") {
		yyyymmdd = strings.Replace(svgname, ".svg", "", -1)
		// actorName = filepath.Base(svgname)
		fmt.Printf("%v => %v", svgname, yyyymmdd)
	} else {
		return
	}

	// canvasText := canvasBase / 2

	BaseColor := "#5AA572"

	// テキストを加工する
	ai := searchBirthDay(yyyymmdd, itemTxt1, itemTxt2, itemTxt3, svgType)
	BaseColor = ai.SexagenaryCycleColor
	pallet := getColorPallet(BaseColor)
	// フォントサイズの導出
	nameLen := len(ai.Text)
	txtLen := CountInString(ai.Text)
	fmt.Println(nameLen)

	// circle
	frameWidth := (FontSize * txtLen) - (FontSize * 5)
	TextShadowX := TextBaseX + (FontSize / 20)
	TextShadowY := TextBaseY + (FontSize / 20)
	canvasWidth := frameWidth + 100

	fs := FontSize
	if drawMode != "wide" {
		fs = FontSize / 2
		frameWidth = fs + 1
		TextShadowX = FontSize / 4
		TextShadowY = FontSize / 4
		canvasWidth = fs + 4
	}

	fmt.Println(pallet.InvertColor)
	rxy := 12

	// 下のtextは黒字で、上のテキストは色を付ける
	svgPage = fmt.Sprintf(`
	<svg width="%v" height="%v" xmlns="http://www.w3.org/2000/svg" 		xmlns:xlink="http://www.w3.org/1999/xlink"		>
		<rect x="%v" y="%v" rx="%v" ry="%v" width="%v" 	height="%v" stroke="%v" fill="transparent" stroke-width="%v" />
		<text x="%v" y="%v" style="text-anchor:start;font-size:%vpx;fill:%v;font-family: Meiryo,  Verdana, Helvetica, Arial, sans-serif;"			>			
		%v
		</text>
		<text x="%v" y="%v" style="text-anchor:start;font-size:%vpx;fill:RGB(2,2,2);font-family: Meiryo,  Verdana, Helvetica, Arial, sans-serif;">
        %v
    	</text>
	</svg>
	`, canvasWidth, canvasHeight,
		FrameXY, FrameXY, rxy, rxy, frameWidth, FrameHeight, pallet.BaseColor, 10,
		TextShadowX, TextShadowY, fs, pallet.InvertColor,
		ai.Text,
		TextBaseX, TextBaseY, fs, ai.Text)

	var cnvs CanvasInfo
	cnvs.CanvasHeight = canvasHeight + fs
	cnvs.TextAreaHeight = TextBaseY + fs
	cnvs.FrameHeight = FrameHeight + fs
	cnvs.FrameWidth = frameWidth / 2
	cnvs.TextAreaUpY = TextShadowY * 2
	cnvs.CanvasHeight = 600
	cnvs.TextAreaHeight = 350
	cnvs.FrameHeight = 400
	cnvs.FrameWidth = 550
	cnvs.FontSize = fs
	pallet.BackgroundColor = svgBGcolor
	// cardサイズが285pxだったのでそれに最適化させよ
	svgPageUniversal := fmt.Sprintf(`
		<svg class="square" viewbox="0 0 100 100" width="270px" height="100px"  xmlns="http://www.w3.org/2000/svg" 		xmlns:xlink="http://www.w3.org/1999/xlink"		>
		    <path d="M0 0 L 640 0 L 640 320 L 0 320" style="fill:%v;stroke-width:0" />
			<circle cx="5" cy="5" r="40" fill="%v" />
			<text x="%v" y="%v" style="text-anchor:start;font-size:%vpx;fill:%v;font-family: Meiryo,  Verdana, Helvetica, Arial, sans-serif;"			>			
			%v
			</text>
			<text x="%v" y="%v" style="text-anchor:start;font-size:%vpx;fill:%v;font-family: Meiryo,  Verdana, Helvetica, Arial, sans-serif;"			>			
			%v
			</text>			
			<text x="%v" y="%v" style="text-anchor:start;font-size:%vpx;fill:%v;font-family: Meiryo,  Verdana, Helvetica, Arial, sans-serif;">
			%v
			</text>
		</svg>
		`, pallet.BackgroundColor, pallet.BaseColor,
		TextShadowX, cnvs.TextAreaUpY, cnvs.FontSize, textColor,
		ai.MultiText1,
		TextShadowX, cnvs.TextAreaUpY+cnvs.FontSize, cnvs.FontSize, pallet.InvertColor,
		ai.MultiText2,
		TextShadowX+5, cnvs.TextAreaUpY+(cnvs.FontSize*2), cnvs.FontSize, textColor,
		ai.MultiText3)

	// Content-Type: image/svg+xml
	// Vary: Accept-Encoding
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Vary", "Accept-Encoding")

	// 描画を2種類（mobile向けと横長）を用意した

	// パラメータによって切り替える
	if drawMode == "wide" {
		fmt.Fprint(w, svgPage)
	} else {
		fmt.Fprint(w, svgPageUniversal)
	}

}

func CountInString(str string) int {
	length := 0
	var it norm.Iter
	it.InitString(norm.NFC, str)

	for !it.Done() {
		length++
		it.Next()
	}

	return length
}

// getColorPallet is 補色や反対色を取得するメソッド
func getColorPallet(c string) ColorInfo {

	var cp ColorInfo

	// 16進数→10進数
	rPt, _ := strconv.ParseUint(c[1:3], 16, 0)
	gPt, _ := strconv.ParseUint(c[3:5], 16, 0)
	bPt, _ := strconv.ParseUint(c[5:7], 16, 0)

	// int->float
	r := float64(rPt)
	g := float64(gPt)
	b := float64(bPt)

	min := math.Min(r, math.Min(g, b)) //Min. value of RGB
	max := math.Max(r, math.Max(g, b)) //Max. value of RGB
	pt := max + min                    //Delta RGB value

	newColorRGB := &RGB{pt - r, pt - g, pt - b}
	newColorRGB2 := &RGB{255 - r, 255 - g, 255 - b}

	// // float->int
	newR := int(newColorRGB.R)
	newG := int(newColorRGB.G)
	newB := int(newColorRGB.B)

	newR2 := int(newColorRGB2.R)
	newG2 := int(newColorRGB2.G)
	newB2 := int(newColorRGB2.B)

	cp.BaseColor = c
	cp.ComplementaryColor = fmt.Sprintf("RGB(%v,%v,%v)", newR, newG, newB)
	cp.InvertColor = fmt.Sprintf("RGB(%v,%v,%v)", newR2, newG2, newB2)
	return cp
}

// 背景色（6桁のカラーコード）から文字色を出す関数
func textColorFromBackgroundColor(bgColor string) string {
	// 背景色の赤・緑・青の成分を取り出す
	// r, _ := strconv.ParseInt(bgColor[0:2], 16, 0)
	// g, _ := strconv.ParseInt(bgColor[2:4], 16, 0)
	// b, _ := strconv.ParseInt(bgColor[4:6], 16, 0)

	purserRGB := func(bgColor string) (int, int, int) {
		// 背景色のRGB値を取得
		var red int
		var green int
		var blue int
		// 16進数で読みだす
		fmt.Printf("start input:%v \n", bgColor)
		// これで16進数を取得できるはずなのにうまくいかない？
		// fmt.Sscanf(bgColor, "%x%x%x", &red, &green, &blue)
		// #FF9900のときはこれで取得できる

		// カラーコードが3桁の場合、各色を2倍する
		if len(bgColor) == 4 {
			fmt.Sscanf(bgColor, "#%01x%01x%01x", &red, &green, &blue)
			//
			red = red * red
			green = green * green
			blue = blue * blue
		} else {
			fmt.Sscanf(bgColor, "#%02x%02x%02x", &red, &green, &blue)
		}
		fmt.Printf("[purserRGB:%v]R:%v G:%v B:%v \n", bgColor, red, green, blue)

		return red, green, blue
	}
	r, g, b := purserRGB(bgColor)
	// 背景色の相対輝度を計算する
	bgLuminance := relativeLuminance(r, g, b)

	// 文字色の候補を定義する
	textColors := []string{"#000000", "#FFFFFF", "#FF0000", "#00FF00", "#0000FF"}

	// 文字色の候補ごとにコントラスト比を計算し、最も高いものを選ぶ
	maxContrast := 0.0
	bestTextColor := ""
	for _, textColor := range textColors {
		r, g, b := purserRGB(textColor)
		textLuminance := relativeLuminance(r, g, b)
		contrast := contrastRatio(bgLuminance, textLuminance)
		fmt.Printf("bgColor:%v(%v) contrast:%v textColor:%v(%v) bestTextColor:%v\n", bgColor, bgLuminance, contrast, textColor, textLuminance, bestTextColor)
		if contrast > maxContrast {
			maxContrast = contrast
			bestTextColor = textColor
		}
	}

	// コントラスト比が4.5:1以上になるように文字色を決める
	if maxContrast >= 4.5 {
		fmt.Printf("%v:%v \n", maxContrast, bestTextColor)
		return fmt.Sprintf("%s", bestTextColor)
	} else {
		return "#000000"
	}
}

// 相対輝度を計算する関数
func relativeLuminance(r, g, b int) float64 {
	fmt.Printf("[relativeLuminance]R:%v G:%v B:%v \n", r, g, b)
	var rs, gs, bs float64
	if r <= 10 {
		rs = float64(r) / 3294.0
	} else {
		rs = math.Pow(float64(r)/269.0+0.0513, 2.4)
	}
	if g <= 10 {
		gs = float64(g) / 3294.0
	} else {
		gs = math.Pow(float64(g)/269.0+0.0513, 2.4)
	}
	if b <= 10 {
		bs = float64(b) / 3294.0
	} else {
		bs = math.Pow(float64(b)/269.0+0.0513, 2.4)
	}
	// コントラスト比 = (明るい色の相対輝度 + 0.05) / (暗い色の相対輝度 + 0.05)
	return 0.2126*rs + 0.7152*gs + 0.0722*bs
}

// コントラスト比を計算する関数
func contrastRatio(l1, l2 float64) float64 {
	if l1 > l2 {
		return (l1 + 0.05) / (l2 + 0.05)
	} else {
		return (l2 + 0.05) / (l1 + 0.05)
	}
}

func getFontColor(colorCode string) string {
	// 背景色の明るさを計算
	textColor := textColorFromBackgroundColor(colorCode)

	return textColor

}

// searchBirthDay is yyyymmddから表示用のテキストに加工する
func searchBirthDay(base string, itemTxt1 string, itemTxt2 string, itemTxt3 string, svgType string) AgeInfo {
	var info AgeInfo
	eto := []string{"子・ねずみ", "丑・うし", "寅・とら", "卯・うさぎ", "辰・たつ", "巳・へび", "午・うま", "未・ひつじ", "申・さる", "酉・とり", "戌・いぬ", "亥・いのしし"}

	if svgType == "card" {
		eto = []string{"🐁", "🐄", "🐅", "🐇", "🐉", "🐍", "🐎", "🐑", "🐒", "🐔", "🐕", "🐗"}
	}

	etoColor := []string{"#edbc5f", "#f8bac1", "#8ac1d4", "#3d9bcf", "#3c895d", "#936791", "#e76739", "#fdfbfe", "#b3d07e", "#c88f4e", "#e3d1b0", "#90664e"}

	// 9FA0A0
	// 辰、B9C998
	year, yerr := strconv.Atoi(base[0:4])
	if yerr != nil {
		fmt.Println(yerr)
	}
	month, merr := strconv.Atoi(base[4:6])
	if merr != nil {
		fmt.Println(merr)
	}
	// 19001230
	// 01234567
	day, derr := strconv.Atoi(base[6:8])
	if derr != nil {
		fmt.Println(derr)
	}
	fmt.Printf("%v:%v-%v-%v \n", base, year, month, day)
	birthDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	duration := now.Sub(birthDate)

	hours0 := int(duration.Hours())
	days := hours0 / 24

	info.BaseDate = birthDate.Format(layoutYMD)
	info.TotalDate = days
	info.Age = days / 365

	// 3才以上は生誕にしよ
	if info.Age > 3 {
		itemTxt2 = "%v歳(生誕%v日)"
	}

	info.SexagenaryCycle = eto[(year-4)%12]
	info.SexagenaryCycleColor = etoColor[(year-4)%12]

	info.MultiText1 = fmt.Sprintf(itemTxt1, info.BaseDate)
	info.MultiText2 = fmt.Sprintf(itemTxt2, info.Age, info.TotalDate)
	info.MultiText3 = fmt.Sprintf(itemTxt3, info.SexagenaryCycle)
	itemTxt := itemTxt1 + " => " + itemTxt2 + ":" + itemTxt3
	if svgType == "card" {
		info.MultiText1 = info.SexagenaryCycle + info.MultiText1
		info.MultiText3 = info.MultiText2
		info.MultiText2 = ""
	}
	txtMain := fmt.Sprintf(itemTxt, info.BaseDate, info.Age, info.TotalDate, info.SexagenaryCycle)
	info.Text = txtMain

	return info
}
