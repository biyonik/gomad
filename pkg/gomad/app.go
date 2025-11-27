// Package gomad, Go ve Angular kullanarak masaüstü uygulamaları geliştirmek için bir çerçeve sunar.
// GOMAD (Go + Nomad), Go'nun gücünü Angular'ın zengin UI yetenekleri ile birleştirerek
// çapraz platform masaüstü uygulamaları oluşturmayı mümkün kılar.
//
// Bu paket, uygulamanın temel yönetimini sağlar: pencere oluşturma, WebView yönetimi
// ve Go-JavaScript köprüsü.
//
// @author Ahmet ALTUN
// @github github.com/biyonik
// @linkedin linkedin.com/in/biyonik
// @email ahmet.altun60@gmail.com
package gomad

import (
	"fmt"
	"runtime"

	"github.com/biyonik/gomad/internal/webview"
)

// Application, GOMAD masaüstü uygulamasını temsil eder.
// Pencereyi, WebView'i ve Go-JavaScript köprüsünü yönetir.
//
// Yeni bir Application oluşturmak için New() kullanılır:
//
//	app := gomad.New(
//	    gomad.WithTitle("My App"),
//	    gomad.WithSize(800, 600),
//	)
//
// Application, aynı anda birden fazla goroutine'den güvenli değildir.
// Tüm metodlar ana goroutine'den çağrılmalıdır.
type Application struct {
	config  *config
	webview *webview.WebViewImpl

	// Durum
	running bool
}

// New, verilen seçeneklerle yeni bir Application oluşturur.
// Eğer seçenek verilmezse mantıklı varsayılanlar kullanılır.
//
// Seçenekler sırasıyla uygulanır; sonradan verilen seçenekler önceki değerleri geçersiz kılar.
//
// Örnek:
//
//	app := gomad.New()
//	app := gomad.New(
//	    gomad.WithTitle("Custom Title"),
//	    gomad.WithSize(1024, 768),
//	    gomad.WithDebug(true),
//	)
func New(opts ...Option) *Application {
	// Varsayılanlar
	cfg := defaultConfig()

	// Seçenekleri uygula
	for _, opt := range opts {
		opt(cfg)
	}

	return &Application{
		config: cfg,
	}
}

// Run, uygulamayı başlatır ve pencere kapanana kadar bloklar.
// Ana goroutine'den çağrılmalıdır.
//
// Başarısız olursa hata döner.
func (a *Application) Run() error {
	// GUI işlemleri ana thread'de olmalı (özellikle macOS için)
	runtime.LockOSThread()

	// WebView oluştur
	wv, err := webview.New(webview.Options{
		Title:  a.config.title,
		Width:  a.config.width,
		Height: a.config.height,
		Debug:  a.config.debug,
		URL:    a.config.url,
		HTML:   a.config.html,
	})
	if err != nil {
		return fmt.Errorf("failed to create webview: %w", err)
	}

	a.webview = wv
	a.running = true

	// OnReady callback
	if a.config.onReady != nil {
		a.config.onReady()
	}

	// Olay döngüsünü başlat (blocking)
	wv.Run()

	// Temizlik
	wv.Destroy()
	a.running = false

	return nil
}

// Bind, JavaScript tarafında çağrılabilecek bir Go fonksiyonu kaydeder.
//
// Fonksiyonun imzalarından biri olmalıdır:
//   - func()
//   - func() error
//   - func() T
//   - func() (T, error)
//   - func(args...) (T, error)
//
// T, JSON-serializable bir tip olmalıdır.
//
// Örnek:
//	app.Bind("getVersion", func() string { return "1.0.0" })
//	app.Bind("add", func(a
