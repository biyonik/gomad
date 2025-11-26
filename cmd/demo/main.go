package main

import (
	"fmt"

	"github.com/biyonik/gomad/internal/platform"
	"github.com/biyonik/gomad/internal/platform/windows"
)

func main() {
	window := windows.NewWindow()

	window.SetTitle("GOMAD - Klavye ve Konum Testi")
	window.SetSize(800, 600)

	// Klavye event'leri
	window.OnKeyDown(func(keyCode int) {
		fmt.Printf("Tuş basıldı: %d\n", keyCode)

		// ESC = 27, pencereyi kapat
		if keyCode == 27 {
			fmt.Println("ESC basıldı, kapatılıyor...")
			window.Close()
		}

		// C = 67, pencereyi ortala
		if keyCode == 67 {
			fmt.Println("Pencere ortalanıyor...")
			window.Center()
		}
	})

	window.OnClick(func(x, y int, button platform.MouseButton) {
		fmt.Printf("Tıklama: (%d, %d)\n", x, y)
	})

	window.OnClose(func() {
		fmt.Println("Hoşçakal! 👋")
	})

	fmt.Println("GOMAD başlıyor...")
	fmt.Println("C = Ortala, ESC = Kapat")
	window.Run()
	fmt.Println("GOMAD kapandı.")
}
