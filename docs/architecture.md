# GOMAD Mimari Dokümantasyonu

Bu belge, GOMAD framework'ünün iç mimarisini açıklar.

## 🏛️ Katmanlı Mimari

GOMAD, **Clean Architecture** prensiplerini takip eder. Her katman sadece altındaki katmana bağımlıdır.

```
┌─────────────────────────────────────────────────────────────────┐
│                         pkg/gomad                               │
│  PUBLIC API - Kullanıcıların gördüğü tek katman                 │
│  • Application struct                                           │
│  • Functional options (WithTitle, WithSize, ...)                │
│  • Bind, Emit, Eval metodları                                   │
├─────────────────────────────────────────────────────────────────┤
│                      internal/bridge                            │
│  BRIDGE LAYER - Go ↔ JavaScript iletişimi                       │
│  • Message: JSON mesaj yapısı                                   │
│  • Registry: Fonksiyon kaydı ve çağrısı                         │
│  • Bridge: İletişim koordinasyonu                               │
├─────────────────────────────────────────────────────────────────┤
│                     internal/webview                            │
│  WEBVIEW LAYER - Browser engine soyutlaması                     │
│  • WebView interface                                            │
│  • webview/webview_go wrapper                                   │
├─────────────────────────────────────────────────────────────────┤
│                    internal/platform                            │
│  PLATFORM LAYER - OS-specific implementasyonlar                 │
│  • Window interface                                             │
│  • windows/ - Win32 API                                         │
│  • darwin/  - Cocoa (gelecekte)                                 │
│  • linux/   - X11/Wayland (gelecekte)                           │
├─────────────────────────────────────────────────────────────────┤
│                     internal/errors                             │
│  ERROR TYPES - Tüm katmanların kullandığı hata tipleri          │
│  • BindingError                                                 │
│  • MessageError                                                 │
│  • WindowError                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## 📡 Bridge Protokolü

Go ve JavaScript arasındaki iletişim JSON mesajları ile sağlanır.

### Mesaj Tipleri

```
┌──────────────────────────────────────────────────────────────┐
│  MessageTypeCall     │  JS → Go fonksiyon çağrısı            │
│  MessageTypeResult   │  Go → JS başarılı yanıt               │
│  MessageTypeError    │  Go → JS hata yanıtı                  │
│  MessageTypeEvent    │  Go → JS broadcast event              │
└──────────────────────────────────────────────────────────────┘
```

### Call Akışı

```
JavaScript                    Bridge                      Go
    │                           │                          │
    │  call("greet", "Ahmet")   │                          │
    │ ─────────────────────────>│                          │
    │                           │  HandleMessage()         │
    │                           │ ─────────────────────────>│
    │                           │                          │
    │                           │  registry.Call("greet")  │
    │                           │ ─────────────────────────>│
    │                           │                          │
    │                           │<───────────────────────── │
    │                           │  return "Hello, Ahmet!"  │
    │                           │                          │
    │<───────────────────────── │                          │
    │  resolve("Hello, Ahmet!") │                          │
    │                           │                          │
```

### Event Akışı

```
Go                          Bridge                    JavaScript
 │                            │                           │
 │  Emit("notification", data)│                           │
 │ ──────────────────────────>│                           │
 │                            │  Eval("_handleEvent(...)")│
 │                            │ ─────────────────────────>│
 │                            │                           │
 │                            │                     listeners["notification"](data)
 │                            │                           │
```

## 🔧 Registry Sistemi

Registry, Go fonksiyonlarını isimle kaydeder ve reflection ile çağırır.

### Desteklenen Fonksiyon İmzaları

```go
func()                        // Void, argümansız
func() error                  // Sadece hata dönebilir
func() T                      // Tek değer döner
func() (T, error)             // Değer ve hata dönebilir
func(args...) (T, error)      // Argüman alır, değer ve hata döner
```

### Tip Dönüşümü

JSON → Go dönüşümü `encoding/json` paketi ile yapılır:

| JSON Tipi | Go Tipi |
|-----------|---------|
| string | string |
| number | int, float64 |
| boolean | bool |
| array | []T, []interface{} |
| object | map[string]interface{}, struct |
| null | nil, pointer |

### Hata Kodları

```go
ErrCodeUnknown        = -1  // Bilinmeyen hata
ErrCodeMethodNotFound = -2  // Fonksiyon bulunamadı
ErrCodeInvalidArgs    = -3  // Geçersiz argümanlar
ErrCodeExecution      = -4  // Çalışma hatası
```

## 🔒 Thread Safety

- **Registry**: Tüm metodlar concurrent-safe (sync.RWMutex)
- **Bridge**: Main goroutine'de kullanılmalı
- **WebView**: Platform kısıtlamaları (özellikle macOS main thread)

## 📦 Paket Yapısı

### pkg/ (Public)

```go
import "github.com/biyonik/gomad/pkg/gomad"

// Kullanıcılar sadece bu paketi import eder
app := gomad.New(gomad.WithTitle("My App"))
```

### internal/ (Private)

```go
// Bu paketler dışarıdan import EDİLEMEZ
// Go derleyicisi bunu engeller

import "github.com/biyonik/gomad/internal/bridge" // ❌ HATA!
```

## 🎨 Design Patterns

### 1. Functional Options

```go
// Builder pattern'a alternatif, daha Go idiomatic
app := gomad.New(
    gomad.WithTitle("Title"),
    gomad.WithSize(800, 600),
    gomad.WithDebug(true),
)
```

### 2. Interface Segregation

```go
// Küçük, odaklı interface'ler
type Evaluator interface {
    Eval(js string) error
}

// WebView bu interface'i implement eder
// Test'te mock kullanılabilir
```

### 3. Dependency Inversion

```go
// Bridge somut tipe değil, interface'e bağımlı
type Bridge struct {
    evaluator Evaluator  // Interface, not concrete type
}
```

## 🧪 Test Stratejisi

### Unit Tests

```go
// internal/bridge/registry_test.go
func TestRegistry_Register(t *testing.T) {
    r := NewRegistry()
    err := r.Register("test", func() {})
    // ...
}
```

### Integration Tests

```go
// WebView olmadan Bridge test etme
type mockEvaluator struct {
    lastJS string
}

func (m *mockEvaluator) Eval(js string) error {
    m.lastJS = js
    return nil
}
```

### E2E Tests

```go
// Gerçek WebView ile tam akış testi
// CI/CD'de headless mode gerektirir
```

## 📚 Referanslar

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [webview/webview](https://github.com/webview/webview)
