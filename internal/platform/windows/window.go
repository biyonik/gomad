//go:build windows
// +build windows

package windows

/*
===========================================================================================================
🪟 Windows Native Window — Olay Döngüsü, Mesaj İşleme, Etkileşim Sistemi (Win32 API Tabanlı)
===========================================================================================================

Bu dosya, Windows işletim sistemine özgü gerçek bir grafik pencere oluşturmayı,
çalıştırmayı, mesaj döngüsünü yönetmeyi ve kullanıcı etkileşimlerini yakalamayı sağlayan
tam teşekküllü `Window` implementasyonunu içerir.

Buradaki yapı yalnızca UI oluşturmak için değil; Win32 API’nin en alt seviyesinde
pencere yaşam döngüsünü kontrol etmek için tasarlanmıştır. Bu nedenle:

📌 `CreateWindowEx` ile *gerçek native pencere* oluşturulur
📌 `GetMessage/DispatchMessage` ile WinAPI event loop aktif tutulur
📌 `wndProc` ile mouse, kapatma, destroy gibi **ham mesajlar yakalanır**
📌 Üst seviye projeler soyutlama katmanında platform bağımsız kullanabilir

Bu sınıfın amacı; modern Go kodunun, WinAPI’nin karmaşık mesaj sistemine doğrudan
dokunmadan pencere oluşturabilmesini sağlamaktır. Kodun içinde:
- Mutex ile thread-safety korunur
- Callback fonksiyonları ile kullanıcı etkileşimi üst seviyeye taşınır
- WM_XXXX mesajları manuel işlenerek gerçek zamanlı input elde edilir
- High-level platform arayüzü ile low-level Win32 API kusursuz biçimde birleşir

Bu sınıfa “Görsel UI’nın kalbi” demek abartı değildir — çünkü sistem her input, her
hareket, her tıklama, her kapanma talimatını burada duyup işler.
Event geçmezse pencere hareket etmez, mesaj okunmazsa yazılım donar.
Burası pencerenin solunum borusu gibidir; kesilirse tüm UI ölür.

----------------------------------------------------------------------------------------
@author   Ahmet ALTUN
@github   github.com/biyonik
@linkedin linkedin.com/in/biyonik
@email    ahmet.altun60@gmail.com
----------------------------------------------------------------------------------------
*/

import (
	"sync"
	"syscall"
	"unsafe"

	"github.com/biyonik/gomad/internal/platform"
)

// Window, native Win32 penceresini temsil eden yapıdır.
// hwnd → gerçek pencere handle'ı
// title,width,height → pencerenin temel özellikleri
// onClose,onMouseMove,onClick → harici callback bağlantıları (event binding)
// mu → veri bütünlüğü için kilit mekanizması (thread-safe çalışma)
type Window struct {
	hwnd   HWND
	title  string
	width  int
	height int

	onClose     func()
	onMouseMove func(x, y int)
	onClick     func(x, y int, button platform.MouseButton)
	onKeyDown   func(keyCode int)
	onKeyUp     func(keyCode int)

	mu sync.Mutex
}

// activeWindow, Windows mesaj işleyicisinin hangi pencereye bağlı olduğunu saklar.
// WinAPI tek global wndProc çalıştırır → aktif pencere buradan yönlendirilir.
var activeWindow *Window

// NewWindow, default değerlerle yeni bir native pencere örneği oluşturur.
// Başlık verilir, genişlik-yükseklik atanır, ancak henüz OS tarafında oluşmaz.
func NewWindow() *Window {
	return &Window{
		title:  "GOMAD Window",
		width:  800,
		height: 600,
	}
}

// SetTitle, pencerenin başlığını değiştirir.
// Win32 handle oluşmuşsa anında OS tarafına yansır.
func (w *Window) SetTitle(title string) {
	w.mu.Lock()
	w.title = title
	w.mu.Unlock()

	if w.hwnd != 0 {
		_, _, err := procSetWindowText.Call(
			uintptr(w.hwnd),
			uintptr(unsafe.Pointer(StringToUTF16Ptr(title))),
		)
		if err != nil {
			return
		}
	}
}

