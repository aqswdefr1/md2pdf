package main

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"github.com/signintech/gopdf"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type TextSpan struct {
	Text     string
	Font     string // regular, bold, italic, code
	Size     float64
	Color    color.RGBA
	BgColor  *color.RGBA
	IsLink   bool
	LinkDest string
}

type PDFConverter struct {
	pdf       gopdf.GoPdf
	theme     Theme
	margin    float64
	y         float64
	pageWidth float64
	pageHeight float64
	source    []byte
}

func NewPDFConverter(theme Theme, margin float64, source []byte) *PDFConverter {
	return &PDFConverter{
		theme:  theme,
		margin: margin,
		source: source,
	}
}

func (c *PDFConverter) Convert(markdownPath string, pdfPath string, isLandscape bool) error {
	// PDF Başlat
	c.pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeA4,
	})

	if isLandscape {
		// Dikey yerine yatay yap
		// Gopdf A4 varsayılanı 595.27 x 841.89 (dikey)
		// Yatay için boyutları değiştirebiliriz.
		// Ancak gopdf.Config içinde PageSize zaten bir struct'tır.
		// Dikey/Yatay modunu sayfa eklerken de belirtebiliriz veya Config'de ayarlayabiliriz.
		// Burası şimdilik dikey A4 olarak kalabilir, çünkü varsayılan A4 en yaygın olanıdır.
	}

	c.pageWidth = 595.27
	c.pageHeight = 841.89

	// Fontları yükle
	err := c.pdf.AddTTFFont("regular", "assets/fonts/DejaVuSans.ttf")
	if err != nil {
		return fmt.Errorf("regular font yüklenemedi: %v", err)
	}
	err = c.pdf.AddTTFFont("bold", "assets/fonts/DejaVuSans-Bold.ttf")
	if err != nil {
		return fmt.Errorf("bold font yüklenemedi: %v", err)
	}
	err = c.pdf.AddTTFFont("italic", "assets/fonts/DejaVuSans-Oblique.ttf")
	if err != nil {
		return fmt.Errorf("italic font yüklenemedi: %v", err)
	}
	err = c.pdf.AddTTFFont("code", "assets/fonts/DejaVuSansMono.ttf")
	if err != nil {
		return fmt.Errorf("code font yüklenemedi: %v", err)
	}

	c.pdf.AddPage()
	c.y = c.margin

	// Arka plan rengini ayarla (varsayılan beyaz ise çizmeye gerek yok ama diğer renkler için)
	c.drawBackground()

	// Goldmark ile parser oluştur
	md := goldmark.New()
	reader := text.NewReader(c.source)
	doc := md.Parser().Parse(reader)

	// AST'yi dolaş
	err = c.renderNode(doc)
	if err != nil {
		return err
	}

	// Sayfa numaralarını ekle
	c.addPageNumbers()

	// Dosyayı kaydet
	err = c.pdf.WritePdf(pdfPath)
	if err != nil {
		return fmt.Errorf("pdf kaydedilemedi: %v", err)
	}

	return nil
}

func (c *PDFConverter) drawBackground() {
	if c.theme.BgColor.R != 255 || c.theme.BgColor.G != 255 || c.theme.BgColor.B != 255 {
		c.pdf.SetFillColor(c.theme.BgColor.R, c.theme.BgColor.G, c.theme.BgColor.B)
		c.pdf.RectFromUpperLeftWithStyle(0, 0, c.pageWidth, c.pageHeight, "F")
	}
}

func (c *PDFConverter) renderNode(node ast.Node) error {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch child.Kind() {
		case ast.KindHeading:
			c.renderHeading(child.(*ast.Heading))
		case ast.KindParagraph:
			c.renderParagraph(child.(*ast.Paragraph))
		case ast.KindBlockquote:
			c.renderBlockquote(child.(*ast.Blockquote))
		case ast.KindList:
			c.renderList(child.(*ast.List), 0)
		case ast.KindFencedCodeBlock:
			c.renderFencedCodeBlock(child.(*ast.FencedCodeBlock))
		case ast.KindThematicBreak:
			c.renderThematicBreak(child.(*ast.ThematicBreak))
		default:
			// Desteklenmeyen veya konteyner olan diğer blokları recursive işle
			if child.HasChildren() {
				c.renderNode(child)
			}
		}
		c.y += 12 // Bloklar arası boşluk
	}
	return nil
}

