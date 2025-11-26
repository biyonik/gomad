package main

import (
	"fmt"
	"time"

	"github.com/biyonik/gomad/internal/platform"
	"github.com/biyonik/gomad/internal/platform/windows"
)

func main() {
	window := windows.NewWindow()

	window.SetTitle("GOMAD - Mouse'u hareket ettir!")
	window.SetSize(800, 600)

	// Throttle için son güncelleme zamanı
	var lastUpdate time.Time
	throttle := 50 * time.Millisecond // 50ms = saniyede max 20 güncelleme

	// Mouse hareket edince başlığı güncelle (throttled)
	window.OnMouseMove(func(x, y int) {
		now := time.Now()
		if now.Sub(lastUpdate) < throttle {
			return // Çok erken, atla
		}
		lastUpdate = now

		title := fmt.Sprintf("GOMAD - Mouse: (%d, %d)", x, y)
		window.SetTitle(title)
	})

	// Click'te hemen güncelle (throttle yok, click nadir)
	window.OnClick(func(x, y int, button platform.MouseButton) {
		buttonName := "?"
		switch button {
		case platform.MouseButtonLeft:
			buttonName = "SOL TIKLAMA"
		case platform.MouseButtonRight:
			buttonName = "SAĞ TIKLAMA"
		case platform.MouseButtonMiddle:
			buttonName = "ORTA TIKLAMA"
		}
		title := fmt.Sprintf("GOMAD - %s: (%d, %d)", buttonName, x, y)
		window.SetTitle(title)
	})

	window.OnClose(func() {
		fmt.Println("Pencere kapanıyor... Hoşçakal! 👋")
	})

	fmt.Println("GOMAD başlıyor...")
	fmt.Println("Mouse'u pencere içinde hareket ettir ve başlığa bak!")
	window.Run()
	fmt.Println("GOMAD kapandı.")
}
