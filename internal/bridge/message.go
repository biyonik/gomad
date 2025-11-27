// ============================================================================
// GOMAD — BRIDGE MESSAGING LAYER
// ----------------------------------------------------------------------------
// Bu dosya, `bridge` paketinin kalbi olan mesaj sistemini tanımlar ve Go ile
// JavaScript arasında iki yönlü (bidirectional) güvenli iletişimin temelini
// oluşturur. Her mesaj JSON formatında taşınır, WebView içinden gönderilir ve
// yine Go tarafında parse edilerek işlenir.
//
// Bu dosyadaki yapı; fonksiyon çağırma, event gönderme, hata taşıma gibi tüm
// iletişim protokollerinin veri modelini oluşturur. Böylece GOMAD mimarisinde
// front-end ile back-end arasında sınır kaybolur, event–call–response zinciri
// akıcı bir şekilde işler. Amaç sadece veri taşımak değildir — aynı zamanda
// geliştiriciye anlaşılır, izlenebilir ve hataları tespit edilebilir bir
// köprü sunmaktır.
//
// Burada tasarlanan sistem neden önemlidir?
// - Çünkü JS → GO fonksiyon çağırma işlemi stabil ve type-safe şekilde yapılır.
// - Çünkü GO → JS event broadcast edebilir; yani Go aktif olarak mesaj atabilir.
// - Çünkü mesajlar timestamp içerir; debugging, logging, replay gibi alanlarda
//   geliştiricinin önünü açar.
// - Çünkü tüm yapı JSON tabanlıdır ve insan gözüyle okunabilir.
//
// Bu dosya, GOMAD köprüsünün kan dolaşımı gibidir — sistem çalıştığı sürece
// mesajlar akar, çağrılar yapılır, sonuçlar döner.
//
// @author    Ahmet ALTUN
// @github    github.com/biyonik
// @linkedin  linkedin.com/in/biyonik
// @email     ahmet.altun60@gmail.com
// ============================================================================

// Package bridge provides Go-JavaScript communication for GOMAD.
//
// Bu paket, Go backend ile WebView içindeki JavaScript arasında
// type-safe, bidirectional iletişim sağlar.
//
// Temel kavramlar:
//   - Message: Go ve JS arasında gidip gelen veri yapısı
//   - Binding: Go fonksiyonlarının JS'ten çağrılabilir hale getirilmesi
//   - Event: Go'dan JS'e broadcast edilen olaylar
package bridge

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of a bridge message.
// ----------------------------------------------------------------------
// Mesajın türünü belirten değer. Her mesaj bir tipe sahiptir ve bu tip,
// mesajın nasıl işleneceğini, nasıl karşılık döneceğini belirler.
//
// Türler:
//
//	call   → JS → GO fonksiyon çağrısı
//	result → GO → JS fonksiyon sonucu
//	error  → hata taşıyan mesaj
//	event  → tek yönlü yayın (broadcast)
type MessageType string

const (
	// MessageTypeCall is a function call from JS to Go.
	// JS bir Go fonksiyonunu çağırmak istediğinde bu tip kullanılır.
	MessageTypeCall MessageType = "call"

	// MessageTypeResult is a successful response from Go to JS.
	// Go fonksiyonu başarıyla çalıştığında sonuç bu tipte döner.
	MessageTypeResult MessageType = "result"

	// MessageTypeError is an error response from Go to JS.
	// Go fonksiyonu hata fırlattığında bu tip kullanılır.
	MessageTypeError MessageType = "error"

	// MessageTypeEvent is a broadcast event from Go to JS.
	// Go'dan JS'e tek yönlü bildirim göndermek için kullanılır.
	MessageTypeEvent MessageType = "event"
)

