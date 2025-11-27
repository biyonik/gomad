# GOMAD 🚀

> **Go** + **Nomad** = Her platformda evinde

Angular-first masaüstü uygulama framework'ü. İki Google teknolojisini masaüstünde birleştiriyoruz.

```
Go'nun gücü + Angular'ın zenginliği = GOMAD
```

## ✨ Ne Yapabilirsin?

```go
// Go'da fonksiyon yaz
app.Bind("readConfig", func(path string) (Config, error) {
    return loadConfigFromFile(path)
})
```

```javascript
// JavaScript'ten çağır
const config = await window.gomad.call("readConfig", "./config.json");
```

**Bu kadar basit.** Go'nun tüm gücü (dosya sistemi, veritabanı, network) + Web'in tüm güzelliği (HTML, CSS, animasyonlar).

---

## 📊 Durum

| Faz | Açıklama | Durum |
|-----|----------|-------|
| 0 | Teorik Temeller | ✅ Tamamlandı |
| 1 | Platform Soyutlaması | ✅ Tamamlandı |
| 2 | WebView + Bridge | ✅ Tamamlandı |
| 3 | Angular Entegrasyonu | ⏳ Sırada |
| 4 | Framework API | ⏳ |
| 5 | Production Ready | ⏳ |

---

## 🚀 Hızlı Başlangıç

### Gereksinimler

- Go 1.21+
- CGO enabled (Windows için MinGW-w64 GCC)
- WebView2 Runtime (Windows - genellikle zaten yüklü)

### Çalıştır

```powershell
# Klonla
git clone https://github.com/biyonik/gomad.git
cd gomad

# Windows için CGO ayarla
$env:CGO_ENABLED = "1"

# Örneği çalıştır
go run ./cmd/examples/hello-world
```

### Ne Göreceksin?

Muhteşem bir arayüz! Ve her buton gerçekten Go fonksiyonlarını çağırıyor:

- **getVersion()** → Argümansız fonksiyon
- **greet(name)** → String argüman, string dönüş
- **add(a, b)** → Matematiksel işlem
- **getUser(id)** → Kompleks obje dönüşü
- **divide(a, b)** → Hata yönetimi (0'a bölmeyi dene!)
- **longTask(seconds)** → Async işlem

---

## 🏗️ Mimari

```
┌─────────────────────────────────────────────────────────────┐
│                        GOMAD                                │
├─────────────────────────────────────────────────────────────┤
│  pkg/gomad              → Public API (kullanıcılar bunu görür)
├─────────────────────────────────────────────────────────────┤
│  internal/bridge        → Go ↔ JavaScript köprüsü           
├─────────────────────────────────────────────────────────────┤
│  internal/webview       → WebView soyutlaması               
├─────────────────────────────────────────────────────────────┤
│  internal/platform      → OS-specific (Win32, Cocoa, X11)   
└─────────────────────────────────────────────────────────────┘
```

---

## 📁 Proje Yapısı

```
gomad/
├── cmd/examples/
│   └── hello-world/main.go    # 🧪 Bridge demo
├── pkg/gomad/
│   ├── app.go                 # 🌟 Application struct
│   └── options.go             # ⚙️ WithTitle, WithSize, ...
├── internal/
│   ├── bridge/
│   │   ├── message.go         # 📨 JSON mesaj protokolü
│   │   ├── registry.go        # 📝 Fonksiyon kaydı
│   │   └── bridge.go          # 🌉 İletişim koordinatörü
│   ├── webview/
│   │   └── webview.go         # 🌐 WebView wrapper
│   ├── platform/
│   │   ├── platform.go        # 🖼️ Window interface
│   │   └── windows/           # 🪟 Win32 API
│   └── errors/
│       └── errors.go          # ❌ Hata tipleri
├── docs/
│   └── architecture.md        # 📚 Detaylı mimari
├── VISION.md                  # 🎯 Proje vizyonu
├── ROADMAP.md                 # 🗺️ Öğrenme yol haritası
├── Makefile                   # 🔧 Build komutları
└── go.mod
```

---

## 📡 API Kullanımı

### Go Tarafı

```go
package main

import (
    "github.com/biyonik/gomad/internal/webview"
)

func main() {
    wv, _ := webview.New(webview.Options{
        Title:  "My App",
        Width:  800,
        Height: 600,
        Debug:  true,
    })

    // Fonksiyon bind et
    wv.Bridge().Bind("greet", func(name string) string {
        return "Merhaba, " + name + "!"
    })

    // Hata dönebilen fonksiyon
    wv.Bridge().Bind("divide", func(a, b float64) (float64, error) {
        if b == 0 {
            return 0, fmt.Errorf("sıfıra bölme hatası")
        }
        return a / b, nil
    })

    // Event gönder
    wv.Emit("app:ready", map[string]any{"version": "1.0"})

    wv.Run()
}
```

### JavaScript Tarafı

```javascript
// Fonksiyon çağır
const greeting = await window.gomad.call("greet", "Ahmet");
// → "Merhaba, Ahmet!"

// Hata yakala
try {
    await window.gomad.call("divide", 10, 0);
} catch (e) {
    console.error(e.message); // "sıfıra bölme hatası"
}

// Event dinle
window.gomad.on("app:ready", (data) => {
    console.log("Versiyon:", data.version);
});
```

---

## 🎯 Desteklenen Fonksiyon İmzaları

| İmza | Açıklama |
|------|----------|
| `func()` | Argümansız, dönüşsüz |
| `func() error` | Sadece hata dönebilir |
| `func() T` | Tek değer döner |
| `func() (T, error)` | Değer + hata |
| `func(args...) (T, error)` | Argümanlı, değer + hata |

---

## 🖥️ Platform Desteği

| Platform | Durum | Backend |
|----------|-------|---------|
| Windows | ✅ Çalışıyor | WebView2 (Edge/Chromium) |
| macOS | ⏳ Planlandı | WKWebView |
| Linux | ⏳ Planlandı | WebKitGTK |

---

## 📖 Dokümantasyon

- [VISION.md](./VISION.md) - Proje vizyonu ve hedefler
- [ROADMAP.md](./ROADMAP.md) - Detaylı öğrenme yol haritası
- [docs/architecture.md](./docs/architecture.md) - Teknik mimari

---

## 🤝 Katkıda Bulunma

Bu bir öğrenme projesi ama katkılara açık!

1. Fork et
2. Feature branch oluştur (`git checkout -b feature/amazing`)
3. Commit et (`git commit -m 'feat: amazing feature'`)
4. Push et (`git push origin feature/amazing`)
5. Pull Request aç

---

## 📄 Lisans

MIT License

---

## 👤 Geliştirici

**Ahmet ALTUN** - [@biyonik](https://github.com/biyonik)

---

<p align="center">
  <i>"Her büyük framework, birinin 'Ben bunu daha iyi yapabilirim' demesiyle başladı."</i>
</p>

<p align="center">
  <b>Go + Angular = ❤️</b>
</p>