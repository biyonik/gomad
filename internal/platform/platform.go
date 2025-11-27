// Package platform
//
// Bu paket, GOMAD mimarisinin işletim sistemi bağımsız çalışabilmesini sağlayan
// temel soyutlama katmanıdır. Amaç, Windows, macOS ve Linux gibi birbirinden
// tamamen farklı API mekanizmalarına sahip sistemler üzerinde **ortak bir pencere,
// input ve event modelini** sağlayabilmektir.
//
// Geliştirici doğrudan bu paketi kullanmaz — tüm erişim `pkg/gomad` üstünden
// yapılır. Böylece uygulama, altında hangi OS API’si olduğunu bilmeden pencere
// oluşturabilir, event yönetebilir, render katmanını bağlayabilir.
//
// Her platform kendi alt klasöründe `platform.Window` arayüzünü native API
// çağrılarıyla implemente eder:
//
//   - **Windows → Win32**   (CreateWindowEx, WndProc, HWND ...)
//   - **macOS → Cocoa/AppKit**   (NSWindow, Objective-C bridge)
//   - **Linux → X11 / Wayland**
//
// Bu soyutlama sayesinde kodun %90’ı işletim sistemi fark etmeksizin çalışır.
package platform

// ============================================================================
// WINDOW INTERFACE
// Üst seviye tüm pencere işlemlerinin ortak sözleşmesidir. Bir OS implementasyonu
// bu interface’i eksiksiz karşıladığı anda GOMAD tarafından pencere oluşturulabilir,
// taşınabilir ve etkileşimli UI sistemleri kurulabilir.
//
// Bu yapı yalnızca API kontratıdır — gerçek davranış platform modüllerinde oluşur.
// ============================================================================
type Window interface {

	// ==================== Lifecycle ====================

	// Show
	// ----------------------------------------------------
	// Pencereyi ekranda görünür hâle getirir. Create aşaması ile yalnızca
	// hafızada oluşturulur; Show çağrılmadığı sürece kullanıcı pencereyi
	// görmez. Tipik kullanım:
	//
	//     win := platform.New()
	//     win.SetTitle("Hello")
	//     win.Show()
	//
	Show()

	// Hide
	// ----------------------------------------------------
	// Pencereyi yok etmeden gizler. UI geçici olarak kaldırılmak isteniyor
	// ancak destroy edilip yeniden oluşturulması istenmiyorsa kullanılır.
	Hide()

	// Close
	// ----------------------------------------------------
	// Pencereyi tamamen yok eder, tüm handle ve resource'ları serbest bırakır.
	// Bir pencere kapatıldıktan sonra tekrar kullanılamaz.
	Close()

	// ==================== Properties ====================

	// SetTitle → Başlık metni ayarlama.
	SetTitle(title string)

	// GetTitle → Geçerli başlık metnini döndürür.
	GetTitle() string

	// SetSize → Pencere iç alan (client area) genişlik & yüksekliği piksel olarak.
	SetSize(width, height int)

	// GetSize → Pencere boyutlarını geri döner.
	GetSize() (width, height int)

	// SetPosition → Pencereyi ekranda ilgili koordinata taşır.
	SetPosition(x, y int)

	// GetPosition → Mevcut x,y koordinatlarını döner.
	GetPosition() (x, y int)

	// Center → Pencereyi ekran ortasına hizalar.
	Center()

	// ==================== State ====================

	// SetResizable → Kullanıcının pencereyi yeniden boyutlandırıp
	// büyütüp küçültemeyeceğini kontrol eder.
	SetResizable(resizable bool)

	// IsResizable → Pencere yeniden boyutlandırılabilir mi?
	IsResizable() bool

	// Minimize → Görev çubuğuna/dock’a küçültür.
	Minimize()

	// Maximize → Pencereyi ekranı dolduracak şekilde büyütür.
	Maximize()

	// Restore → Minimize/Maximize sonrası normal hâle getirir.
	Restore()

	// ==================== Events ====================

	// OnClose
	// -------------------------------------------------------------------------
	// Pencere kapatılmadan önce tetiklenir. Geri dönen **bool** önemlidir:
	// `false` dönerse pencere kapanmaz (ör. "Kaydedilmemiş değişiklikler var"
	// diyerek kullanıcıyı durdurmak).
	OnClose(callback func() bool)

	// OnResize → Boyut değiştiğinde tetiklenir.
	OnResize(callback func(width, height int))

	// OnMove → Konum değiştiğinde tetiklenir.
	OnMove(callback func(x, y int))

	// OnFocus → Odak kazanıldığında.
	OnFocus(callback func())

	// OnBlur → Odak kaybedildiğinde.
	OnBlur(callback func())

	// ==================== Native ====================

	// Handle
	// -------------------------------------------------------------------------
	// Native OS pencere tanımlayıcısını döner.
	//
	//   Windows → HWND
	//   macOS   → NSWindow*
	//   Linux   → X11 Window (veya daha sonrası Wayland surface)
	//
	// WebView gömme, OpenGL bağlantısı gibi sistem-altı kullanım için gereklidir.
	Handle() uintptr
}

// ============================================================================
// MOUSE BUTTON ENUM
// Fare butonlarını soyut bir enum olarak taşır. UI event sisteminde kullanılır.
// ============================================================================
type MouseButton int

const (
	MouseButtonLeft MouseButton = iota
	MouseButtonRight
	MouseButtonMiddle
)

// String → Buton adını okunabilir formatta döndürür.
func (b MouseButton) String() string {
	switch b {
	case MouseButtonLeft:
		return "Left"
	case MouseButtonRight:
		return "Right"
	case MouseButtonMiddle:
		return "Middle"
	default:
		return "Unknown"
	}
}

// ============================================================================
// KEY MODIFIER FLAGS
// Klavye üzerindeki mod tuşları (Shift, Ctrl, Alt, Super) bitmask ile tutulur.
// ============================================================================
type KeyModifiers uint8

const (
	ModShift KeyModifiers = 1 << iota
	ModCtrl
	ModAlt
	ModSuper // Windows Key / Command Key
)

func (m KeyModifiers) HasShift() bool { return m&ModShift != 0 }
func (m KeyModifiers) HasCtrl() bool  { return m&ModCtrl != 0 }
func (m KeyModifiers) HasAlt() bool   { return m&ModAlt != 0 }
func (m KeyModifiers) HasSuper() bool { return m&ModSuper != 0 }

// ============================================================================
// WINDOW CONFIG
// Pencere oluşturma parametrelerini tek paket hâlinde taşıyan yapı.
// ============================================================================
type WindowConfig struct {
	Title     string // Başlık
	Width     int    // Genişlik
	Height    int    // Yükseklik
	Resizable bool   // Boyutlandırılabilir mi?
	Centered  bool   // Ortalansın mı?
}

// DefaultWindowConfig
// -----------------------------------------------------------------------------
// Yeni bir pencere oluşturmak için kullanılabilecek “ideal başlangıç değerleri”
// döner. Geliştirici çoğu durumda ekstra bir ayar yapmadan hızlıca çalışabilir.
//
//	cfg := platform.DefaultWindowConfig()
//	win := platform.NewWindow(cfg)
func DefaultWindowConfig() WindowConfig {
	return WindowConfig{
		Title:     "GOMAD Application",
		Width:     800,
		Height:    600,
		Resizable: true,
		Centered:  true,
	}
}