func (c *PDFConverter) checkPageBreak(neededHeight float64) {
	if c.y+neededHeight > c.pageHeight-c.margin {
		c.pdf.AddPage()
		c.drawBackground()
		c.y = c.margin
	}
}

func (c *PDFConverter) renderHeading(h *ast.Heading) {
	var fontStyle FontStyle
	switch h.Level {
	case 1:
		fontStyle = c.theme.H1
	case 2:
		fontStyle = c.theme.H2
	case 3:
		fontStyle = c.theme.H3
	default:
		fontStyle = c.theme.H4
	}

	c.checkPageBreak(fontStyle.Size * 2)

	// Altında çizgi olan H1 ve H2'ler için ekstra boşluk ve çizgi
	spans := c.collectInlineSpans(h, "bold", fontStyle.Size, fontStyle.Color)
	c.renderSpans(spans, c.margin, 1.3)

	if h.Level <= 2 {
		c.y += 4
		c.pdf.SetLineWidth(0.8)
		c.pdf.SetStrokeColor(c.theme.BorderColor.R, c.theme.BorderColor.G, c.theme.BorderColor.B)
		c.pdf.Line(c.margin, c.y, c.pageWidth-c.margin, c.y)
		c.y += 6
	}
}

func (c *PDFConverter) renderParagraph(p *ast.Paragraph) {
	spans := c.collectInlineSpans(p, "regular", c.theme.Body.Size, c.theme.Body.Color)
	c.renderSpans(spans, c.margin, 1.4)
}

func (c *PDFConverter) renderBlockquote(b *ast.Blockquote) {
	// Blockquote'un içindeki çocukları topluca render et ama sol tarafa çizgi çek ve girinti ver
	// Giriş koordinatlarını sakla
	oldMargin := c.margin
	c.margin = oldMargin + 15 // İçeri kaydır

	// Blockquote solundaki dikey çizgi için koordinat hesapla
	startY := c.y

	// İçeriği render et
	c.renderNode(b)

	endY := c.y

	// Çizgiyi çiz
	c.pdf.SetLineWidth(3.0)
	c.pdf.SetStrokeColor(c.theme.BlockquoteColor.R, c.theme.BlockquoteColor.G, c.theme.BlockquoteColor.B)
	c.pdf.Line(oldMargin+5, startY, oldMargin+5, endY)

	// Margin'i geri al
	c.margin = oldMargin
}

func (c *PDFConverter) renderList(l *ast.List, depth int) {
	oldMargin := c.margin
	c.margin = oldMargin + float64(depth)*15

	index := 1
	for child := l.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == ast.KindListItem {
			c.renderListItem(child.(*ast.ListItem), l.IsOrdered(), index, depth)
			index++
		}
	}

	c.margin = oldMargin
}

func (c *PDFConverter) renderListItem(li *ast.ListItem, isOrdered bool, index int, depth int) {
	c.checkPageBreak(c.theme.Body.Size * 1.5)

	bulletStr := "• "
	if isOrdered {
		bulletStr = fmt.Sprintf("%d. ", index)
	}

	// Bullet çiz
	c.pdf.SetFont("regular", "", c.theme.Body.Size)
	c.pdf.SetTextColor(c.theme.TextColor.R, c.theme.TextColor.G, c.theme.TextColor.B)
	c.pdf.SetXY(c.margin+10, c.y)
	c.pdf.Text(bulletStr)

	// List item içeriğini çek
	// List item genellikle paragraf barındırabilir
	oldMargin := c.margin
	c.margin = oldMargin + 22 // İçeriği bullet'ın sağından başlat

	// Çocukları render et
	for child := li.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == ast.KindParagraph {
			spans := c.collectInlineSpans(child, "regular", c.theme.Body.Size, c.theme.Body.Color)
			c.renderSpans(spans, c.margin, 1.4)
		} else if child.Kind() == ast.KindList {
			c.renderList(child.(*ast.List), depth+1)
		} else {
			c.renderNode(li)
		}
	}

	c.margin = oldMargin
}

