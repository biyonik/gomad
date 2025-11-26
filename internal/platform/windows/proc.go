//go:build windows
// +build windows

package windows

/*
===========================================================================================
🔹 WinAPI Dynamic Linking & Sistem Fonksiyon Erişim Katmanı
===========================================================================================

Bu dosya, Go ile Windows işletim sistemi arasındaki en kritik köprülerden birini oluşturur.
Burada amaç; user32.dll ve kernel32.dll gibi temel WinAPI kütüphanelerini dinamik olarak
yükleyip, pencere yönetimi ve çekirdek seviye sistem işlemlerinde kullanılan işlevlere
erişim sağlamaktır.

WinAPI, grafik arayüz (UI), mesaj döngüsü, pencere davranışı ve input kontrolü gibi
konularda uygulamanın kalbidir. Ancak doğrudan C kodu kullanmak yerine, Go üzerinden bu
fonksiyonlara erişmek için *LazyDLL* ve *NewProc* mekanizmalarıyla fonksiyon pointer’ları
açığa çıkarılır. Böylece uygulama daha güvenli, portable olmayan ama çok güçlü bir native
yeteneğe sahip olur.

@author Ahmet ALTUN
@github github.com/biyonik
@linkedin linkedin.com/in/biyonik
@email ahmet.altun60@gmail.com
*/

import (
	"syscall"
)

var (
	// user32.dll: Windows kullanıcı arabirimi API'lerini içerir.
	// Pencere oluşturma, mesaj sistemi, input yönetimi gibi tüm UI operasyonlarının temelidir.
	user32 = syscall.NewLazyDLL("user32.dll")

	// kernel32.dll: Sistem seviyesi fonksiyonların (işlem, bellek, thread vb.) çekirdeğini temsil eder.
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	// =======================
	// user32.dll Fonksiyonları
	// =======================

	// procRegisterClassEx -> Windows pencere sınıfı kaydı
	// Bir pencere oluşturulmadan önce mutlaka sınıf tanımlanmalı & sisteme bildirilmelidir.
	procRegisterClassEx = user32.NewProc("RegisterClassExW")

	// procCreateWindowEx -> Pencere oluşturur ve HANDLE döndürür.
	programCreateWindowEx = user32.NewProc("CreateWindowExW")

	// procShowWindow -> Oluşturulan pencereyi gösterir (SW_SHOW vb. flag ile).
	procShowWindow = user32.NewProc("ShowWindow")

	// procDestroyWindow -> Bir pencereyi yok eder, kaynakları temizler.
	procDestroyWindow = user32.NewProc("DestroyWindow")

	// procDefWindowProc -> Varsayılan pencere mesaj işleyicisi.
	// Kullanıcı işleyemezse sistem burada devreye girer.
	procDefWindowProc = user32.NewProc("DefWindowProcW")

	// procGetMessage -> Mesaj kuyruğundan event çeker (blocking loop).
	procGetMessage = user32.NewProc("GetMessageW")

	// procTranslateMessage -> Klavye mesajlarını çözümler.
	procTranslateMessage = user32.NewProc("TranslateMessage")

	// procDispatchMessage -> Mesajı window procedure'e yollar.
	procDispatchMessage = user32.NewProc("DispatchMessageW")

	// procPostQuitMessage -> UI loop'u sonlandırmak için kullanılır.
	procPostQuitMessage = user32.NewProc("PostQuitMessage")

	// procSetWindowText -> Pencere başlığı değiştirme fonksiyonudur.
	procSetWindowText = user32.NewProc("SetWindowTextW")

	// procLoadCursor -> Sistem default cursorlarını yükler (ör. IDC_ARROW).
	procLoadCursor = user32.NewProc("LoadCursorW")

	// procSetWindowPos -> Pencere boyut ve pozisyonunu günceller.
	procSetWindowPos = user32.NewProc("SetWindowPos")

	// procGetWindowRect -> Belirtilen pencerenin sınırlayıcı dikdörtgeninin sınırlarını alır
	procGetWindowRect = user32.NewProc("GetWindowRect")

	// procGetSystemMetrics -> Windows API'sinden sistem ölçümlerini veya sistem yapılandırma ayarlarını alır.
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")

	// ==========================
	// kernel32.dll Fonksiyonları
	// ==========================

	// procGetModuleHandle -> Uygulamanın kendi instance adresini döndürür.
	// CreateWindowEx için çoğunlukla HINSTANCE burada elde edilir.
	procGetModuleHandle = kernel32.NewProc("GetModuleHandleW")
)

// StringToUTF16Ptr, Go string'ini WinAPI ile uyumlu *UTF16 pointer'a dönüştürür.
// Windows API fonksiyonları çoğunlukla UTF-16 bekler — bu da gerekli dönüşüm katmanıdır.
// Dönüş pointer'ı doğrudan C tarzı fonksiyonlara parametre olarak geçilebilir.
func StringToUTF16Ptr(s string) *uint16 {
	ptr, _ := syscall.UTF16PtrFromString(s)
	return ptr
}
