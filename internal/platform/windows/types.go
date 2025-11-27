// ============================================================================
// Windows API Sabitler, Yapılar ve Yardımcı Fonksiyonlar
//
// Bu dosya, GOMAD için Win32 API sabitlerini, yapıları ve yardımcı fonksiyonları
// içerir. Amacı; pencere oluşturma, mesaj döngüsü ve sistem ölçümlerini
// yönetmek için gerekli temel taşları sağlamaktır.
//
// @author   Ahmet ALTUN
// @github   github.com/biyonik
// @linkedin linkedin.com/in/biyonik
// @email    ahmet.altun60@gmail.com
// ============================================================================

//go:build windows

package windows

import (
	"syscall"
	"unsafe"
)

// ==================== Window Styles ====================

// Windows pencere stilleri
const (
	WS_OVERLAPPED   = 0x00000000 // Standart üst pencere
	WS_POPUP        = 0x80000000 // Popup pencere
	WS_CHILD        = 0x40000000 // Child window
	WS_MINIMIZE     = 0x20000000 // Minimize edilmiş pencere
	WS_VISIBLE      = 0x10000000 // Başlangıçta görünür
	WS_DISABLED     = 0x08000000 // Başlangıçta devre dışı
	WS_CLIPSIBLINGS = 0x04000000 // Çakışan child pencereleri kırpar
	WS_CLIPCHILDREN = 0x02000000 // Child pencereleri kırpar
	WS_MAXIMIZE     = 0x01000000 // Maksimize edilmiş pencere
	WS_CAPTION      = 0x00C00000 // Başlık + kenarlık
	WS_BORDER       = 0x00800000 // Kenarlık
	WS_DLGFRAME     = 0x00400000 // Dialog çerçevesi
	WS_VSCROLL      = 0x00200000 // Dikey scrollbar
	WS_HSCROLL      = 0x00100000 // Yatay scrollbar
	WS_SYSMENU      = 0x00080000 // Sistem menüsü
	WS_THICKFRAME   = 0x00040000 // Yeniden boyutlandırılabilir
	WS_GROUP        = 0x00020000 // Tab grubu başlat
	WS_TABSTOP      = 0x00010000 // Tab ile odaklanabilir
	WS_MINIMIZEBOX  = 0x00020000 // Minimize kutusu
	WS_MAXIMIZEBOX  = 0x00010000 // Maximize kutusu

	// Sık kullanılan kombinasyon
	WS_OVERLAPPEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU |
		WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX
)

// Extended window styles (WS_EX)
const (
	WS_EX_DLGMODALFRAME    = 0x00000001 // Dialog frame
	WS_EX_TOPMOST          = 0x00000008 // Her zaman üstte
	WS_EX_ACCEPTFILES      = 0x00000010 // Dosya bırakılabilir
	WS_EX_TRANSPARENT      = 0x00000020 // Transparent pencere
	WS_EX_APPWINDOW        = 0x00040000 // Görev çubuğunda görünür
	WS_EX_OVERLAPPEDWINDOW = 0x00000300 // Kombine overlapped window
)

// ==================== Window Messages ====================

const (
	WM_NULL              = 0x0000
	WM_CREATE            = 0x0001
	WM_DESTROY           = 0x0002
	WM_MOVE              = 0x0003
	WM_SIZE              = 0x0005
	WM_ACTIVATE          = 0x0006
	WM_SETFOCUS          = 0x0007
	WM_KILLFOCUS         = 0x0008
	WM_ENABLE            = 0x000A
	WM_SETTEXT           = 0x000C
	WM_GETTEXT           = 0x000D
	WM_GETTEXTLENGTH     = 0x000E
	WM_PAINT             = 0x000F
	WM_CLOSE             = 0x0010
	WM_QUIT              = 0x0012
	WM_ERASEBKGND        = 0x0014
	WM_SHOWWINDOW        = 0x0018
	WM_ACTIVATEAPP       = 0x001C
	WM_SETCURSOR         = 0x0020
	WM_MOUSEACTIVATE     = 0x0021
	WM_GETMINMAXINFO     = 0x0024
	WM_WINDOWPOSCHANGING = 0x0046
	WM_WINDOWPOSCHANGED  = 0x0047
	WM_NOTIFY            = 0x004E
	WM_NCCREATE          = 0x0081
	WM_NCDESTROY         = 0x0082
	WM_NCHITTEST         = 0x0084
	WM_NCPAINT           = 0x0085
	WM_NCACTIVATE        = 0x0086

	// Klavye mesajları
	WM_KEYDOWN    = 0x0100
	WM_KEYUP      = 0x0101
	WM_CHAR       = 0x0102
	WM_SYSKEYDOWN = 0x0104
	WM_SYSKEYUP   = 0x0105
	WM_SYSCHAR    = 0x0106

	// Mouse mesajları
	WM_MOUSEMOVE     = 0x0200
	WM_LBUTTONDOWN   = 0x0201
	WM_LBUTTONUP     = 0x0202
	WM_LBUTTONDBLCLK = 0x0203
	WM_RBUTTONDOWN   = 0x0204
	WM_RBUTTONUP     = 0x0205
	WM_RBUTTONDBLCLK = 0x0206
	WM_MBUTTONDOWN   = 0x0207
	WM_MBUTTONUP     = 0x0208
	WM_MBUTTONDBLCLK = 0x0209
	WM_MOUSEWHEEL    = 0x020A

	// Boyutlandırma
	WM_SIZING        = 0x0214
	WM_MOVING        = 0x0216
	WM_ENTERSIZEMOVE = 0x0231
	WM_EXITSIZEMOVE  = 0x0232
)

