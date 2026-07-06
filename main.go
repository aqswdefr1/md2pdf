package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fsnotify/fsnotify"
)

func main() {
	input := flag.String("input", "", "Giriş markdown dosyası (.md) veya dizini")
	output := flag.String("output", "", "Çıkış PDF dosyası (.pdf) veya dizini")
	theme := flag.String("theme", "modern", "Tema seçimi (sadece modern mevcuttur)")
	margin := flag.Float64("margin", 50.0, "Kenar boşluğu (point birimiyle)")
	watch := flag.Bool("watch", false, "Canlı izleme modu (dosya değiştikçe PDF güncellenir)")
	landscape := flag.Bool("landscape", false, "Yatay sayfa yönlendirmesi")
	interactive := flag.Bool("interactive", false, "İnteraktif CLI modu")
	flag.BoolVar(interactive, "i", false, "İnteraktif CLI modu (kısayol)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Kullanım: md2pdf [seçenekler]\n\nSeçenekler:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Eğer interaktif mod seçilmişse veya hiçbir flag girilmemişse
	if *interactive || (len(os.Args) == 1 && *input == "") {
		runInteractiveMode()
		return
	}

	if *input == "" {
		fmt.Println("Hata: Lütfen giriş dosyası veya dizini belirtin (-input)")
		flag.Usage()
		os.Exit(1)
	}

	// Giriş yolunun varlığını kontrol et
	fileInfo, err := os.Stat(*input)
	if err != nil {
		fmt.Printf("Hata: Giriş yolu bulunamadı: %v\n", err)
		os.Exit(1)
	}

	isDir := fileInfo.IsDir()

	// Tek dosya veya toplu dönüştürme yap
	if isDir {
		if *watch {
			fmt.Println("Hata: Dizin izleme (watch modu) şimdilik sadece tek dosya için desteklenmektedir.")
			os.Exit(1)
		}
		runBatchConversion(*input, *output, *theme, *margin)
	} else {
		// Tek dosya dönüştürme
		destPath := *output
		if destPath == "" {
			// Giriş dosyasıyla aynı isimde ama .pdf uzantılı yap
			destPath = strings.TrimSuffix(*input, filepath.Ext(*input)) + ".pdf"
		}

		runSingleConversion(*input, destPath, *theme, *margin, *landscape)

		if *watch {
			runWatcher(*input, destPath, *theme, *margin, *landscape)
		}
	}
}

func runSingleConversion(src, dest, theme string, margin float64, landscape bool) {
	fmt.Printf("[%s] Dönüştürülüyor: %s -> %s\n", time.Now().Format("15:04:05"), src, dest)
	start := time.Now()
	err := ConvertMarkdownToPDF(src, dest, theme, margin, landscape)
	if err != nil {
		fmt.Printf("Dönüştürme hatası: %v\n", err)
		return
	}
	fmt.Printf("Başarıyla tamamlandı! Süre: %v\n", time.Since(start))
}

func runBatchConversion(srcDir, destDir, theme string, margin float64) {
	fmt.Printf("Toplu dönüştürme başlatıldı: %s\n", srcDir)
	start := time.Now()
	err := ConvertMultipleMarkdownToPDF(srcDir, destDir, theme, margin)
	if err != nil {
		fmt.Printf("Toplu dönüştürme hatası: %v\n", err)
		return
	}
	fmt.Printf("Toplu dönüştürme tamamlandı! Toplam süre: %v\n", time.Since(start))
}

func runWatcher(src, dest, theme string, margin float64, landscape bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Watcher oluşturulamadı: %v\n", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(src)
	if err != nil {
		fmt.Printf("Dosya izlenemedi: %v\n", err)
		return
	}

	fmt.Printf("\n[Canlı İzleme Aktif] %s izleniyor. Değişiklik yaptığınızda PDF otomatik güncellenecektir...\n", src)
	fmt.Println("Durdurmak için Ctrl+C tuşlarına basın.")

	// Debounce işlemi için son değişiklik zamanını tutalım (editörler bazen peş peşe event tetikler)
	var lastEventTime time.Time

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Sadece yazma/güncelleme olaylarını yakala
			if event.Op&fsnotify.Write == fsnotify.Write {
				// 500ms debounce
				if time.Since(lastEventTime) > 500*time.Millisecond {
					lastEventTime = time.Now()
					fmt.Printf("\n[%s] Değişiklik algılandı, PDF güncelleniyor...\n", time.Now().Format("15:04:05"))
					runSingleConversion(src, dest, theme, margin, landscape)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher hatası: %v\n", err)
		}
	}
}

func runInteractiveMode() {
	fmt.Println("=========================================")
	fmt.Println("    md2pdf İnteraktif Sihirbazı          ")
	fmt.Println("=========================================")
	fmt.Println("İpucu: Dosyanızı bu terminal penceresine sürükleyip bırakabilirsiniz.")
	fmt.Println()

	var inputPath string
	inputPrompt := &survey.Input{
		Message: "Giriş Markdown dosyası veya klasörünün yolu:",
	}
	err := survey.AskOne(inputPrompt, &inputPath, survey.WithValidator(survey.Required))
	if err != nil {
		fmt.Println("İşlem iptal edildi:", err)
		return
	}

	// Sürükle-bırak tırnaklarını temizle
	inputPath = strings.Trim(inputPath, "\"'` ")

	// Giriş dosyasının/klasörünün varlığını kontrol et
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		fmt.Printf("Hata: Belirtilen yol bulunamadı: %v\n", err)
		return
	}
	isDir := fileInfo.IsDir()

	// Çıkış yolu sorusu
	var outputPath string
	outputPrompt := &survey.Input{
		Message: "Çıkış PDF dosyasının/klasörünün yolu (Varsayılan için boş bırakın):",
	}
	err = survey.AskOne(outputPrompt, &outputPath)
	if err != nil {
		fmt.Println("İşlem iptal edildi:", err)
		return
	}
	outputPath = strings.Trim(outputPath, "\"'` ")

	theme := "modern"

	// Kenar boşluğu (Margin)
	var marginStr string
	marginPrompt := &survey.Input{
		Message: "Kenar boşluğu değeri (point):",
		Default: "50",
	}
	err = survey.AskOne(marginPrompt, &marginStr)
	if err != nil {
		fmt.Println("İşlem iptal edildi:", err)
		return
	}
	margin, err := strconv.ParseFloat(marginStr, 64)
	if err != nil {
		fmt.Println("Geçersiz değer girildi, varsayılan (50.0) kullanılacak.")
		margin = 50.0
	}

	// Yönlendirme (Dikey / Yatay)
	var orientation string
	orientationPrompt := &survey.Select{
		Message: "Sayfa yönlendirmesi:",
		Options: []string{"Dikey (Portrait)", "Yatay (Landscape)"},
		Default: "Dikey (Portrait)",
	}
	err = survey.AskOne(orientationPrompt, &orientation)
	if err != nil {
		fmt.Println("İşlem iptal edildi:", err)
		return
	}
	landscape := (orientation == "Yatay (Landscape)")

	// Watch modu (Sadece dosya ise)
	var watch bool
	if !isDir {
		watchPrompt := &survey.Confirm{
			Message: "Dosya değişikliklerini canlı izlemek ister misiniz? (Watch Mode)",
			Default: false,
		}
		err = survey.AskOne(watchPrompt, &watch)
		if err != nil {
			fmt.Println("İşlem iptal edildi:", err)
			return
		}
	}

	// Çalıştırma aşaması
	if isDir {
		runBatchConversion(inputPath, outputPath, theme, margin)
	} else {
		destPath := outputPath
		if destPath == "" {
			destPath = strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".pdf"
		}
		runSingleConversion(inputPath, destPath, theme, margin, landscape)
		if watch {
			runWatcher(inputPath, destPath, theme, margin, landscape)
		}
	}
}