// SetSize, pencerenin genişlik-yüksekliğini günceller.
// Pencere oluşturulmuşsa SetWindowPos ile Windows API'ye yansıtılır.
func (w *Window) SetSize(width, height int) {
	w.mu.Lock()
	w.width = width
	w.height = height
	w.mu.Unlock()

	if w.hwnd != 0 {
		const SWP_NOMOVE = 0x0002
		const SWP_NOZORDER = 0x0004
		procSetWindowPos.Call(
			uintptr(w.hwnd),
			0,
			0, 0,
			uintptr(width), uintptr(height),
			SWP_NOMOVE|SWP_NOZORDER,
		)
	}
}

// OnClose, pencere kapanmadan önce tetiklenecek fonksiyonu kayıt eder.
func (w *Window) OnClose(callback func()) {
	w.mu.Lock()
	w.onClose = callback
	w.mu.Unlock()
}

// OnMouseMove, fare hareketi olduğunda çağrılacak callback'i kayıt eder.
func (w *Window) OnMouseMove(callback func(x, y int)) {
	w.mu.Lock()
	w.onMouseMove = callback
	w.mu.Unlock()
}

// OnClick, mouse tıklaması algılandığında tetiklenecek fonksiyonu kayıt eder.
func (w *Window) OnClick(callback func(x, y int, button platform.MouseButton)) {
	w.mu.Lock()
	w.onClick = callback
	w.mu.Unlock()
}

// OnKeyDown, klavyede bir tuşa basıldığında tetiklenecek fonksiyonu kayıt eder.
func (w *Window) OnKeyDown(callback func(keyCode int)) {
	w.mu.Lock()
	w.onKeyDown = callback
	w.mu.Unlock()
}

// OnKeyUp, klavyede basılı tuş bırakıldığında tetiklenecek fonksiyonu kayıt eder.
func (w *Window) OnKeyUp(callback func(keyCode int)) {
	w.mu.Lock()
	w.onKeyUp = callback
	w.mu.Unlock()
}

// Show, oluşturulmuş pencereyi ekranda görünür hale getirir.
func (w *Window) Show() {
	if w.hwnd != 0 {
		procShowWindow.Call(uintptr(w.hwnd), SW_SHOW)
	}
}

// Close, pencereyi kapatır ve DestroyWindow tetikler.
func (w *Window) Close() {
	if w.hwnd != 0 {
		procDestroyWindow.Call(uintptr(w.hwnd))
	}
}

// Run, pencereyi oluşturur ve sonsuz mesaj döngüsünü başlatır.
// Uygulama bu fonksiyonda yaşar, kullanıcı kapatınca sona erer.
func (w *Window) Run() {
	activeWindow = w

	className := StringToUTF16Ptr("GOMAD_WINDOW_CLASS")

	hInstance, _, _ := procGetModuleHandle.Call(0)
	cursor, _, _ := procLoadCursor.Call(0, IDC_ARROW)

	wndClass := WNDCLASSEX{
		Style:         CS_HREDRAW | CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     HINSTANCE(hInstance),
		HCursor:       HCURSOR(cursor),
		HbrBackground: HBRUSH(COLOR_WINDOW + 1),
		LpszClassName: className,
	}
	wndClass.CbSize = wndClass.Size()

	procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wndClass)))

	// 2. Pencere oluşturma
	hwnd, _, _ := programCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(StringToUTF16Ptr(w.title))),
		WS_OVERLAPPEDWINDOW,
		CW_USEDEFAULT,
		CW_USEDEFAULT,
		uintptr(w.width),
		uintptr(w.height),
		0,
		0,
		hInstance,
		0,
	)

	w.hwnd = HWND(hwnd)

	w.Show()

	var msg MSG
	for {
		ret, _, _ := procGetMessage.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
		)

		if ret == 0 {
			break
		}

		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

