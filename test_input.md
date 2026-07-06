# Markdown to PDF Test Belgesi

Bu, Go diliyle yazılmış saf Go PDF motorumuzun dönüştürme yeteneklerini test etmek için hazırlanmış bir belgedir.
Uygulama, **hiçbir harici sistem bağımlılığı** (Chrome, wkhtmltopdf vb.) gerektirmeden, doğrudan GoPDF kütüphanesi ve unicode uyumlu Inter fontu ile çalışmaktadır.

## 1. Yazı Stilleri ve Biçimlendirmeler

Markdown formatındaki farklı inline stilleri mükemmel şekilde destekliyoruz:
- **Kalın Yazı** (`**bold**`): Bu metin kalın olarak render edilmelidir.
- *Eğik Yazı* (`*italic*`): Bu metin italik olarak render edilmelidir.
- **Bold ve *Italic* Bir Arada**: Bu cümle içinde **hem kalın hem de *italik* olan** kısımlar yan yana düzgün şekilde çizilmelidir.
- Satır içi Kod (`code span`): `fmt.Println("Merhaba Dünya!")` kodu paragraf içinde monospace yazı tipi ve hafif gri arka planla görünmelidir.
- Bağlantılar (Links): [Antigravity GitHub](https://github.com) bağlantısı mavi renkte, altı çizili ve tıklanabilir olmalıdır.

---

## 2. Listeler

Hem sıralı hem de sırasız listeler ve bunların iç içe geçmiş halleri düzgün hizalanmalıdır:

### Sırasız Liste
- Birinci öğe
- İkinci öğe
  - Alt liste öğesi 1
  - Alt liste öğesi 2
- Üçüncü öğe

### Sıralı Liste
1. İlk adım: Markdown dosyasını hazırlayın.
2. İkinci adım: `md2pdf -input test_input.md` komutunu çalıştırın.
3. Üçüncü adım: Oluşan PDF belgesini açıp keyifle inceleyin.

---

## 3. Kod Blokları (Fenced Code Blocks)

Kod blokları ayrı bir kutu içinde, monospace yazı tipiyle ve uzun satırları otomatik sararak (wrap) render edilir:

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	message := "md2pdf aracı Go diliyle yazıldı ve milisaniyeler içinde çalışıyor!"
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), message)
}
```

---

## 4. Alıntılar (Blockquotes)

> "Hız en önemli önceliğimizdir. Ancak hız kadar, üretilen çıktının kalitesi ve sistemin taşınabilirliği de hayati önem taşır."
>
> — md2pdf Tasarım Ekibi