// ============================================================================
//
//	Message
//
// ----------------------------------------------------------------------------
// İletişimde taşınan ana veri paketidir. JSON olarak WebView'e gider ve gelir.
// Aynı tip struct ile hem request hem response üretilebilir.
// Örn: JS → GO fonksiyon çağrısı | GO → JS event yayını | hata mesajı dönüşü.
//
// İçerdiği alanlar; method adı, argümanlar, timestamp, event bilgisi, hata
// detayı gibi her şeyi kapsar. Esnek, genişletilebilir ve izlenebilir şekilde
// tasarlanmıştır.
type Message struct {
	// ID is a unique identifier for request-response matching.
	// JS tarafı bu ID ile hangi call'a response geldiğini eşleştirir.
	// Event mesajlarında boş olabilir.
	ID string `json:"id,omitempty"`

	// Type indicates the message type (call, result, error, event).
	Type MessageType `json:"type"`

	// Method is the name of the function to call (only for "call" type).
	Method string `json:"method,omitempty"`

	// Args contains the function arguments (only for "call" type").
	// JSON array olarak gelir, her eleman bir argüman.
	Args json.RawMessage `json:"args,omitempty"`

	// Result contains the function return value (only for "result" type).
	// Herhangi bir JSON değeri olabilir.
	Result json.RawMessage `json:"result,omitempty"`

	// Error contains error information (only for "error" type).
	Error *ErrorPayload `json:"error,omitempty"`

	// Event is the event name (only for "event" type).
	Event string `json:"event,omitempty"`

	// Data contains event data (only for "event" type").
	Data json.RawMessage `json:"data,omitempty"`

	// Timestamp is when the message was created (optional, for debugging).
	Timestamp int64 `json:"timestamp,omitempty"`
}

// ============================================================================
//
//	ErrorPayload
//
// ----------------------------------------------------------------------------
// GO tarafında oluşan hatayı JS'e taşımak için kullanılan veri yapısıdır.
// Hata kodu, mesajı ve opsiyonel açıklama (stack, context, detay) içerir.
type ErrorPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ---------------------------------------------------------------------------
// Standart hata kodları (sistemde evrensel olarak kullanılabilir)
const (
	ErrCodeUnknown        = -1
	ErrCodeMethodNotFound = -2
	ErrCodeInvalidArgs    = -3
	ErrCodeExecution      = -4
)

// ============================================================================
//
//	NewCallMessage
//
// ----------------------------------------------------------------------------
// JS tarafından GO metodunu çağırmak için mesaj oluşturur.
// id → benzersiz request kimliği
// method → çalıştırılacak fonksiyon adı
// args → fonksiyon argümanları JSON’a serialize edilir
func NewCallMessage(id, method string, args interface{}) (*Message, error) {
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:        id,
		Type:      MessageTypeCall,
		Method:    method,
		Args:      argsJSON,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

// ============================================================================
//
//	NewResultMessage
//
// ----------------------------------------------------------------------------
// Bir fonksiyon çağrısı başarıyla tamamlandığında GO → JS dönüş tipi.
// Result JSON formatına çevrilerek gönderilir.
func NewResultMessage(id string, result interface{}) (*Message, error) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:        id,
		Type:      MessageTypeResult,
		Result:    resultJSON,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

// ============================================================================
//
//	NewErrorMessage
//
// ----------------------------------------------------------------------------
// Go tarafında hata oluşursa, karşı tarafa bu yapı gönderilir.
// Hem insan tarafından okunabilir hem makineler tarafından işlenebilir.
func NewErrorMessage(id string, code int, message string, details string) *Message {
	return &Message{
		ID:   id,
		Type: MessageTypeError,
		Error: &ErrorPayload{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UnixMilli(),
	}
}

// ============================================================================
//
//	NewEventMessage
//
// ----------------------------------------------------------------------------
// JS'e broadcast event göndermek için kullanılır.
// Fonksiyon sonucu değildir → bildirimdir.
func NewEventMessage(event string, data interface{}) (*Message, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:      MessageTypeEvent,
		Event:     event,
		Data:      dataJSON,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

// ============================================================================
// ParseArgs — ParseResult — ParseData
// ----------------------------------------------------------------------------
// Gelen JSON mesaj içeriğini uygun tipe unmarshall eder.
// Pointer verilmezse veriyi dolduramaz → her zaman pointer bekler.
func (m *Message) ParseArgs(args interface{}) error {
	if m.Args == nil {
		return nil
	}
	return json.Unmarshal(m.Args, args)
}

func (m *Message) ParseResult(result interface{}) error {
	if m.Result == nil {
		return nil
	}
	return json.Unmarshal(m.Result, result)
}

func (m *Message) ParseData(data interface{}) error {
	if m.Data == nil {
		return nil
	}
	return json.Unmarshal(m.Data, data)
}

// ============================================================================
//
//	ToJSON / FromJSON
//
// ----------------------------------------------------------------------------
// Mesajı JSON’a çevirir veya JSON'dan geri struct’a dönüştürür.
// WebView köprüsünden geçişte en sık kullanılan iki fonksiyondur.
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

func FromJSON(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
