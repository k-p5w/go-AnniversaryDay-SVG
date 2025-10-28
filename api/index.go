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
	// Legacy/Wide Modeで使用するサイズ
	FontSize       = 40
	FrameRoundness = FontSize / 2
	FrameHeight    = FontSize
	FrameXY        = 20
	// Simple Card Modeで使用するサイズ
	SimpleFontSize = 20
	SimpleHeight   = 100
	SimpleWidth    = 270
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
	BaseDate             string // YYYY年M月D日形式 (例: 2001年1月5日)
	BaseDateDescription  string // YYYY年M月D日 + 説明 (例: 2001年1月5日生まれ)
	Text                 string // Legacy/Wide用の一行テキスト
	MultiText1           string // カード用Line1, 干支+日付の説明
	MultiText2           string // カード用Line2
	MultiText3           string // カード用Line3, 年齢+日数
	SexagenaryCycle      string // 干支の名称または絵文字 (例: 巳・へび または 🐍)
	SexagenaryCycleColor string // 干支の色
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

	// ★ typeパラメータで3モードを明確に制御 (ユーザーの最終リクエストに従う)
	svgType := q.Get("type")

	// ② Modern Layout: パラメータなし、modern、cardが該当 (テンプレート選択ロジック)
	isModernMode := len(svgType) == 0 || svgType == "modern" || svgType == "card"

	// ① Simple Card Mode: simplecard が該当 (コンパクトな複数行カード)
	isSimpleCardMode := svgType == "simplecard"

	// ③ Legacy Card Mode: legacy が該当 (ワイドな単一行フレーム)
	isLegacyCardMode := svgType == "legacy"

	svgBGcolor := "#FFF"
	qColor := q.Get("color")
	// 3桁か6桁なら色扱いにする
	if len(qColor) == 3 {
		// 3桁の場合は6桁に展開して格納
		svgBGcolor = fmt.Sprintf("#%c%c%c%c%c%c", qColor[0], qColor[0], qColor[1], qColor[1], qColor[2], qColor[2])
	} else if len(qColor) == 6 {
		svgBGcolor = fmt.Sprintf("#%v", qColor)
	}

	// [修正点1] アプリ名/ロゴをクエリパラメータから取得。パラメータ名を "dispname" に変更。
	// 指定がない場合は空文字列を使用し、デフォルト名を表示しない。
	appName := q.Get("dispname")

	fmt.Println(svgType)
	svgname := ""
	itemTxt1 := ""
	itemTxt2 := ""
	itemTxt3 := "【%v】\n" // Modern/Legacyで使われるが、Simple Cardでは上書きされる

	// 誕生日or記念日モードの判定
	svgnameBirth := q.Get("birthday")
	isBirthMode := len(svgnameBirth) > 0

	if isBirthMode {
		itemTxt1 = " %v生まれ"
		itemTxt2 = "%v歳[生後%v日]" // 初期値は「生後」
		svgname = svgnameBirth
	}
	svgnameAniv := q.Get("anniversaryday")
	if len(svgnameAniv) > 0 {
		itemTxt1 = "%v開始"
		itemTxt2 = " %v周年[%v日目]"
		svgname = svgnameAniv
	}
	svgPage := "<h1>エラーが発生しました.</h1>"

	yyyymmdd := ""
	// SVGで終わっていること
	if strings.HasSuffix(svgname, ".svg") {
		yyyymmdd = strings.Replace(svgname, ".svg", "", -1)
		fmt.Printf("%v => %v", svgname, yyyymmdd)
	} else {
		return
	}

	BaseColor := "#5AA572"

	// テキストを加工する (全てのモードに必要な情報を取得)
	// useCardEmojiがtrueのとき絵文字 (Simple Card Mode)
	// ★ Modern Modeでも絵文字のみにしたいので、isModernModeでもtrueにする
	useCardEmoji := isSimpleCardMode || isModernMode

	// ★ itemTxt2が searchBirthDay内で「生誕」に書き換えられる
	ai := searchBirthDay(yyyymmdd, itemTxt1, itemTxt2, itemTxt3, useCardEmoji)
	BaseColor = ai.SexagenaryCycleColor
	pallet := getColorPallet(BaseColor)

	rxy := 12

	// Content-Type: image/svg+xml
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Vary", "Accept-Encoding")

	// Goでは三項演算子が使えないため、単位と接頭辞を先に決定
	var unitText string
	if isBirthMode {
		unitText = "歳"
	} else {
		unitText = "周年"
	}

	var prefixText string
	if isBirthMode {
		prefixText = "生後" // デフォルトは「生後」
	} else {
		prefixText = ""
	}

	// 総日数が1000日以上なら「生誕」に切り替える
	if isBirthMode {
		if ai.TotalDate >= 1000 {
			prefixText = "生誕"
		} else {
			prefixText = "生後" // 1000日未満
		}
	}

	// 背景色から決定された最適な文字色
	dynamicFontColor := getFontColor(svgBGcolor)

	// ★★★ 描画モードの判定と出力の切り替え ★★★

	// 1. MODERN Style: type="", type=modern, type=card (①)
	if isModernMode {

		// ai.SexagenaryCycle は、searchBirthDayで干支の絵文字になっている

		svgPageModern := fmt.Sprintf(`
<svg xmlns="http://www.w3.org/2000/svg" width="100%%" height="100%%" viewBox="0 0 400 120">
    <!-- 背景: 角丸長方形 -->
    <rect x="0" y="0" rx="12" ry="12" width="400" height="120" fill="%v" stroke="#e0e0e0" stroke-width="1"/>
    
    <!-- 左側のアクセントカラーのエリア -->
    <rect x="0" y="0" rx="12" ry="12" width="12" height="120" fill="%v"/>

    <!-- 年齢/周年 (メインの強調表示: 右上) -->
    <text x="380" y="60" text-anchor="end" font-size="40" fill="%v" stroke="%v" stroke-width="2" font-weight="900" font-family="Inter, Meiryo, sans-serif">
        %v%v
    </text>

    <!-- 日数 (サブ情報: 右下) - ★ prefixTextは修正された値を使用 -->
    <text x="380" y="90" text-anchor="end" font-size="18" fill="%v" font-family="Meiryo, sans-serif">
        [%v%v日]
    </text>
    
    <!-- 干支 (左上: 絵文字のみを大きく表示) -->
    <text x="30" y="50" text-anchor="start" font-size="30" fill="%v" font-weight="bold" font-family="Meiryo, sans-serif">
        %v
    </text>

    <!-- 誕生日/開始日の説明 (左中: 分離) -->
    <text x="30" y="80" text-anchor="start" font-size="18" fill="%v" font-family="Meiryo, sans-serif">
        %v
    </text>

    <!-- タイトル/ロゴ (下部) - appNameが空の場合は表示されない -->
    <text x="30" y="105" text-anchor="start" font-size="14" fill="%v" font-family="Inter, Meiryo, sans-serif">
        %v  
    </text>
</svg>
		`,
			svgBGcolor,             // 1. 背景色
			pallet.BaseColor,       // 2. アクセントバーの色 (干支色)
			pallet.BaseColor,       // 3. 年齢の文字色 (干支色) **(FILL)**
			dynamicFontColor,       // 4. 年齢の縁取り色: 動的コントラストカラー **(STROKE)**
			ai.Age,                 // 5. 年齢
			unitText,               // 6. 単位 ("歳" or "周年")
			dynamicFontColor,       // 7. 日数文字色: 動的コントラストカラー
			prefixText,             // 8. 接頭辞 ("生後" or "生誕" or "")
			ai.TotalDate,           // 9. 日数
			pallet.BaseColor,       // 10. (使わないがプレースホルダに干支色を設定)
			ai.SexagenaryCycle,     // 11. 干支の絵文字 **(常に絵文字)**
			dynamicFontColor,       // 12. 日付の説明文字色: 動的コントラストカラー
			ai.BaseDateDescription, // 13. ベースの日付の説明 (例: "2001年1月5日生まれ")
			dynamicFontColor,       // 14. タイトル/ロゴ文字色: 動的コントラストカラー
			appName,                // 15. アプリ名/ロゴテキスト (空文字列の可能性あり)
		)

		fmt.Fprint(w, svgPageModern)
		return
	}

	// 2. SIMPLE CARD Mode: type=simplecard (コンパクトな複数行カード)
	if isSimpleCardMode {
		// Simple Card Modeは、ユーザーが提供したコンパクトなカードのデザイン（複数行）を再現します。

		// 背景色 (qColorが3CFの場合、#33CCFFになる)
		bgColor := fmt.Sprintf("#%s", strings.TrimPrefix(svgBGcolor, "#"))

		svgPageCompact := fmt.Sprintf(`
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" class="square" viewBox="0 0 100 100" width="%vpx" height="%vpx">
    <!-- 背景パス: fillをユーザー指定の背景色に設定 -->
    <path d="M0 0 L 100 0 L 100 100 L 0 0" style="fill:%v;stroke-width:0"/> 

    <!-- アクセントサークル (干支の色ベース) -->
    <circle cx="5" cy="5" r="40" fill="%v"/>

    <!-- Line 1: 干支絵文字 + 日付の説明 (例: 🐍 2001年1月5日生まれ) -->
    <text x="10" y="20" style="text-anchor:start;font-size:20px;fill:%v;font-family: Meiryo, Verdana, Helvetica, Arial, sans-serif;">
        %v
    </text>

    <!-- Line 2: (空行の再現) -->
    <text x="10" y="40" style="text-anchor:start;font-size:20px;fill:%v;font-family: Meiryo, Verdana, Helvetica, Arial, sans-serif;">
        
    </text> 
    
    <!-- Line 3: 年齢 + 日数 (例: 24歳[生誕9062日]) - 強調表示 -->
    <text x="15" y="60" style="text-anchor:start;font-size:20px;fill:%v;font-family: Meiryo, Verdana, Helvetica, Arial, sans-serif;">
        %v
    </text>
</svg>
		`,
			SimpleWidth, SimpleHeight, // width="270px" height="100px" (物理サイズ)
			bgColor,                   // 1. 背景色 (path fill)
			ai.SexagenaryCycleColor,   // 2. アクセントサークルの色 (干支の色)
			getFontColor(bgColor),     // 3. Line 1の文字色 (背景色に合わせて黒/白)
			ai.MultiText1,             // 4. Line 1の内容 (🐍 2001年1月5日生まれ)
			pallet.ComplementaryColor, // 5. Line 2の色 (サンプルではRGB(108,152,110))
			getFontColor(bgColor),     // 6. Line 3の文字色 (背景色に合わせて黒/白)
			ai.MultiText3,             // 7. Line 3の内容 (24歳[生誕9062日])
		)

		fmt.Fprint(w, svgPageCompact)
		return
	}

	// 3. LEGACY CARD Mode: type=legacy (従来のワイドな単一行フレーム)
	if isLegacyCardMode {
		// SVGの物理サイズ
		svgWidth := 1420
		svgHeight := 80
		frameWidth := 1320

		// ai.Textは Legacy/Wideモード向けの単一行テキストを使用

		// シャドウの色 (RGB(18,67,160) - ユーザーサンプルより)
		shadowColor := "RGB(18,67,160)"
		// メインテキストの色 (RGB(2,2,2) - ユーザーサンプルより)
		mainColor := "RGB(2,2,2)"

		svgPageLegacy := fmt.Sprintf(`
<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%v" height="%v">
    <!-- フレーム -->
    <rect x="%v" y="%v" rx="%v" ry="%v" width="%v" height="%v" stroke="%v" fill="transparent" stroke-width="10"/>
    
    <!-- テキストシャドウ（青っぽい色で下右にずらす） -->
    <text x="42" y="55" style="text-anchor:start;font-size:%vpx;fill:%v;font-family: Meiryo, Verdana, Helvetica, Arial, sans-serif;">          
        %v
    </text>
    
    <!-- メインテキスト（黒色でシャドウの上に重ねる） -->
    <text x="40" y="53" style="text-anchor:start;font-size:%vpx;fill:%v;font-family: Meiryo, Verdana, Helvetica, Arial, sans-serif;">
        %v
    </text>
</svg>
		`,
			svgWidth, svgHeight,
			FrameXY, FrameXY, rxy, rxy, frameWidth, FrameHeight, pallet.BaseColor, // フレーム
			FontSize, shadowColor, ai.Text, // シャドウテキスト
			FontSize, mainColor, ai.Text, // メインテキスト
		)

		fmt.Fprint(w, svgPageLegacy)
		return
	}

	// どのモードにもマッチしない場合の安全策（エラーページ）
	fmt.Fprint(w, svgPage)
}

