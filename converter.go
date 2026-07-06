package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type PDFConverter struct {
	theme  Theme
	margin float64
	source []byte
}

func NewPDFConverter(theme Theme, margin float64, source []byte) *PDFConverter {
	return &PDFConverter{
		theme:  theme,
		margin: margin,
		source: source,
	}
}

func (c *PDFConverter) Convert(markdownPath string, pdfPath string, isLandscape bool) error {
	// 1. Markdown'ı HTML'e dönüştür
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // GitHub Flavored Markdown (tablolar, autolink vb.)
			extension.Footnote,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	var htmlBuf strings.Builder
	if err := md.Convert(c.source, &htmlBuf); err != nil {
		return fmt.Errorf("markdown html'e dönüştürülemedi: %v", err)
	}

	htmlContent := htmlBuf.String()

	// 2. Alert kutularını HTML içinde düzenle
	htmlContent = convertAlertsToHTML(htmlContent)

	// 3. Kapak sayfasını ayrıştır
	coverHTML, remainingHTML := extractCoverPage(htmlContent)

	// 4. CSS ve Şablon ekle
	marginMm := fmt.Sprintf("%dmm", int(c.margin * 0.352778)) // point'ten mm'ye dönüşüm
	if c.margin == 50.0 {
		marginMm = "20mm" // Varsayılan değer
	}
	
	themeCSS := getThemeCSS(c.theme.Name, marginMm, isLandscape)
	
	fullHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<style>
		%s
	</style>
</head>
<body>
	<div class="markdown-body">
		%s
		<div class="content-page">
			%s
		</div>
	</div>
</body>
</html>`, themeCSS, coverHTML, remainingHTML)

	// 4. Geçici bir HTML dosyası yaz
	tempHTMLPath := pdfPath + ".temp.html"
	err := os.WriteFile(tempHTMLPath, []byte(fullHTML), 0644)
	if err != nil {
		return fmt.Errorf("geçici html yazılamadı: %v", err)
	}
	defer os.Remove(tempHTMLPath)

	// 5. Chromedp ile headless tarayıcıyı başlat ve PDF yazdır
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// 30 saniye timeout
	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	// Mutlak yol bul
	absHTMLPath, err := filepath.Abs(tempHTMLPath)
	if err != nil {
		return fmt.Errorf("dosya yolu çözülemedi: %v", err)
	}
	url := "file:///" + filepath.ToSlash(absHTMLPath)

	var pdfBuffer []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				WithLandscape(isLandscape).
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfBuffer = buf
			return nil
		}),
	)
	if err != nil {
		return fmt.Errorf("chrome pdf yazdıramadı: %v", err)
	}

	// 6. PDF dosyasını diske kaydet
	err = os.WriteFile(pdfPath, pdfBuffer, 0644)
	if err != nil {
		return fmt.Errorf("pdf kaydedilemedi: %v", err)
	}

	return nil
}

func convertAlertsToHTML(html string) string {
	re := regexp.MustCompile(`(?is)<blockquote>\s*<p>\s*\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*(.*?)</p>\s*(.*?)</blockquote>`)
	
	return re.ReplaceAllStringFunc(html, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		alertType := strings.ToUpper(submatches[1])
		firstPara := submatches[2]
		remaining := submatches[3]
		
		var title string
		switch alertType {
		case "NOTE": title = "NOT"
		case "TIP": title = "İPUCU"
		case "IMPORTANT": title = "ÖNEMLİ"
		case "WARNING": title = "UYARI"
		case "CAUTION": title = "DİKKAT"
		default: title = alertType
		}
		
		classType := strings.ToLower(alertType)
		fullContent := fmt.Sprintf("<p>%s</p>%s", firstPara, remaining)
		
		return fmt.Sprintf(`<div class="markdown-alert markdown-alert-%s"><p class="markdown-alert-title">%s</p>%s</div>`, classType, title, fullContent)
	})
}

func getThemeCSS(themeName string, marginMm string, isLandscape bool) string {
	pageSize := "A4 portrait"
	if isLandscape {
		pageSize = "A4 landscape"
	}

	commonCSS := fmt.Sprintf(`
		@page {
			size: %s;
			margin: %s;
			@bottom-right {
				content: counter(page);
				font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
				font-size: 8pt;
				color: #718096;
			}
		}
		@page :first {
			@bottom-right {
				content: normal; /* Kapak sayfasında sayfa numarasını gizle */
			}
		}
		body {
			line-height: 1.6;
			word-wrap: break-word;
			margin: 0;
			padding: 0;
		}
		
		/* Kapak Sayfasi Tasarimi */
		.cover-page {
			height: 85vh;
			display: flex;
			flex-direction: column;
			justify-content: space-between;
			page-break-after: always;
			box-sizing: border-box;
			padding: 60px 0;
		}
		.cover-center {
			margin-top: auto;
			margin-bottom: auto;
			text-align: center;
		}
		.cover-title {
			font-family: "Inter", system-ui, -apple-system, sans-serif;
			font-size: 28pt;
			font-weight: 800;
			color: #0f172a;
			line-height: 1.25;
			margin-bottom: 20px;
		}
		.cover-divider {
			width: 80px;
			height: 4px;
			background-color: #6366f1;
			margin: 24px auto;
			border-radius: 2px;
		}
		.cover-subtitle {
			font-size: 13pt;
			color: #475569;
			max-width: 550px;
			margin: 0 auto;
			line-height: 1.5;
		}
		.cover-footer {
			margin-top: auto;
			text-align: right;
			border-top: 1px solid #e2e8f0;
			padding-top: 20px;
		}
		.cover-date {
			font-size: 10pt;
			color: #64748b;
			font-weight: 500;
		}
		.content-page {
			page-break-before: always;
		}
		
		/* GFM Tablo Tasarımı */
		table {
			border-spacing: 0;
			border-collapse: collapse;
			width: 100%%;
			margin-top: 12px;
			margin-bottom: 20px;
			page-break-inside: avoid;
		}
		table th, table td {
			padding: 10px 14px;
			border: 1px solid #e2e8f0;
			text-align: left;
		}
		table tr {
			background-color: #ffffff;
		}
		
		/* Kod Blokları */
		pre {
			padding: 16px;
			overflow: auto;
			font-size: 85%%;
			line-height: 1.45;
			background-color: #f6f8fa;
			border-radius: 6px;
			border: 1px solid #d0d7de;
			word-wrap: normal;
		}
		code {
			font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace;
			font-size: 85%%;
			margin: 0;
			padding: .2em .4em;
			background-color: rgba(175,184,193,0.2);
			border-radius: 6px;
		}
		pre code {
			background-color: transparent;
			padding: 0;
			font-size: 100%%;
			word-break: normal;
		}
		
		/* Blockquotes */
		blockquote {
			padding: 0 1em;
			color: #475569;
			border-left: .25em solid #cbd5e1;
			margin: 0 0 20px 0;
		}
		blockquote p {
			margin-top: 0;
			margin-bottom: 8px;
		}
		blockquote p:last-child {
			margin-bottom: 0;
		}
		
		/* Alert Kutuları */
		.markdown-alert {
			padding: 12px 16px;
			margin-bottom: 20px;
			border-left: .25em solid #cbd5e1;
			border-radius: 0 6px 6px 0;
			background-color: #f8fafc;
			page-break-inside: avoid;
		}
		.markdown-alert-title {
			display: flex;
			align-items: center;
			font-weight: 600;
			font-size: 14px;
			margin-top: 0;
			margin-bottom: 6px;
		}
		.markdown-alert-note {
			border-left-color: #0969da;
			background-color: #f0f7ff;
		}
		.markdown-alert-note .markdown-alert-title {
			color: #0969da;
		}
		.markdown-alert-important {
			border-left-color: #8250df;
			background-color: #fbefff;
		}
		.markdown-alert-important .markdown-alert-title {
			color: #8250df;
		}
		.markdown-alert-warning {
			border-left-color: #9a6700;
			background-color: #fff8ec;
		}
		.markdown-alert-warning .markdown-alert-title {
			color: #9a6700;
		}
		.markdown-alert-caution {
			border-left-color: #cf222e;
			background-color: #fff0f0;
		}
		.markdown-alert-caution .markdown-alert-title {
			color: #cf222e;
		}
		.markdown-alert-tip {
			border-left-color: #1a7f37;
			background-color: #f0fdf4;
		}
		.markdown-alert-tip .markdown-alert-title {
			color: #1a7f37;
		}
		
		/* Dipnotlar */
		.footnotes {
			margin-top: 40px;
			padding-top: 20px;
			border-top: 1px solid #cbd5e1;
			font-size: 9pt;
			color: #475569;
		}
		.footnotes ol {
			padding-left: 20px;
		}
		.footnotes li {
			margin-bottom: 8px;
		}
		.footnotes li p {
			margin-bottom: 0;
			display: inline;
		}
	`, pageSize, marginMm)

	specificCSS := `
		body {
			font-family: "Inter", system-ui, -apple-system, sans-serif;
			color: #1e293b; /* Slate 800 */
			font-size: 10.5pt;
			line-height: 1.6;
		}
		h1, h2, h3, h4, h5, h6 {
			color: #0f172a; /* Slate 900 */
			font-weight: 700;
			margin-top: 28px;
			margin-bottom: 14px;
		}
		h1 {
			font-size: 24pt;
			border-bottom: 2px solid #6366f1; /* Indigo 500 */
			padding-bottom: 10px;
			color: #4f46e5; /* Indigo 600 */
		}
		h2 {
			font-size: 16pt;
			border-bottom: 1px solid #e2e8f0;
			padding-bottom: 6px;
		}
		h3 {
			font-size: 13pt;
		}
		p {
			margin-top: 0;
			margin-bottom: 16px;
			text-align: justify;
		}
		table {
			border-radius: 8px;
			overflow: hidden;
			border: 1px solid #e2e8f0;
			box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
		}
		table th {
			background-color: #f8fafc; /* Slate 50 */
			color: #0f172a;
			font-weight: 600;
			border-bottom: 2px solid #e2e8f0;
		}
		table td {
			border-bottom: 1px solid #f1f5f9;
		}
		table tr:nth-child(even) {
			background-color: #f8fafc;
		}
		blockquote {
			border-left: 4px solid #6366f1; /* Indigo 500 */
			background-color: #f8fafc; /* Slate 50 */
			padding: 12px 16px;
			border-radius: 0 8px 8px 0;
		}
	`

	return commonCSS + specificCSS
}

func ConvertMarkdownToPDF(sourcePath string, destPath string, themeName string, margin float64, isLandscape bool) error {
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("kaynak dosya okunamadı: %v", err)
	}

	theme, exists := Themes[strings.ToLower(themeName)]
	if !exists {
		theme = Themes["modern"] // Varsayılan tema
	}

	converter := NewPDFConverter(theme, margin, source)
	return converter.Convert(sourcePath, destPath, isLandscape)
}

func ConvertMultipleMarkdownToPDF(srcDir string, destDir string, themeName string, margin float64) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("kaynak dizin okunamadı: %v", err)
	}

	if destDir != "" {
		err = os.MkdirAll(destDir, 0755)
		if err != nil {
			return fmt.Errorf("hedef dizin oluşturulamadı: %v", err)
		}
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			srcPath := filepath.Join(srcDir, entry.Name())
			pdfName := strings.TrimSuffix(entry.Name(), ".md") + ".pdf"
			
			var destPath string
			if destDir != "" {
				destPath = filepath.Join(destDir, pdfName)
			} else {
				destPath = filepath.Join(srcDir, pdfName)
			}

			fmt.Printf("Dönüştürülüyor: %s -> %s\n", srcPath, destPath)
			err = ConvertMarkdownToPDF(srcPath, destPath, themeName, margin, false)
			if err != nil {
				fmt.Printf("Hata: %s dönüştürülemedi: %v\n", srcPath, err)
			}
		}
	}
	return nil
}

func extractCoverPage(html string) (string, string) {
	h1Re := regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`)
	h1Match := h1Re.FindStringSubmatch(html)
	if len(h1Match) == 0 {
		return "", html
	}
	
	title := h1Match[1]
	h1Loc := h1Re.FindStringIndex(html)
	afterH1 := html[h1Loc[1]:]
	
	pRe := regexp.MustCompile(`(?is)^\s*<p[^>]*>(.*?)</p>`)
	pMatch := pRe.FindStringSubmatch(afterH1)
	
	subtitle := ""
	remainingHTML := afterH1
	if len(pMatch) > 0 {
		subtitle = pMatch[1]
		pLoc := pRe.FindStringIndex(afterH1)
		remainingHTML = afterH1[pLoc[1]:]
	}
	
	hrRe := regexp.MustCompile(`^\s*<hr[^>]*>`)
	if hrRe.MatchString(remainingHTML) {
		loc := hrRe.FindStringIndex(remainingHTML)
		remainingHTML = remainingHTML[loc[1]:]
	}
	
	coverHTML := fmt.Sprintf(`
	<div class="cover-page">
		<div class="cover-center">
			<h1 class="cover-title">%s</h1>
			<div class="cover-divider"></div>
			%s
		</div>
		<div class="cover-footer">
			<span class="cover-date">%s</span>
		</div>
	</div>
	`, title, formatSubtitle(subtitle), time.Now().Format("02.01.2006"))
	
	return coverHTML, remainingHTML
}

func formatSubtitle(sub string) string {
	if sub == "" {
		return ""
	}
	return fmt.Sprintf(`<p class="cover-subtitle">%s</p>`, sub)
}