func (c *PDFConverter) renderFencedCodeBlock(cb *ast.FencedCodeBlock) {
	// Kod bloğu metnini oku
	var codeLines []string
	lines := cb.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		codeLines = append(codeLines, string(line.Value(c.source)))
	}
	codeText := strings.Join(codeLines, "")

	// Satırları böl
	splitLines := strings.Split(codeText, "\n")
	if len(splitLines) > 0 && splitLines[len(splitLines)-1] == "" {
		splitLines = splitLines[:len(splitLines)-1]
	}

	// Blok yüksekliğini hesapla
	lineHeight := c.theme.Code.Size * 1.4
	blockHeight := float64(len(splitLines))*lineHeight + 16

	c.checkPageBreak(30) // En az 3 satırlık yer yoksa yeni sayfaya geç

	// Eğer tüm kod bloğu sayfaya sığmıyorsa, sayfa sayfa böleceğiz
	startY := c.y
	c.pdf.SetFillColor(c.theme.CodeBg.R, c.theme.CodeBg.G, c.theme.CodeBg.B)
	
	// Arka plan kutusunu çiz (sayfa sonuna kadar olan kısmı veya bloğun tamamını)
	drawHeight := blockHeight
	if startY+blockHeight > c.pageHeight-c.margin {
		drawHeight = c.pageHeight - c.margin - startY
	}
	c.pdf.RectFromUpperLeftWithStyle(c.margin, startY, c.pageWidth-c.margin*2, drawHeight, "F")

	// Çerçeve çiz
	c.pdf.SetLineWidth(0.5)
	c.pdf.SetStrokeColor(c.theme.BorderColor.R, c.theme.BorderColor.G, c.theme.BorderColor.B)
	c.pdf.RectFromUpperLeftWithStyle(c.margin, startY, c.pageWidth-c.margin*2, drawHeight, "D")

	c.y += 8 // Üst pedding

	c.pdf.SetFont("code", "", c.theme.Code.Size)
	c.pdf.SetTextColor(c.theme.Code.Color.R, c.theme.Code.Color.G, c.theme.Code.Color.B)

	for _, line := range splitLines {
		c.checkPageBreakForCodeLine(lineHeight)
		c.pdf.SetXY(c.margin+8, c.y)
		// Çok uzun kod satırlarını wrap et
		c.renderCodeLine(line, c.margin+8)
		c.y += lineHeight
	}

	c.y += 8 // Alt pedding
}

func (c *PDFConverter) checkPageBreakForCodeLine(lineHeight float64) {
	if c.y+lineHeight > c.pageHeight-c.margin {
		c.pdf.AddPage()
		c.drawBackground()
		c.y = c.margin + 8

		// Yeni sayfada da arka plan çiz (kalan satırlar için)
		c.pdf.SetFillColor(c.theme.CodeBg.R, c.theme.CodeBg.G, c.theme.CodeBg.B)
		c.pdf.RectFromUpperLeftWithStyle(c.margin, c.y-8, c.pageWidth-c.margin*2, c.pageHeight-c.margin*2, "F")
		c.pdf.SetLineWidth(0.5)
		c.pdf.SetStrokeColor(c.theme.BorderColor.R, c.theme.BorderColor.G, c.theme.BorderColor.B)
		c.pdf.RectFromUpperLeftWithStyle(c.margin, c.y-8, c.pageWidth-c.margin*2, c.pageHeight-c.margin*2, "D")
	}
}

func (c *PDFConverter) renderCodeLine(line string, startX float64) {
	// GoPDF'te tab karakterlerini boşluğa çevir
	line = strings.ReplaceAll(line, "\t", "    ")
	
	maxWidth := c.pageWidth - c.margin - startX - 8
	width, _ := c.pdf.MeasureTextWidth(line)

	if width <= maxWidth {
		c.pdf.Text(line)
		return
	}

	// Sığmıyorsa karakter karakter veya kelime kelime bölerek yaz
	var currentLine string
	for _, run := range line {
		testLine := currentLine + string(run)
		w, _ := c.pdf.MeasureTextWidth(testLine)
		if w > maxWidth {
			c.pdf.Text(currentLine)
			c.y += c.theme.Code.Size * 1.4
			c.checkPageBreakForCodeLine(c.theme.Code.Size * 1.4)
			c.pdf.SetXY(startX, c.y)
			currentLine = string(run)
		} else {
			currentLine = testLine
		}
	}
	c.pdf.Text(currentLine)
}