// ----------------------------------------------------------------------
// 補助関数群
// ----------------------------------------------------------------------

// searchBirthDayは日付計算とテキスト整形を行う
func searchBirthDay(yyyymmdd string, itemTxt1 string, itemTxt2 string, itemTxt3 string, useCardEmoji bool) AgeInfo {

	// yyyymmddから日付を解析
	t, err := time.Parse("20060102", yyyymmdd)
	if err != nil {
		fmt.Printf("Error parsing date: %v\n", err)
		return AgeInfo{}
	}

	today := time.Now()

	// 時間差を計算
	duration := today.Sub(t)
	totalDate := int(duration.Hours() / 24)

	// 満年齢または経過年数を計算 (うるう年考慮)
	age := today.Year() - t.Year()
	// 誕生日/記念日が来ていない場合は年齢を1減らす
	if today.Month() < t.Month() || (today.Month() == t.Month() && today.Day() < t.Day()) {
		age--
	}

	// 干支と色を取得
	zodiac, color := getZodiac(t.Year())

	// ai.SexagenaryCycle には、モードに応じて表示したいテキストを設定する
	var displayZodiac string
	if useCardEmoji {
		// Simple Card Mode & Modern Mode (絵文字のみ)
		displayZodiac = zodiac.Emoji
	} else {
		// Legacy Mode (名称のみ) - (このロジックは使われなくなるが残しておく)
		displayZodiac = zodiac.Name
	}

	// AgeInfoを作成
	ai := AgeInfo{
		Age:                  age,
		TotalDate:            totalDate,
		BaseDate:             t.Format(layoutYMD),
		BaseDateDescription:  fmt.Sprintf(itemTxt1, t.Format(layoutYMD)),
		SexagenaryCycle:      displayZodiac, // "🐍" (絵文字)
		SexagenaryCycleColor: color,
	}

	// ★★★ [修正点] 総日数が1000日未満なら「生後」を維持し、1000日以上なら「生誕」に切り替える ★★★
	if totalDate < 1000 {
		// 1000日未満の場合は「生後」を維持
		ai.Text = fmt.Sprintf(itemTxt2, ai.Age, ai.TotalDate)
		ai.MultiText3 = fmt.Sprintf(itemTxt2, ai.Age, ai.TotalDate)
	} else {
		// 1000日以上の場合は「生誕」に置き換える
		modifiedItemTxt2 := strings.Replace(itemTxt2, "生後", "生誕", 1)

		ai.Text = fmt.Sprintf(modifiedItemTxt2, ai.Age, ai.TotalDate)
		ai.MultiText3 = fmt.Sprintf(modifiedItemTxt2, ai.Age, ai.TotalDate)
	}

	// Simple Card Mode の Line 1 (ai.MultiText1) にZodiac情報を含める (絵文字のみ)
	if useCardEmoji {
		ai.MultiText1 = fmt.Sprintf("%v %v", zodiac.Emoji, ai.BaseDateDescription) // "🐍 2001年1月5日生まれ"
	} else {
		// Legacy Mode のテキスト処理 (既存のまま)
		ai.MultiText1 = fmt.Sprintf(itemTxt3, ai.BaseDateDescription) // 【2001年1月5日生まれ】
	}

	return ai
}

