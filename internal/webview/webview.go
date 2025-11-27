// Package webview, GOMAD uygulamasında kullanılan WebView soyutlamasını sunar.
// Bu paket, platforma özel WebView implementasyonlarını tek bir arayüz altında toplar.
// Go tarafı ile JavaScript arasında güvenli ve tip güvenli bir iletişim köprüsü sağlar.
// Kod, webview/webview_go kütüphanesi üzerine inşa edilmiştir.
//
// Buradaki amaç, uygulamanın HTML/JS tabanlı arayüzünü Go koduyla etkileşimli
// bir şekilde yönetebilmek ve aynı zamanda platform bağımsız bir yapı sunmaktır.
//
// @author Ahmet ALTUN
// @github github.com/biyonik
// @linkedin linkedin.com/in/biyonik
// @email ahmet.altun60@gmail.com
package webview

import (
	"fmt"
	"sync"
	_ "unsafe"

	"github.com/biyonik/gomad/internal/bridge"
	webview "github.com/webview/webview_go"
)

// WebView, HTML içeriğini görüntüleyebilen ve Go tarafıyla iletişim kurabilen
// bir WebView örneğini temsil eder.
type WebView interface {
	// Navigate, verilen URL'ye gider.
	Navigate(url string)

	// SetHTML, HTML içeriğini doğrudan ayarlar.
	SetHTML(html string)

	// SetTitle, pencere başlığını ayarlar.
	SetTitle(title string)

	// SetSize, pencere boyutlarını ayarlar.
	// hint: 0=none, 1=min, 2=max, 3=fixed
	SetSize(width, height int, hint int)

	// Eval, JavaScript kodunu yürütür.
	Eval(js string) error

	// Bind, Go fonksiyonunu JavaScript tarafına bağlar.
	// Deprecated: Tip güvenliği için Bridge.Bind kullanılmalıdır.
	Bind(name string, fn interface{}) error

	// Init, JavaScript başlatma kodunu çalıştırır.
	Init(js string) error

	// Run, WebView olay döngüsünü başlatır.
	Run()

	// Destroy, WebView'i kapatır ve kaynakları serbest bırakır.
	Destroy()

	// Window, native pencere tutamacını döndürür.
	Window() uintptr
}

// WebViewImpl, webview/webview_go kullanılarak oluşturulmuş WebView implementasyonudur.
type WebViewImpl struct {
	w      webview.WebView
	bridge *bridge.Bridge

	// Durum bilgisi
	ready   bool
	readyMu sync.RWMutex

	// Geri çağırma fonksiyonları
	onReady func()
	mu      sync.Mutex
}

// Options, WebView oluşturulurken yapılandırma seçeneklerini temsil eder.
type Options struct {
	// Title, pencere başlığıdır.
	Title string

	// Width, pencere genişliğini piksel cinsinden belirler.
	Width int

	// Height, pencere yüksekliğini piksel cinsinden belirler.
	Height int

	// Debug, geliştirici araçlarını (F12) etkinleştirir.
	Debug bool

	// URL, başlangıçta gidilecek URL'dir.
	// Boşsa SetHTML kullanılmalıdır.
	URL string

	// HTML, başlangıç HTML içeriğidir.
	// URL belirtilmişse göz ardı edilir.
	HTML string
}

// DefaultOptions, mantıklı varsayılan seçenekleri döndürür.
func DefaultOptions() Options {
	return Options{
		Title:  "GOMAD Application",
		Width:  800,
		Height: 600,
		Debug:  false,
	}
}