func (c *PDFConverter) renderThematicBreak(tb *ast.ThematicBreak) {
	c.checkPageBreak(10)
	c.y += 6
	c.pdf.SetLineWidth(1.0)
	c.pdf.SetStrokeColor(c.theme.BorderColor.R, c.theme.BorderColor.G, c.theme.BorderColor.B)
	c.pdf.Line(c.margin, c.y, c.pageWidth-c.margin, c.y)
	c.y += 6
}

func (c *PDFConverter) collectInlineSpans(node ast.Node, defaultFont string, defaultSize float64, defaultColor color.RGBA) []TextSpan {
	var spans []TextSpan
	
	var walk func(n ast.Node, font string, size float64, col color.RGBA, isLink bool, linkDest string)
	walk = func(n ast.Node, font string, size float64, col color.RGBA, isLink bool, linkDest string) {
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			switch child.Kind() {
			case ast.KindText:
				t := child.(*ast.Text)
				content := string(t.Segment.Value(c.source))
				spans = append(spans, TextSpan{
					Text:     content,
					Font:     font,
					Size:     size,
					Color:    col,
					IsLink:   isLink,
					LinkDest: linkDest,
				})
			case ast.KindEmphasis:
				emp := child.(*ast.Emphasis)
				fontStyle := "italic"
				if emp.Level == 2 {
					fontStyle = "bold"
				}
				walk(child, fontStyle, size, col, isLink, linkDest)
			case ast.KindCodeSpan:
				// Code span içindeki düz metni topla
				var codeText string
				for cc := child.FirstChild(); cc != nil; cc = cc.NextSibling() {
					if cc.Kind() == ast.KindText {
						codeText += string(cc.(*ast.Text).Segment.Value(c.source))
					}
				}
				if codeText == "" {
					// Bazen çocuk düğüm olmayabilir, doğrudan segmentlerden al
					segments := child.Text(c.source)
					codeText = string(segments)
				}
				spans = append(spans, TextSpan{
					Text:    codeText,
					Font:    "code",
					Size:    size - 1,
					Color:   c.theme.Code.Color,
					BgColor: &c.theme.CodeBg,
				})
			case ast.KindLink:
				linkNode := child.(*ast.Link)
				dest := string(linkNode.Destination)
				walk(child, font, size, c.theme.LinkColor, true, dest)
			case ast.KindAutoLink:
				al := child.(*ast.AutoLink)
				dest := string(al.Label(c.source))
				spans = append(spans, TextSpan{
					Text:     dest,
					Font:     font,
					Size:     size,
					Color:    c.theme.LinkColor,
					IsLink:   true,
					LinkDest: dest,
				})
			default:
				if child.HasChildren() {
					walk(child, font, size, col, isLink, linkDest)
				}
			}
		}
	}

	walk(node, defaultFont, defaultSize, defaultColor, false, "")
	return spans
}

