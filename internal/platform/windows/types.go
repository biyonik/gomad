//go:build windows
// +build windows

package windows

/*
======================================================================================
🪟 Windows API Veri Türleri, Mesaj Kodları ve Yapısal Tanımlar - Platform Katmanı Çekirdeği
======================================================================================

Bu dosya, Windows işletim sistemi üzerinde pencere oluşturma, yönetme, sistem mesajlarını
işleme ve kullanıcı etkileşimlerini yakalama sürecinde kullanılan düşük seviye WinAPI
tiplerini içerir. Linux/macOS gibi platformlarda benzeri mekanizmalar farklıdır, bu yüzden
buradaki yapılar yalnızca Windows üzerinde geçerlidir ve doğrudan sistem çağrılarına hitap eder.

Amaç, Go uygulamasında Win32 API kullanırken her tipin yeniden tanımlanmasını engellemek,
anlaşılır bir soyutlama sunmak ve daha üst seviyedeki pencere yönetim modüllerinin sağlam ve
temiz bir temel üzerinde çalışmasını garanti etmektir. Bu yapılar olmadan CreateWindowEx,
DefWindowProc, MessageLoop gibi mekanizmaların kullanımı mümkün değildir.

Bu dosya aslında **pencere sisteminin anatomisidir.**
Handle nedir, mesaj nasıl akar, fare tıklaması nereden geçer, pencere boyutu nasıl tutulur,
her biri burada atom düzeyinde tanımlanmıştır.

@author Ahmet ALTUN
@github github.com/biyonik
@linkedin linkedin.com/in/biyonik
@email ahmet.altun60@gmail.com
*/

import (
	"unsafe"
)

/*
-----------------------
📌 WinAPI Temel Typedef
-----------------------
HANDLE, HWND, WPARAM vb. yapıların tamamı Windows'un C tabanlı mimarisinden gelir.
Burada Go karşılıkları verilmiştir — sistem fonksiyonlarıyla iletişimi mümkün kılar.
*/
type (
	HANDLE    uintptr // Genel amaçlı 64-bit/32-bit adres işaretçisi
	HWND      HANDLE  // Pencere handle'ı, tüm pencere işlemlerinin kimliği
	HINSTANCE HANDLE  // Çalışan uygulamanın instance adresi
	HICON     HANDLE  // Pencere ikonu için işaretçi
	HCURSOR   HANDLE  // İmleç işaretçisi
	HBRUSH    HANDLE  // Boyama ve arkaplan fırçası
	HMENU     HANDLE  // Menü handle'ı
	WPARAM    uintptr // Mesaj parametresi, ek data taşır
	LPARAM    uintptr // Mesaj parametresi, koordinat dahil veri taşır
	LRESULT   uintptr // Windows mesaj dönüş türü
	ATOM      uint16  // Sistem kaynaklarını temsil eden kısa kimlik
)

/*
📍 POINT Yapısı
Fare konumu ve mesajlarda koordinat tutmak için kullanılan temel tip.
*/
type POINT struct {
	X, Y int32
}

/*
📍 RECT Yapısı
Pencere boyutu, çizim alanı ve yerleşim hesaplamalarında kullanılan temel dikdörtgen alan tanımı.
*/
type RECT struct {
	Left, Top, Right, Bottom int32
}

/*
📍 MSG Yapısı - Windows Mesaj Kuyruğu Ögesi
Her pencere olayı mesaj döngüsünden geçer. Kullanıcı tıklar → sistem MSG üretir → uygulama işler.
*/
type MSG struct {
	Hwnd    HWND
	Message uint32
	WParam  WPARAM
	LParam  LPARAM
	Time    uint32
	Pt      POINT
}

/*
📍 WNDCLASSEX
Pencere sınıfı tanımlayan yapı — ikon, cursor, className gibi bilgiler burada tutulur.
Windows'ta pencere oluşturmanın ilk adımı budur.
*/
type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     HINSTANCE
	HIcon         HICON
	HCursor       HCURSOR
	HbrBackground HBRUSH
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       HICON
}

// Size, WNDCLASSEX yapısının RAM üzerindeki byte uzunluğunu döndürür.
// CreateWindowEx'le uyümlü çalışması için her zaman struct boyutunun bildirilmesi gerekir.
func (w *WNDCLASSEX) Size() uint32 {
	return uint32(unsafe.Sizeof(*w))
}

/*
=========================
Windows Mesaj Sabitleri
=========================
Her pencere olayı bir mesajla ifade edilir (mousemove, click, destroy vb).
*/
const (
	WM_DESTROY     = 0x0002
	WM_CLOSE       = 0x0010
	WM_PAINT       = 0x000F
	WM_MOUSEMOVE   = 0x0200
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_RBUTTONDOWN = 0x0204
	WM_RBUTTONUP   = 0x0205
	WM_MBUTTONDOWN = 0x0207
	WM_MBUTTONUP   = 0x0208
)

/*
=========================
Pencere Stil Sabitleri
=========================
Pencere çerçevesi, başlık barı, minimize tuşu vb. özellikleri belirler.
*/
const (
	WS_OVERLAPPED       = 0x00000000
	WS_CAPTION          = 0x00C00000
	WS_SYSMENU          = 0x00080000
	WS_THICKFRAME       = 0x00040000
	WS_MINIMIZEBOX      = 0x00020000
	WS_MAXIMIZEBOX      = 0x00010000
	WS_OVERLAPPEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX
)

/*
=========================
Ek UI Sabitleri
=========================
Sistem brush ID’leri, cursor, show/hide flag’leri vb.
*/
const (
	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001

	IDC_ARROW = 32512

	COLOR_WINDOW = 5

	SW_SHOW = 5
	SW_HIDE = 0

	CW_USEDEFAULT = ^0x7FFFFFFF // Varsayılan pencere pozisyonu
)

/*
GET_X_LPARAM & GET_Y_LPARAM
Windows LParam değerinden mouse koordinatlarını çeker.
*/
func GET_X_LPARAM(lp LPARAM) int {
	return int(int16(lp & 0xFFFF))
}

func GET_Y_LPARAM(lp LPARAM) int {
	return int(int16((lp >> 16) & 0xFFFF))
}