// ==================== Show Window Commands ====================

const (
	SW_HIDE            = 0
	SW_SHOWNORMAL      = 1
	SW_SHOWMINIMIZED   = 2
	SW_SHOWMAXIMIZED   = 3
	SW_SHOWNOACTIVATE  = 4
	SW_SHOW            = 5
	SW_MINIMIZE        = 6
	SW_MAXIMIZE        = 7
	SW_SHOWMINNOACTIVE = 8
	SW_SHOWNA          = 9
	SW_RESTORE         = 10
	SW_SHOWDEFAULT     = 11
)

// ==================== System Metrics ====================

const (
	SM_CXSCREEN = 0 // Ekran genişliği
	SM_CYSCREEN = 1 // Ekran yüksekliği
)

// ==================== Special Values ====================

const (
	CW_USEDEFAULT = ^0x7FFFFFFF // Başlangıç boyutu/pozisyonu için varsayılan
)

// ==================== Structures ====================

// WNDCLASSEX: Windows pencere sınıf bilgisi
type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     syscall.Handle
	HIcon         syscall.Handle
	HCursor       syscall.Handle
	HbrBackground syscall.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       syscall.Handle
}

// MSG: Thread mesaj kuyruğu mesaj bilgisi
type MSG struct {
	HWnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

// POINT: X/Y koordinatı
type POINT struct {
	X, Y int32
}

// RECT: Dikdörtgen (üst sol ve alt sağ)
type RECT struct {
	Left, Top, Right, Bottom int32
}

// Width: RECT genişliği
func (r *RECT) Width() int32 {
	return r.Right - r.Left
}

// Height: RECT yüksekliği
func (r *RECT) Height() int32 {
	return r.Bottom - r.Top
}

// ==================== Helper Functions ====================

// UTF16PtrFromString: Go string → UTF16 pointer
func UTF16PtrFromString(s string) *uint16 {
	ptr, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		return nil
	}
	return ptr
}

// UTF16ToString: UTF16 pointer → Go string
func UTF16ToString(p *uint16) string {
	if p == nil {
		return ""
	}
	ptr := unsafe.Pointer(p)
	length := 0
	for *(*uint16)(unsafe.Pointer(uintptr(ptr) + uintptr(length)*2)) != 0 {
		length++
	}
	slice := make([]uint16, length)
	for i := 0; i < length; i++ {
		slice[i] = *(*uint16)(unsafe.Pointer(uintptr(ptr) + uintptr(i)*2))
	}
	return syscall.UTF16ToString(slice)
}

// LOWORD: Param içinden alt 16 bit
func LOWORD(l uintptr) int16 {
	return int16(l & 0xFFFF)
}

// HIWORD: Param içinden üst 16 bit
func HIWORD(l uintptr) int16 {
	return int16((l >> 16) & 0xFFFF)
}

// GET_X_LPARAM: lParam'dan X koordinatı
func GET_X_LPARAM(lp uintptr) int32 {
	return int32(LOWORD(lp))
}

// GET_Y_LPARAM: lParam'dan Y koordinatı
func GET_Y_LPARAM(lp uintptr) int32 {
	return int32(HIWORD(lp))
}
