# GOMAD 🚀

> **Go** + **Nomad** = Her platformda evinde

Angular-first masaüstü uygulama framework'ü.

## Vizyon
```
Go'nun gücü + Angular'ın zenginliği = GOMAD
```

İki Google teknolojisini masaüstünde birleştiriyoruz.

## Durum

🚧 **Aktif Geliştirme - Faz 1**

| Faz | Açıklama | Durum |
|-----|----------|-------|
| 0 | Teorik Temeller | ✅ Tamamlandı |
| 1 | Platform Soyutlaması | 🔄 Devam Ediyor |
| 2 | WebView Entegrasyonu | ⏳ Sırada |
| 3 | Angular Bridge | ⏳ |
| 4-7 | ... | ⏳ |

## Hızlı Başlangıç
```bash
# Klonla
git clone https://github.com/AhmetCanSolak/gomad.git
cd gomad

# Demo'yu çalıştır (Windows)
go build ./cmd/demo
./demo.exe
```

## Proje Yapısı
```
gomad/
├── cmd/
│   └── demo/
│       └── main.go           # Test uygulaması
├── internal/
│   └── platform/
│       ├── platform.go       # Window interface
│       ├── types.go          # Ortak tipler (MouseButton)
│       └── windows/
│           ├── types.go      # Win32 API tipleri
│           ├── proc.go       # DLL prosedürleri
│           └── window.go     # Windows implementasyonu
├── VISION.md                 # Proje vizyonu
├── ROADMAP.md               # Öğrenme yol haritası
├── README.md                # Bu dosya
└── go.mod
```

## Örnek Kullanım
```go
package main

import (
    "fmt"
    "github.com/AhmetCanSolak/gomad/internal/platform"
    "github.com/AhmetCanSolak/gomad/internal/platform/windows"
)

func main() {
    window := windows.NewWindow()
    
    window.SetTitle("Merhaba GOMAD!")
    window.SetSize(800, 600)
    
    window.OnClick(func(x, y int, button platform.MouseButton) {
        fmt.Printf("Tıklama: (%d, %d)\n", x, y)
    })
    
    window.OnClose(func() {
        fmt.Println("Hoşçakal!")
    })
    
    window.Run()
}
```

## Platform Desteği

| Platform | Durum | Notlar |
|----------|-------|--------|
| Windows | ✅ Çalışıyor | Win32 API |
| macOS | ⏳ Planlandı | Cocoa |
| Linux | ⏳ Planlandı | X11/Wayland |

## Bilinen Sorunlar

- [ ] Mouse move + SetTitle kombinasyonu donmaya yol açabiliyor (throttle gerekli)

## Yol Haritası

Detaylı yol haritası için: [ROADMAP.md](./ROADMAP.md)

## Lisans

MIT

---

*"Her büyük framework, birinin 'Ben bunu daha iyi yapabilirim' demesiyle başladı."*