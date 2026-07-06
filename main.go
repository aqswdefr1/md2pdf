package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	input := flag.String("input", "", "Giriş markdown dosyası (.md) veya dizini")
	output := flag.String("output", "", "Çıkış PDF dosyası (.pdf) veya dizini")
	theme := flag.String("theme", "github", "Tema seçimi: github, modern, academic")
	margin := flag.Float64("margin", 50.0, "Kenar boşluğu (point birimiyle)")
	watch := flag.Bool("watch", false, "Canlı izleme modu (dosya değiştikçe PDF güncellenir)")
	landscape := flag.Bool("landscape", false, "Yatay sayfa yönlendirmesi")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Kullanım: md2pdf [seçenekler]\n\nSeçenekler:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

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
