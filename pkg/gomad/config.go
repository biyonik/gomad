// Package gomad, Go ve Angular kullanılarak masaüstü uygulamaları geliştirmek için fonksiyonel seçenekler sunar.
// Bu paket, Application yapılandırmasını temiz ve genişletilebilir bir şekilde yönetmek için Option pattern kullanır.
//
// Her Option, uygulamanın yapılandırmasını değiştirir ve New() çağrısında uygulanır.
//
// Örnek kullanım:
//
//	app := gomad.New(
//	    gomad.WithTitle("My App"),
//	    gomad.WithSize(1024, 768),
//	    gomad.WithDebug(true),
//	)
//
// @author Ahmet ALTUN
// @github github.com/biyonik
// @linkedin linkedin.com/in/biyonik
// @email ahmet.altun60@gmail.com
package gomad

// Option, Application yapılandırmasını değiştiren fonksiyonel bir seçenektir.
// Fonksiyonel seçenekler deseni, API'nin genişletilebilir ve okunabilir olmasını sağlar.
type Option func(*config)

// config, uygulama konfigürasyonunu tutar.
type config struct {
	// Pencere ayarları
	title     string
	width     int
	height    int
	resizable bool

	// WebView ayarları
	debug bool
	url   string
	html  string

	// Callbacks
	onReady func()
}

// defaultConfig, mantıklı varsayılan değerler döner.
func defaultConfig() *config {
	return &config{
		title:     "GOMAD Application",
		width:     800,
		height:    600,
		resizable: true,
		debug:     false,
	}
}

// WithTitle, pencere başlığını ayarlar.
//
// Örnek:
//
//	app := gomad.New(gomad.WithTitle("My Awesome App"))
func WithTitle(title string) Option {
	return func(c *config) {
		c.title = title
	}
}

// WithSize, pencerenin başlangıç boyutlarını piksel cinsinden ayarlar.
//
// Örnek:
//
//	app := gomad.New(gomad.WithSize(1280, 720))
func WithSize(width, height int) Option {
	return func(c *config) {
		c.width = width
		c.height = height
	}
}

// WithResizable, pencerenin yeniden boyutlandırılabilir olup olmadığını ayarlar.
// Varsayılan: true
//
// Örnek:
//
//	app := gomad.New(gomad.WithResizable(false)) // Sabit boyutlu pencere
func WithResizable(resizable bool) Option {
	return func(c *config) {
		c.resizable = resizable
	}
}
