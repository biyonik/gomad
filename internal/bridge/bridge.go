package bridge

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

//
// =============================================================
//  BRIDGE — Go ↔ JavaScript Haberleşme Katmanı
// -------------------------------------------------------------
// Bu yapı GOMAD mimarisinin can damarlarından biridir. Amaç; Go tarafındaki
// fonksiyonların JavaScript tarafından çağrılabilmesini, aynı şekilde Go'nun
// JS’e veri, event ve cevap gönderebilmesini sağlamaktır.
//
// Yani:
//   - UI WebView içinde çalışır (HTML/JS)
//   - İş mantığı Go’da çalışır
//   - Bridge ikisi arasında görünmez bir köprüdür.
//
// Kullanıcı arayüzü bir butona bastığında Go fonksiyonu tetiklenebilir,
// Go tarafında oluşan olay JS'e event olarak aktarılabilir.
//
// Tüm fonksiyonlar thread-safe olarak tasarlanmıştır — yani UI sürekli event
// gönderirken aynı anda JS fonksiyon çağırabilir, veriler çakışmaz.
//
// @author   Ahmet ALTUN
// @github   github.com/biyonik
// @linkedin linkedin.com/in/biyonik
// @email    ahmet.altun60@gmail.com
// =============================================================
//

// ============================================================
// EVAL INTERFACE
// ------------------------------------------------------------
// WebView içerisinde JavaScript kodu çalıştırmak için kullanılan soyutlama.
// Bu interface doğrudan WebView bağımlılığını ortadan kaldırır.
// Böylece ister WebView Go, ister Wry, ister Tauri embed edilsin
// Bridge aynı kodla çalışmaya devam eder.
// ============================================================
type Evaluator interface {
	// Eval executes JavaScript code in the WebView.
	// JS kodunu string olarak alır ve o anda tarayıcıya enjekte eder.
	// UI güncelleme, event tetikleme gibi tüm işlemler bu kanal üzerinden yapılır.
	Eval(js string) error
}

// ============================================================
// BRIDGE STRUCT
// ------------------------------------------------------------
// Go–JS haberleşmesinin tamamı bu yapıda yönetilir.
//
// Sorumlulukları:
//
//	✓ JS'ten gelen fonksiyon çağrılarını Go fonksiyonlarına yönlendirir
//	✓ Go içindeki fonksiyonları JS'e bind eder
//	✓ Go’dan JS’e event broadcast edebilir
//	✓ Gelecekte: Go → JS fonksiyon çağrımı ve async cevap yakalama
//
// thread-safe olması için mutex ve atomic sayaçlar kullanılır.
// ============================================================
type Bridge struct {
	evaluator Evaluator // JavaScript çalıştırmak için gerekli eval interface’i
	registry  *Registry // Kayıtlı Go fonksiyonlarını tutar

	eventListeners map[string][]func(data interface{}) // JS event aboneleri
	eventMu        sync.RWMutex                        // event eşzamanlama

	msgIDCounter uint64                   // JS’e giden çağrılarda id üretmek için atomic sayaç
	pendingCalls map[string]chan *Message // JS’ten gelecek async cevaplar bekletilir
	pendingMu    sync.RWMutex             // pending işler eşzamanlı çalışabilir

	initialized bool // JS bridge kodu yüklendi mi?
	initMu      sync.RWMutex
}

// ============================================================
// NewBridge()
// ------------------------------------------------------------
// Yeni bir köprü oluşturur. Evaluator verilmesi zorunludur çünkü
// Bridge JS kodlarının yürütülmesini sağlayamazsa haberleşme kurulamaz.
//
// Amaç: Bridge soyut katman olmalı — UI teknolojisi değişse bile
// iletişim protokolü sabit kalmalıdır.
// ============================================================
func NewBridge(evaluator Evaluator) *Bridge {
	return &Bridge{
		evaluator:      evaluator,
		registry:       NewRegistry(),
		eventListeners: make(map[string][]func(data interface{})),
		pendingCalls:   make(map[string]chan *Message),
	}
}

// ============================================================
// FUNCTION BINDING
// ------------------------------------------------------------

// Bind()
// ------------------------------------------------------------
// Go fonksiyonunu JS’ten çağrılabilir hale getirir.
// Örn:
//
//	bridge.Bind("add", func(a,b int) int { return a+b })
//
// JS tarafında artık:
//
//	const v = await gomad.call("add", 3,4)
//
// Bu yapı masaüstü arayüz ile native backend etkileşimini sağlar.
// ============================================================
func (b *Bridge) Bind(name string, fn interface{}) error { return b.registry.Register(name, fn) }