// getZodiacは年号から干支（名称、絵文字、色）を取得する
func getZodiac(year int) (struct{ Name, Emoji string }, string) {
	zodiacs := []struct{ Name, Emoji, Color string }{
		{"申・さる", "🐵", "#8B4513"},   // Brown
		{"酉・とり", "🐔", "#FFD700"},   // Gold
		{"戌・いぬ", "🐕", "#A0522D"},   // Sienna
		{"亥・いのしし", "🐗", "#D2B48C"}, // Tan
		{"子・ねずみ", "🐀", "#C0C0C0"},  // Silver
		{"丑・うし", "🐂", "#4B0082"},   // Indigo
		{"寅・とら", "🐅", "#FFA500"},   // Orange
		{"卯・うさぎ", "🐇", "#F0FFFF"},  // Azure
		{"辰・たつ", "🐉", "#3CB371"},   // MediumSeaGreen
		{"巳・へび", "🐍", "#32CD32"},   // LimeGreen
		{"午・うま", "🐎", "#B22222"},   // Firebrick
		{"未・ひつじ", "🐐", "#F5F5DC"},  // Beige
	}
	// 1900年が子年を基準に、配列のインデックスを計算
	zodiacOrder := []int{4, 5, 6, 7, 8, 9, 10, 11, 0, 1, 2, 3} // 子, 丑, 寅, ... の順に配列インデックスをマッピング
	idx := (year - 1900) % 12
	if idx < 0 {
		idx += 12
	}

	finalIndex := zodiacOrder[idx]

	z := zodiacs[finalIndex]
	return struct{ Name, Emoji string }{z.Name, z.Emoji}, z.Color
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
func getFontColor(colorCode string) string {
	// 背景色の明るさを計算
	textColor := textColorFromBackgroundColor(colorCode)

	return textColor

}

// 背景色（6桁のカラーコード）から文字色を出す関数
func textColorFromBackgroundColor(bgColor string) string {

	purserRGB := func(bgColor string) (int, int, int) {
		// 背景色のRGB値を取得
		var red int
		var green int
		var blue int

		// カラーコードが3桁または6桁の場合の処理
		if len(bgColor) == 4 { // 例: #3CF
			// 3桁から6桁へ展開したものを解析
			expandedColor := fmt.Sprintf("#%c%c%c%c%c%c", bgColor[1], bgColor[1], bgColor[2], bgColor[2], bgColor[3], bgColor[3])
			fmt.Sscanf(expandedColor, "#%02x%02x%02x", &red, &green, &blue)
		} else if len(bgColor) == 7 { // 例: #33CCFF
			fmt.Sscanf(bgColor, "#%02x%02x%02x", &red, &green, &blue)
		} else if strings.HasPrefix(bgColor, "RGB") { // 例: RGB(255,255,255)
			// RGB(r,g,b)形式の場合、数値を抽出
			parts := strings.Trim(bgColor, "RGB()")
			nums := strings.Split(parts, ",")
			if len(nums) == 3 {
				red, _ = strconv.Atoi(strings.TrimSpace(nums[0]))
				green, _ = strconv.Atoi(strings.TrimSpace(nums[1]))
				blue, _ = strconv.Atoi(strings.TrimSpace(nums[2]))
			}
		}

		fmt.Printf("[purserRGB:%v]R:%v G:%v B:%v \n", bgColor, red, green, blue)

		return red, green, blue
	}
	// 3桁カラーコードの考慮（#3CF -> #33CCFF）
	processedBgColor := bgColor
	if len(bgColor) == 4 {
		processedBgColor = fmt.Sprintf("#%c%c%c%c%c%c", bgColor[1], bgColor[1], bgColor[2], bgColor[2], bgColor[3], bgColor[3])
	} else if len(bgColor) == 7 && strings.HasPrefix(bgColor, "#") {
		processedBgColor = bgColor
	} else {
		// #FF00FFのような標準的な形式ではない場合は、黒をデフォルトとします
		return "#000000"
	}

	r, g, b := purserRGB(processedBgColor)
	// 背景色の相対輝度を計算する
	bgLuminance := relativeLuminance(r, g, b)

	// 文字色の候補を定義する（白と黒のみで判断するのが一般的）
	textColors := []string{"#000000", "#FFFFFF"}

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
		// 4.5未満でも、最もコントラストの高い色を返す（最低限の視認性を確保）
		return bestTextColor
	}
}

