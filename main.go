package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	input := flag.String("input", "", "Giriş markdown dosyası (.md) veya dizini")
	output := flag.String("output", "", "Çıkış PDF dosyası (.pdf) veya dizini")
	theme := flag.String("theme", "modern", "Tema seçimi (sadece modern mevcuttur)")
	margin := flag.Float64("margin", -1.0, "Kenar boşluğu (point birimiyle, varsayılan: otomatik)")
	watch := flag.Bool("watch", false, "Canlı izleme modu (dosya değiştikçe PDF güncellenir)")
	landscape := flag.Bool("landscape", false, "Yatay sayfa yönlendirmesi")
	interactive := flag.Bool("interactive", false, "İnteraktif CLI modu")
	flag.BoolVar(interactive, "i", false, "İnteraktif CLI modu (kısayol)")
	cover := flag.Bool("cover", true, "Kapak sayfası oluşturulsun mu?")

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
		runBatchConversion(*input, *output, *theme, *margin, *cover)
	} else {
		// Tek dosya dönüştürme
		destPath := *output
		if destPath == "" {
			// Giriş dosyasıyla aynı isimde ama .pdf uzantılı yap
			destPath = strings.TrimSuffix(*input, filepath.Ext(*input)) + ".pdf"
		}

		runSingleConversion(*input, destPath, *theme, *margin, *landscape, *cover)

		if *watch {
			runWatcher(*input, destPath, *theme, *margin, *landscape, *cover)
		}
	}
}

func runSingleConversion(src, dest, theme string, margin float64, landscape bool, cover bool) {
	fmt.Printf("\n⚡ [%s] Dönüştürülüyor:\n   ➔ Giriş: %s\n   ➔ Çıkış: %s\n", time.Now().Format("15:04:05"), src, dest)
	start := time.Now()
	err := ConvertMarkdownToPDF(src, dest, theme, margin, landscape, cover)
	if err != nil {
		fmt.Printf(" ❌ Dönüştürme hatası: %v\n", err)
		return
	}
	fmt.Printf(" ✔  Başarıyla tamamlandı! (Süre: %v)\n", time.Since(start).Round(time.Millisecond))
}

func runBatchConversion(srcDir, destDir, theme string, margin float64, cover bool) {
	fmt.Printf("\n⚡ Toplu dönüştürme başlatıldı: %s\n", srcDir)
	start := time.Now()
	err := ConvertMultipleMarkdownToPDF(srcDir, destDir, theme, margin, cover)
	if err != nil {
		fmt.Printf(" ❌ Toplu dönüştürme hatası: %v\n", err)
		return
	}
	fmt.Printf(" ✔  Toplu dönüştürme tamamlandı! (Toplam süre: %v)\n", time.Since(start).Round(time.Millisecond))
}

func runWatcher(src, dest, theme string, margin float64, landscape bool, cover bool) {
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
					runSingleConversion(src, dest, theme, margin, landscape, cover)
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
	fmt.Println("┌────────────────────────────────────────────────────────┐")
	fmt.Println("│             📄 md2pdf İnteraktif Sihirbazı             │")
	fmt.Println("└────────────────────────────────────────────────────────┘")
	fmt.Println(" ℹ  İpucu: Dosyanızı bu terminal penceresine sürükleyip bırakabilirsiniz.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// 1. Giriş Yolu
	var inputPath string
	for {
		fmt.Print(" ➔  Giriş Markdown dosyası veya klasörünün yolu:\n    » ")
		input, _ := reader.ReadString('\n')
		inputPath = strings.TrimSpace(input)
		inputPath = strings.Trim(inputPath, "\"'` ") // Sürükle-bırak tırnaklarını temizle
		if inputPath != "" {
			break
		}
		fmt.Println(" ❌ Hata: Giriş yolu boş olamaz.\n")
	}

	// Giriş dosyasının/klasörünün varlığını kontrol et
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		fmt.Printf(" ❌ Hata: Belirtilen yol bulunamadı: %v\n", err)
		return
	}
	isDir := fileInfo.IsDir()

	// 2. Çıkış Yolu
	fmt.Print(" ➔  Çıkış PDF yolu (Varsayılan için Enter'a basın):\n    » ")
	outputPath, _ := reader.ReadString('\n')
	outputPath = strings.TrimSpace(outputPath)
	outputPath = strings.Trim(outputPath, "\"'` ")

	theme := "modern"

	// 3. Kenar Boşluğu
	fmt.Print(" ➔  Kenar boşluğu değeri (Point birimiyle, Otomatik için Enter):\n    » ")
	marginStr, _ := reader.ReadString('\n')
	marginStr = strings.TrimSpace(marginStr)
	margin := -1.0
	if marginStr != "" {
		m, err := strconv.ParseFloat(marginStr, 64)
		if err == nil {
			margin = m
		} else {
			fmt.Println(" ⚠️  Geçersiz değer girildi, otomatik kullanılacak.")
		}
	}

	// 4. Sayfa Yönlendirmesi
	fmt.Print(" ➔  Sayfa yönlendirmesi [1: Dikey (Portrait), 2: Yatay (Landscape)] (Varsayılan: 1):\n    » ")
	orientationStr, _ := reader.ReadString('\n')
	orientationStr = strings.TrimSpace(orientationStr)
	landscape := false
	if orientationStr == "2" {
		landscape = true
	}

	// 5. Kapak Sayfası
	fmt.Print(" ➔  Kapak sayfası oluşturulsun mu? [E/h] (Varsayılan: E):\n    » ")
	coverStr, _ := reader.ReadString('\n')
	coverStr = strings.TrimSpace(strings.ToLower(coverStr))
	cover := true
	if coverStr == "h" || coverStr == "hayır" || coverStr == "n" || coverStr == "no" {
		cover = false
	}

	// 6. Watch modu (Sadece dosya ise)
	watch := false
	if !isDir {
		fmt.Print(" ➔  Değişiklikleri canlı izlemek ister misiniz? (Watch Mode) [e/H] (Varsayılan: H):\n    » ")
		watchStr, _ := reader.ReadString('\n')
		watchStr = strings.TrimSpace(strings.ToLower(watchStr))
		if watchStr == "e" || watchStr == "evet" || watchStr == "y" || watchStr == "yes" {
			watch = true
		}
	}

	fmt.Println("\n⏳ İşlem başlatılıyor...")

	// Çalıştırma aşaması
	if isDir {
		runBatchConversion(inputPath, outputPath, theme, margin, cover)
	} else {
		destPath := outputPath
		if destPath == "" {
			destPath = strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".pdf"
		}
		runSingleConversion(inputPath, destPath, theme, margin, landscape, cover)
		if watch {
			runWatcher(inputPath, destPath, theme, margin, landscape, cover)
		}
	}
}