// New, verilen seçeneklerle yeni bir WebView oluşturur.
func New(opts Options) (*WebViewImpl, error) {
	// webview/webview_go oluştur
	w := webview.New(opts.Debug)
	if w == nil {
		return nil, fmt.Errorf("failed to create webview")
	}

	impl := &WebViewImpl{
		w: w,
	}

	// Bridge oluştur
	impl.bridge = bridge.NewBridge(impl)

	// Pencere ayarları
	w.SetTitle(opts.Title)
	w.SetSize(opts.Width, opts.Height, webview.HintNone)

	// Go fonksiyonlarını JS'ten çağırma mekanizması
	// webview/webview_go'nun Bind fonksiyonu string alır ve string döner
	err := w.Bind("__gomad_invoke", func(msgJSON string) string {
		return impl.bridge.HandleMessage(msgJSON)
	})
	if err != nil {
		return nil, err
	}

	// Bridge'i başlat ve invoke wrapper'ı ekle
	initJS := bridge.JSBridgeCode + `
	
	// Override the call mechanism to use __gomad_invoke
	(function() {
		const originalCall = window.gomad.call;
		window.gomad.call = async function(method, ...args) {
			const id = 'js_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
			
			const message = {
				id: id,
				type: 'call',
				method: method,
				args: args,
				timestamp: Date.now()
			};
			
			try {
				// __gomad_invoke returns a Promise, so we need await
				const responseJSON = await __gomad_invoke(JSON.stringify(message));
				
				if (responseJSON) {
					const response = JSON.parse(responseJSON);
					if (response.type === 'error') {
						const error = new Error(response.error.message);
						error.code = response.error.code;
						throw error;
					} else {
						return response.result;
					}
				}
				return undefined;
			} catch (e) {
				// JSON parse hatası değilse, orijinal hatayı fırlat
				if (e instanceof SyntaxError) {
					console.error('GOMAD: Invalid response JSON:', e);
					throw new Error('Invalid response from Go');
				}
				throw e;
			}
		};
		
		console.log('GOMAD: Call mechanism initialized');
	})();
	`

	w.Init(initJS)

	// İçerik yükle
	if opts.URL != "" {
		w.Navigate(opts.URL)
	} else if opts.HTML != "" {
		w.SetHtml(opts.HTML)
	}

	return impl, nil
}

// ==================== WebView Interface Implementation ====================

// Navigate, WebView'i verilen URL'ye yönlendirir.
func (wv *WebViewImpl) Navigate(url string) {
	wv.w.Navigate(url)
}

// SetHTML, WebView içerisine HTML içeriği yükler.
func (wv *WebViewImpl) SetHTML(html string) {
	wv.w.SetHtml(html)
}

// SetTitle, WebView pencere başlığını ayarlar.
func (wv *WebViewImpl) SetTitle(title string) {
	wv.w.SetTitle(title)
}

// SetSize, WebView pencere boyutlarını ayarlar.
func (wv *WebViewImpl) SetSize(width, height int, hint int) {
	wv.w.SetSize(width, height, webview.Hint(hint))
}

// Eval, WebView içinde JavaScript kodunu yürütür.
func (wv *WebViewImpl) Eval(js string) error {
	wv.w.Eval(js)
	return nil // webview/webview_go hata dönmüyor
}

// Bind, düşük seviyede Go fonksiyonunu JS tarafına bağlar.
// Yeni kod için Bridge.Bind kullanılması önerilir.
func (wv *WebViewImpl) Bind(name string, fn interface{}) error {
	err := wv.w.Bind(name, fn)
	if err != nil {
		return err
	}
	return nil
}

// Init, WebView'e JavaScript başlatma kodunu uygular.
func (wv *WebViewImpl) Init(js string) error {
	wv.w.Init(js)
	return nil
}

// Run, WebView olay döngüsünü başlatır. Pencere kapanana kadar bloklar.
func (wv *WebViewImpl) Run() {
	wv.w.Run()
}

// Destroy, WebView'i kapatır ve kaynakları serbest bırakır.
func (wv *WebViewImpl) Destroy() {
	wv.w.Destroy()
}

// Window, native pencere tutamacını döndürür.
func (wv *WebViewImpl) Window() uintptr {
	return uintptr(wv.w.Window())
}

// ==================== Bridge Access ====================

// Bridge, WebView ile JS arasındaki iletişim köprüsünü döndürür.
func (wv *WebViewImpl) Bridge() *bridge.Bridge {
	return wv.bridge
}

// BindFunc, Bridge üzerinden fonksiyon bağlamayı kolaylaştırır.
func (wv *WebViewImpl) BindFunc(name string, fn interface{}) error {
	return wv.bridge.Bind(name, fn)
}

// Emit, JavaScript tarafına bir olay gönderir.
func (wv *WebViewImpl) Emit(event string, data interface{}) error {
	return wv.bridge.Emit(event, data)
}

// Size hint constants
const (
	HintNone  = int(webview.HintNone)
	HintMin   = int(webview.HintMin)
	HintMax   = int(webview.HintMax)
	HintFixed = int(webview.HintFixed)
)
