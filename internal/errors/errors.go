// Package errors
//
// Bu paket, GOMAD çatısı altında kullanılan merkezi hata yönetim sistemini temsil eder.
// Yazılımın büyüyüp karmaşıklaştığı durumlarda aynı türden hataları yakalayabilmek,
// analiz edebilmek ve daha anlamlı bir yaklaşımla işleyebilmek için tiplenmiş hata
// modelleri kullanılır.
//
// Burada amaç sadece bir *string* olarak hata döndürmek değil, yapısal olarak
// hata üretmek ve bu hataları yakalarken `errors.Is()`, `errors.As()` gibi Go'nun
// modern hata kontrol mekanizmalarını etkin biçimde kullanabilmektir.
//
// Bu paket içinde hem basit anlamda *sentinel errors* dediğimiz sabit hatalar,
// hem de daha detaylı bağlam taşıyabilen struct tabanlı hata modelleri
// tanımlanmıştır. Böylece hem düşük seviye hem yüksek seviye mantık
// tek noktadan yönetilir.
//
// @author   Ahmet ALTUN
// @github   github.com/biyonik
// @linkedin linkedin.com/in/biyonik
// @email    ahmet.altun60@gmail.com
package errors

import (
	"errors"
	"fmt"
)

// ─────────────────────────────────────────────────────────────────────────────
// SENTINEL ERROR TANIMLARI
// Yazılım genelinde tekrar eden belirli durumlar için oluşturulmuş sabit hata nesneleri.
// Bunlar `errors.Is()` ile doğrudan karşılaştırılarak kontrol yapılabilir.
// ─────────────────────────────────────────────────────────────────────────────

var (
	// ErrNotReady → Bir bileşene erişmeye çalıştığımızda henüz hazır olmadığı
	// durumda dönen standart hata. Örn: Başlatılmamış bir servise erişme.
	ErrNotReady = errors.New("component not ready")

	// ErrAlreadyExists → Aynı kaynağı tekrar kaydetmeye veya oluşturmaya çalıştığımızda
	// ortaya çıkar. Registry, map ya da sistemsel kayıt alanlarında sıkça kullanılır.
	ErrAlreadyExists = errors.New("already exists")

	// ErrNotFound → Aranan kaynağın mevcut olmaması durumunda fırlatılan hata.
	// Özellikle arama, sorgulama, API dönüşleri için ideal.
	ErrNotFound = errors.New("not found")

	// ErrInvalidArgument → Bir fonksiyona hatalı parametre verildiğinde üretilir.
	ErrInvalidArgument = errors.New("invalid argument")

	// ErrClosed → Kapalı veya sonlandırılmış bir kaynak üzerinde işlem yapılmaya
	// çalışıldığında dönen hata.
	ErrClosed = errors.New("resource closed")
)

// ─────────────────────────────────────────────────────────────────────────────
// BindingError
// Go fonksiyonlarının JS tarafına bağlanması (binding) esnasında oluşabilecek
// spesifik hataları temsil eder. WebAssembly, JS Bridge veya runtime extensions
// gibi noktalarda fonksiyon eşleşmesi yapılamadığında ya da parametre uyumsuzluğu
// yaşandığında kullanılmak üzere tasarlanmıştır.
// ─────────────────────────────────────────────────────────────────────────────

// BindingError → Bağlama sürecinde oluşan hataların tutulduğu yapı.
type BindingError struct {
	FunctionName string // Bağlanmaya çalışılan fonksiyon adı
	Reason       string // Neden başarısız olduğu
	Cause        error  // Alt hata (varsa zincirlenebilir hata)
}

// Error → error interface gereği insan tarafından okunabilir hata çıktısı üretir.
func (e *BindingError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("binding '%s' failed: %s: %v",
			e.FunctionName, e.Reason, e.Cause)
	}
	return fmt.Sprintf("binding '%s' failed: %s",
		e.FunctionName, e.Reason)
}

// Unwrap → zincirli hata sistemlerinde altta yatan hatayı yakalamak için kullanılır.
func (e *BindingError) Unwrap() error { return e.Cause }

// NewBindingError → Yeni bir BindingError örneği üreten yardımcı yapıcı fonksiyon.
func NewBindingError(funcName, reason string, cause error) *BindingError {
	return &BindingError{
		FunctionName: funcName,
		Reason:       reason,
		Cause:        cause,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// MessageError
// Go ↔︎ JavaScript veya başka sistemler arası mesaj alışverişi sırasında yaşanan
// hataları kapsar. Bir mesajın işlenememesi, encode/decode hatası, routing sorunu
// veya operasyon başarısızlıkları bu yapı ile daha anlamlı şekilde taşınır.
// ─────────────────────────────────────────────────────────────────────────────

// MessageError → Mesaj sürecindeki problemlerin detayını tutan hata yapısı.
type MessageError struct {
	MessageID string // Hata üreten mesajın kimliği
	Operation string // Hangi işlemde hata oluştu
	Reason    string // Neden başarısız olduğu
	Cause     error  // Alt hata (opsiyonel)
}

// Error → Hatanın okunabilir formatını döner.
func (e *MessageError) Error() string {
	if e.MessageID != "" {
		if e.Cause != nil {
			return fmt.Sprintf("message %s: %s failed: %s: %v",
				e.MessageID, e.Operation, e.Reason, e.Cause)
		}
		return fmt.Sprintf("message %s: %s failed: %s",
			e.MessageID, e.Operation, e.Reason)
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s failed: %s: %v", e.Operation, e.Reason, e.Cause)
	}
	return fmt.Sprintf("%s failed: %s", e.Operation, e.Reason)
}

// Unwrap → Alt hatayı zincirden çekmeye yarar.
func (e *MessageError) Unwrap() error { return e.Cause }

// NewMessageError → Mesaja bağlı hata üretmek için fabrika fonksiyonu.
func NewMessageError(msgID, operation, reason string, cause error) *MessageError {
	return &MessageError{
		MessageID: msgID,
		Operation: operation,
		Reason:    reason,
		Cause:     cause,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// WindowError
// Pencere sistemi / UI işlemleri sırasında ortaya çıkan operasyonel hatalar için
// kullanılır. Örn: pencere oluşturulamaması, render hatası, event-hook hataları...
// ─────────────────────────────────────────────────────────────────────────────

// WindowError → UI/pencere yönetimine özgü hata modeli.
type WindowError struct {
	Operation string // Hangi işlemde hata gerçekleşti
	Reason    string // Hata nedeni
	Cause     error  // Alt neden (varsa)
}

// Error → Hatanın okunabilir hâlini üretir.
func (e *WindowError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("window %s failed: %s: %v",
			e.Operation, e.Reason, e.Cause)
	}
	return fmt.Sprintf("window %s failed: %s", e.Operation, e.Reason)
}

// Unwrap → Alt hata erişimi sağlar.
func (e *WindowError) Unwrap() error { return e.Cause }

// NewWindowError → Yeni bir WindowError oluşturur.
func NewWindowError(operation, reason string, cause error) *WindowError {
	return &WindowError{
		Operation: operation,
		Reason:    reason,
		Cause:     cause,
	}
}