// Unbind() → kaydı kaldırır
// ------------------------------------------------------------
func (b *Bridge) Unbind(name string) bool { return b.registry.Unregister(name) }

// IsBound() → fonksiyon bağlı mı sorgular
// ------------------------------------------------------------
func (b *Bridge) IsBound(name string) bool { return b.registry.Has(name) }

// ListBindings() → tüm kayıtlı fonksiyonları listeler
// ------------------------------------------------------------
func (b *Bridge) ListBindings() []string { return b.registry.List() }

// Registry returns the bridge registry instance.
// Public erişim için getter.
func (b *Bridge) Registry() *Registry {
	return b.registry
}

// ============================================================
// MESSAGE HANDLING
// ------------------------------------------------------------
// JavaScript'ten gelen her mesaj buradan geçer.
// Eğer mesaj "call" tipindeyse Go’da karşılık bulan fonksiyon çalıştırılır
// ve çıktısı yine JSON olarak JS'e döndürülür.
//
// MessageTypeResult ve Error ise, bunlar Go → JS async request cevabıdır.
// ============================================================
func (b *Bridge) HandleMessage(msgJSON string) string {
	msg, err := FromJSON([]byte(msgJSON))
	if err != nil {
		errMsg := NewErrorMessage("", ErrCodeUnknown, "failed to parse message", err.Error())
		result, _ := errMsg.ToJSON()
		return string(result)
	}

	var response *Message

	switch msg.Type {
	case MessageTypeCall:
		// JS → Go fonksiyon çağrısı
		response = b.registry.CallWithMessage(msg)

	case MessageTypeResult, MessageTypeError:
		// Go → JS async cevabı
		b.handlePendingResponse(msg)
		return "" // JS’e tekrar cevap göndermeye gerek yok

	default:
		response = NewErrorMessage(msg.ID, ErrCodeUnknown,
			fmt.Sprintf("unknown message type: %s", msg.Type), "")
	}

	result, _ := response.ToJSON()
	return string(result)
}

// handlePendingResponse()
// ------------------------------------------------------------
// JS’e async fonksiyon göndermemiz durumunda gelen cevabı yakalar.
// ID eşleşirse cevap bekleyen goroutine devam ettirilir.
// UI donmadan async veri dönüşü sağlanır.
// ============================================================
func (b *Bridge) handlePendingResponse(msg *Message) {
	if msg.ID == "" {
		return
	}

	b.pendingMu.RLock()
	ch, exists := b.pendingCalls[msg.ID]
	b.pendingMu.RUnlock()

	if exists {
		ch <- msg
		b.pendingMu.Lock()
		delete(b.pendingCalls, msg.ID)
		close(ch)
		b.pendingMu.Unlock()
	}
}

// ============================================================
// EVENTS — Go → JS Broadcast
// ------------------------------------------------------------
// Bridge.Emit("user:login", {id:1})
// JS tarafı:
//
//	gomad.on("user:login", data => console.log(data))
//
// Yani UI backend olaylarını canlı dinleyebilir.
// Socket gerekmez, WebView üzerinde uçtan uca data akışı.
// ============================================================
func (b *Bridge) Emit(event string, data interface{}) error {
	msg, err := NewEventMessage(event, data)
	if err != nil {
		return fmt.Errorf("failed to create event message: %w", err)
	}

	msgJSON, err := msg.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	js := fmt.Sprintf("window.gomad && window.gomad._handleEvent(%s)", string(msgJSON))
	return b.evaluator.Eval(js)
}

// ============================================================
// INIT() — Köprünün JS Kodunu WebView'e Enjekte Eder
// ------------------------------------------------------------
// Bu fonksiyon bir kere çalıştırılmalıdır.
// UI yüklenir yüklenmez Bridge aktif hale gelir.
//
// Aksi halde JS → Go çağrıları mümkün olmaz.
// ============================================================
func (b *Bridge) Init() error {
	b.initMu.Lock()
	defer b.initMu.Unlock()

	if b.initialized {
		return nil
	}

	if err := b.evaluator.Eval(JSBridgeCode); err != nil {
		return fmt.Errorf("failed to inject bridge code: %w", err)
	}

	b.initialized = true
	return nil
}

// IsInitialized() → Bridge aktif mi?
func (b *Bridge) IsInitialized() bool {
	b.initMu.RLock()
	defer b.initMu.RUnlock()
	return b.initialized
}

// generateMsgID() → Async istekler için benzersiz ID üretir.
func (b *Bridge) generateMsgID() string {
	id := atomic.AddUint64(&b.msgIDCounter, 1)
	return fmt.Sprintf("gomad_%d", id)
}