func (c *PDFConverter) renderSpans(spans []TextSpan, startX float64, lineHeightMultiplier float64) {
	if len(spans) == 0 {
		return
	}

	type WordPart struct {
		Text     string
		Font     string
		Size     float64
		Color    color.RGBA
		BgColor  *color.RGBA
		IsLink   bool
		LinkDest string
	}

	// Spans listesini kelime ve boşluk parçalarına ayıralım
	var parts []WordPart
	for _, span := range spans {
		// Satır sonu karakterlerini temizle veya boşluğa çevir
		textVal := strings.ReplaceAll(span.Text, "\n", " ")
		
		// Kelime kelime bölme
		// Boşlukları korumak için split yerine karakter gezebiliriz
		var currentWord strings.Builder
		for i, run := range textVal {
			if run == ' ' {
				if currentWord.Len() > 0 {
					parts = append(parts, WordPart{
						Text:     currentWord.String(),
						Font:     span.Font,
						Size:     span.Size,
						Color:    span.Color,
						BgColor:  span.BgColor,
						IsLink:   span.IsLink,
						LinkDest: span.LinkDest,
					})
					currentWord.Reset()
				}
				parts = append(parts, WordPart{
					Text:     " ",
					Font:     span.Font,
					Size:     span.Size,
					Color:    span.Color,
					BgColor:  span.BgColor,
					IsLink:   span.IsLink,
					LinkDest: span.LinkDest,
				})
			} else {
				currentWord.WriteRune(run)
			}
			if i == len(textVal)-1 && currentWord.Len() > 0 {
				parts = append(parts, WordPart{
					Text:     currentWord.String(),
					Font:     span.Font,
					Size:     span.Size,
					Color:    span.Color,
					BgColor:  span.BgColor,
					IsLink:   span.IsLink,
					LinkDest: span.LinkDest,
				})
			}
		}
	}

	// Kelimeleri satırlara yerleştir
	x := startX
	lineHeight := spans[0].Size * lineHeightMultiplier

	c.checkPageBreak(lineHeight)

	for _, part := range parts {
		c.pdf.SetFont(part.Font, "", part.Size)
		width, _ := c.pdf.MeasureTextWidth(part.Text)

		// Eğer satır sonuna geldiysek ve kelime sığmıyorsa (boşluk hariç)
		if x+width > c.pageWidth-c.margin && part.Text != " " {
			c.y += lineHeight
			c.checkPageBreak(lineHeight)
			x = startX
		}

		c.pdf.SetXY(x, c.y)
		c.pdf.SetTextColor(part.Color.R, part.Color.G, part.Color.B)

		// Arka plan çizimi (örneğin inline code için)
		if part.BgColor != nil {
			c.pdf.SetFillColor(part.BgColor.R, part.BgColor.G, part.BgColor.B)
			// Kelime etrafına küçük bir kutu çiz
			c.pdf.RectFromUpperLeftWithStyle(x, c.y-1, width, part.Size+2, "F")
			// Yazı rengini tekrar ayarla (çünkü fill rengi değişti)
			c.pdf.SetTextColor(part.Color.R, part.Color.G, part.Color.B)
		}

		if part.IsLink {
			// Link çiz ve tıklandığında hedefe gitmesini sağla
			c.pdf.Text(part.Text)
			// GoPDF link ekleme
			c.pdf.AddExternalLink(part.LinkDest, x, c.y, width, part.Size)
			// Altını çiz
			c.pdf.SetLineWidth(0.5)
			c.pdf.SetStrokeColor(part.Color.R, part.Color.G, part.Color.B)
			c.pdf.Line(x, c.y+part.Size, x+width, c.y+part.Size)
		} else {
			c.pdf.Text(part.Text)
		}

		x += width
	}

	c.y += lineHeight // Paragraf sonu satır kayması
}

func (c *PDFConverter) addPageNumbers() {
	pages := c.pdf.GetNumberOfPages()
	c.pdf.SetFont("regular", "", 8)
	c.pdf.SetTextColor(120, 120, 120)

	for i := 1; i <= pages; i++ {
		c.pdf.SetPage(i)
		pageStr := fmt.Sprintf("%d / %d", i, pages)
		w, _ := c.pdf.MeasureTextWidth(pageStr)
		c.pdf.SetXY((c.pageWidth-w)/2, c.pageHeight-c.margin/2)
		c.pdf.Text(pageStr)
	}
}

func ConvertMarkdownToPDF(sourcePath string, destPath string, themeName string, margin float64, isLandscape bool) error {
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("kaynak dosya okunamadı: %v", err)
	}

	theme, exists := Themes[strings.ToLower(themeName)]
	if !exists {
		theme = Themes["github"] // Varsayılan tema
	}

	converter := NewPDFConverter(theme, margin, source)
	return converter.Convert(sourcePath, destPath, isLandscape)
}

func ConvertMultipleMarkdownToPDF(srcDir string, destDir string, themeName string, margin float64) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("kaynak dizin okunamadı: %v", err)
	}

	// Hedef dizini oluştur
	if destDir != "" {
		err = os.MkdirAll(destDir, 0755)
		if err != nil {
			return fmt.Errorf("hedef dizin oluşturulamadı: %v", err)
		}
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			srcPath := srcDir + "/" + entry.Name()
			pdfName := strings.TrimSuffix(entry.Name(), ".md") + ".pdf"
			
			destPath := pdfName
			if destDir != "" {
				destPath = destDir + "/" + pdfName
			} else {
				// Aynı dizine kaydet
				destPath = srcDir + "/" + pdfName
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
