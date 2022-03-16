package anniversary

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// layoutYMD is SVGで描画する年月日
const layoutYMD = "2006/1/2"

// CANVAS向け定数
const (
	FrameRoundness = 20
	FrameBase      = 200
	FontSize       = 30
	FrameTextXY    = 1300
	canvasHeight   = FontSize * 2
	FrameHeight    = canvasHeight / 3
	TextBaseX      = 30
	TextBaseY      = FontSize + 10
	FrameXY        = FontSize / 2
)

type ColorInfo struct {
	BaseColor          string
	ComplementaryColor string
	InvertColor        string
}
type RGB struct {
	R, G, B float64
}

type AgeInfo struct {
	Age             int
	TotalDate       int
	BaseDate        string
	Text            string
	SexagenaryCycle string
}

// Handler is /APIから呼ばれる
func Handler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Handler start.")

	// getパラメータの解析
	q := r.URL.Query()
	svgname := ""
	itemTxt := ""
	svgname = q.Get("birthday")
	if len(svgname) == 0 {
		svgname = q.Get("anniversaryday")
		itemTxt = "start:%v ⇒ %v周年（開始後%v日）【%v】\n"
	} else {
		itemTxt = " %v生まれ ⇒ %v歳（生後%v日）【%v】\n"
	}

	svgPage := "<h1>エラーが発生しました.</h1>"
	fmt.Println(q)

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
	pallet := getColorPallet(BaseColor)

	ai := searchBirthDay(yyyymmdd, itemTxt)
	// フォントサイズの導出
	nameLen := len(ai.Text)

	// circle
	frameWidth := (FontSize * nameLen) / 2
	TextShadowX := TextBaseX + (FontSize / 10)
	TextShadowY := TextBaseY + (FontSize / 10)
	canvasWidth := frameWidth + 100

	fmt.Println(pallet.InvertColor)
	rxy := 10

	// 下のtextは黒字で、上のテキストは色を付ける
	svgPage = fmt.Sprintf(`
	<svg width="%v" height="%v" xmlns="http://www.w3.org/2000/svg" 		xmlns:xlink="http://www.w3.org/1999/xlink"		>
		<rect x="%v" y="%v" rx="%v" ry="%v" width="%v" 	height="%v" 			stroke="%v" 			fill="transparent" stroke-width="%v" 			/>
		<text x="%v" y="%v" style="text-anchor:start;font-size:%vpx;fill:%v;font-family: Meiryo,  Verdana, Helvetica, Arial, sans-serif;"			>			
		%v
		</text>
		<text x="%v" y="%v" style="text-anchor:start;font-size:%vpx;fill:RGB(2,2,2);font-family: Meiryo,  Verdana, Helvetica, Arial, sans-serif;">
        %v
    	</text>
	</svg>
	`, canvasWidth, canvasHeight,
		FrameXY, FrameXY, rxy, rxy, frameWidth, FrameHeight, pallet.BaseColor, 2,
		TextShadowX, TextShadowY, FontSize, pallet.InvertColor,
		ai.Text,
		TextBaseX, TextBaseY, FontSize, ai.Text)

	// Content-Type: image/svg+xml
	// Vary: Accept-Encoding
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Vary", "Accept-Encoding")

	fmt.Fprint(w, svgPage)
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

func searchBirthDay(base string, itemTxt string) AgeInfo {
	var info AgeInfo
	eto := []string{"子・ねずみ", "丑・うし", "寅・とら", "卯・うさぎ", "辰・たつ", "巳・へび", "午・うま", "未・ひつじ", "申・さる", "酉・とり", "戌・いぬ", "亥・いのしし"}

	// 9FA0A0
	// 辰、B9C998
	year, yerr := strconv.Atoi(base[0:4])
	if yerr != nil {
		fmt.Println(yerr)
	}
	month, merr := strconv.Atoi(base[5:6])
	if merr != nil {
		fmt.Println(merr)
	}
	day, derr := strconv.Atoi(base[7:8])
	if derr != nil {
		fmt.Println(derr)
	}
	birthDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	duration := now.Sub(birthDate)

	hours0 := int(duration.Hours())
	days := hours0 / 24

	info.BaseDate = birthDate.Format(layoutYMD)
	info.TotalDate = days
	info.Age = days / 365
	info.SexagenaryCycle = eto[(year-4)%12]

	txtMain := fmt.Sprintf(itemTxt, info.BaseDate, info.Age, info.TotalDate, info.SexagenaryCycle)
	info.Text = txtMain

	return info
}