// SetPosition, pencerenin boyutunu değiştirmeden ekrandaki konumunu belirtilen x ve y koordinatlarına ayarlar.
func (w *Window) SetPosition(x, y int) {
	if w.hwnd != 0 {
		const SWP_NOSIZE = 0x0001
		const SWP_NOZORDER = 0x0004
		procSetWindowPos.Call(
			uintptr(w.hwnd),
			0,
			uintptr(x), uintptr(y),
			0, 0,
			SWP_NOSIZE|SWP_NOZORDER,
		)
	}
}

// GetPosition pencerenin ekran koordinatlarındaki geçerli konumunu (x, y) olarak döndürür.
func (w *Window) GetPosition() (x, y int) {
	if w.hwnd != 0 {
		var rect RECT
		procGetWindowRect.Call(
			uintptr(w.hwnd),
			uintptr(unsafe.Pointer(&rect)),
		)
		return int(rect.Left), int(rect.Top)
	}
	return 0, 0
}

// Center, pencereyi ekranın ortasına taşır.
func (w *Window) Center() {
	if w.hwnd == 0 {
		return
	}

	screenWidth, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	screenHeight, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)

	var rect RECT
	procGetWindowRect.Call(
		uintptr(w.hwnd),
		uintptr(unsafe.Pointer(&rect)),
	)
	windowWidth := int(rect.Right - rect.Left)
	windowHeight := int(rect.Bottom - rect.Top)

	x := (int(screenWidth) - windowWidth) / 2
	y := (int(screenHeight) - windowHeight) / 2

	w.SetPosition(x, y)
}

// wndProc, WinAPI mesajlarının işlendiği kalp fonksiyondur.
// Mouse, close, destroy gibi tüm event’ler buradan geçer.
func wndProc(hwnd HWND, msg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	w := activeWindow
	if w == nil {
		ret, _, _ := procDefWindowProc.Call(
			uintptr(hwnd), uintptr(msg), uintptr(wParam), uintptr(lParam),
		)
		return LRESULT(ret)
	}

	switch msg {
	case WM_CLOSE:
		if w.onClose != nil {
			w.onClose()
		}
		procDestroyWindow.Call(uintptr(hwnd))
		return 0

	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0

	case WM_MOUSEMOVE:
		if w.onMouseMove != nil {
			x := GET_X_LPARAM(lParam)
			y := GET_Y_LPARAM(lParam)
			w.onMouseMove(x, y)
		}
		return 0

	case WM_LBUTTONDOWN:
		if w.onClick != nil {
			x := GET_X_LPARAM(lParam)
			y := GET_Y_LPARAM(lParam)
			w.onClick(x, y, platform.MouseButtonLeft)
		}
		return 0

	case WM_RBUTTONDOWN:
		if w.onClick != nil {
			x := GET_X_LPARAM(lParam)
			y := GET_Y_LPARAM(lParam)
			w.onClick(x, y, platform.MouseButtonRight)
		}
		return 0

	case WM_MBUTTONDOWN:
		if w.onClick != nil {
			x := GET_X_LPARAM(lParam)
			y := GET_Y_LPARAM(lParam)
			w.onClick(x, y, platform.MouseButtonMiddle)
		}
		return 0

	case WM_KEYDOWN:
		if w.onKeyDown != nil {
			keyCode := int(wParam)
			w.onKeyDown(keyCode)
		}
		return 0

	case WM_KEYUP:
		if w.onKeyUp != nil {
			keyCode := int(wParam)
			w.onKeyUp(keyCode)
		}
		return 0
	}

	ret, _, _ := procDefWindowProc.Call(
		uintptr(hwnd), uintptr(msg), uintptr(wParam), uintptr(lParam),
	)
	return LRESULT(ret)
}