// GenerateTSDefinitions, frontend için .d.ts dosyasını belirtilen yola yazar.
// Bu fonksiyon main.go içinden çağrılabilir.
func (b *Bridge) GenerateTSDefinitions(path string) error {
	defs := b.registry.GenerateTypeDefinitions()
	return os.WriteFile(path, []byte(defs), 0644)
}

// ============================================================
// AŞAĞIDAKİ KOD JS TARAFINDA ÇALIŞIR
// ------------------------------------------------------------
// Bridge başarılı şekilde initialize edildiğinde WebView’e inject edilir.
// Artık window.gomad.call(...) ve window.gomad.on(...) kullanılabilir.
// ============================================================
const JSBridgeCode = `
(function() {
    'use strict';
    
    // Zaten yüklüyse tekrar yükleme
    if (window.gomad && window.gomad._initialized) {
        return;
    }
    
    // Pending promises for call responses
    const pendingCalls = new Map();
    
    // Event listeners
    const eventListeners = new Map();
    
    // Generate unique ID
    let callIdCounter = 0;
    function generateId() {
        return 'js_' + (++callIdCounter);
    }
    
    window.gomad = {
        _initialized: true,
        
        // Call a Go function
        // Usage: const result = await window.gomad.call("functionName", arg1, arg2);
        call: function(method, ...args) {
            return new Promise((resolve, reject) => {
                const id = generateId();
                
                const message = {
                    id: id,
                    type: 'call',
                    method: method,
                    args: args,
                    timestamp: Date.now()
                };
                
                pendingCalls.set(id, { resolve, reject });
                
                // Send to Go
                // WebView kütüphanesine göre bu değişebilir
                // webview/webview_go için: window.external.invoke
                try {
                    if (window.external && window.external.invoke) {
                        // webview/webview_go
                        window.external.invoke(JSON.stringify(message));
                    } else if (window.webkit && window.webkit.messageHandlers && window.webkit.messageHandlers.gomad) {
                        // WKWebView (macOS)
                        window.webkit.messageHandlers.gomad.postMessage(message);
                    } else {
                        reject(new Error('No bridge available'));
                    }
                } catch (e) {
                    pendingCalls.delete(id);
                    reject(e);
                }
            });
        },
        
        // Subscribe to an event
        // Usage: window.gomad.on("eventName", (data) => { ... });
        on: function(event, callback) {
            if (!eventListeners.has(event)) {
                eventListeners.set(event, []);
            }
            eventListeners.get(event).push(callback);
            
            // Return unsubscribe function
            return () => {
                const listeners = eventListeners.get(event);
                if (listeners) {
                    const index = listeners.indexOf(callback);
                    if (index > -1) {
                        listeners.splice(index, 1);
                    }
                }
            };
        },
        
        // Unsubscribe from an event
        off: function(event, callback) {
            const listeners = eventListeners.get(event);
            if (listeners) {
                if (callback) {
                    const index = listeners.indexOf(callback);
                    if (index > -1) {
                        listeners.splice(index, 1);
                    }
                } else {
                    eventListeners.delete(event);
                }
            }
        },
        
        // Internal: Handle response from Go
        _handleResponse: function(msgJson) {
            try {
                const msg = typeof msgJson === 'string' ? JSON.parse(msgJson) : msgJson;
                
                if (!msg.id) return;
                
                const pending = pendingCalls.get(msg.id);
                if (!pending) return;
                
                pendingCalls.delete(msg.id);
                
                if (msg.type === 'error') {
                    const error = new Error(msg.error.message);
                    error.code = msg.error.code;
                    error.details = msg.error.details;
                    pending.reject(error);
                } else if (msg.type === 'result') {
                    pending.resolve(msg.result);
                }
            } catch (e) {
                console.error('GOMAD: Failed to handle response:', e);
            }
        },
        
        // Internal: Handle event from Go
        _handleEvent: function(msgJson) {
            try {
                const msg = typeof msgJson === 'string' ? JSON.parse(msgJson) : msgJson;
                
                if (msg.type !== 'event' || !msg.event) return;
                
                const listeners = eventListeners.get(msg.event);
                if (listeners) {
                    const data = msg.data;
                    listeners.forEach(callback => {
                        try {
                            callback(data);
                        } catch (e) {
                            console.error('GOMAD: Event listener error:', e);
                        }
                    });
                }
            } catch (e) {
                console.error('GOMAD: Failed to handle event:', e);
            }
        }
    };
    
    console.log('GOMAD Bridge initialized');
})();
`