// 相対輝度を計算する関数
func relativeLuminance(r, g, b int) float64 {
	fmt.Printf("[relativeLuminance]R:%v G:%v B:%v \n", r, g, b)
	var rs, gs, bs float64

	// sRGBから線形RGBへの変換を簡略化
	// 0-255を0-1に正規化
	normR := float64(r) / 255.0
	normG := float64(g) / 255.0
	normB := float64(b) / 255.0

	// sRGBから線形へ変換 (WCAG 2.1準拠の簡略化バージョン)
	if normR <= 0.03928 {
		rs = normR / 12.92
	} else {
		rs = math.Pow((normR+0.055)/1.055, 2.4)
	}
	if normG <= 0.03928 {
		gs = normG / 12.92
	} else {
		gs = math.Pow((normG+0.055)/1.055, 2.4)
	}
	if normB <= 0.03928 {
		bs = normB / 12.92
	} else {
		bs = math.Pow((normB+0.055)/1.055, 2.4)
	}

	// 相対輝度L
	return 0.2126*rs + 0.7152*gs + 0.0722*bs
}

// コントラスト比を計算する関数
func contrastRatio(l1, l2 float64) float64 {
	// L1, L2は相対輝度 (0.0-1.0)
	// コントラスト比 = (L1 + 0.05) / (L2 + 0.05)
	var darker, lighter float64
	if l1 > l2 {
		lighter = l1
		darker = l2
	} else {
		lighter = l2
		darker = l1
	}
	return (lighter + 0.05) / (darker + 0.05)
}
