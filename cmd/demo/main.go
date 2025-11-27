// Hello World example for GOMAD
//
// Bu örnek:
// 1. WebView ile pencere açar
// 2. Go fonksiyonlarını JavaScript'e bind eder
// 3. JS'ten Go fonksiyonlarını çağırır
// 4. Go'dan JS'e event gönderir
//
// Çalıştırma:
//
//	CGO_ENABLED=1 go run ./cmd/examples/hello-world
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/biyonik/gomad/internal/webview"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

func main() {
	log.Println("GOMAD Hello World Example")
	log.Println("==========================")

	// WebView oluştur
	wv, err := webview.New(webview.Options{
		Title:  "GOMAD - Hello World",
		Width:  900,
		Height: 700,
		Debug:  true, // F12 ile DevTools açılabilir
	})
	if err != nil {
		log.Fatalf("Failed to create webview: %v", err)
	}

	// Go fonksiyonlarını bind et
	bridge := wv.Bridge()

	// 1. Basit fonksiyon - argümansız, string döner
	bridge.Bind("getVersion", func() string {
		return "GOMAD v0.1.0"
	})

	// 2. Tek argümanlı fonksiyon
	bridge.Bind("greet", func(name string) string {
		log.Printf("[Go] greet called with: %s", name)
		return fmt.Sprintf("Merhaba, %s! 🎉", name)
	})

	// 3. Çoklu argümanlı fonksiyon
	bridge.Bind("add", func(a, b int) int {
		log.Printf("[Go] add called with: %d, %d", a, b)
		return a + b
	})

	// 4. Karmaşık tip dönen fonksiyon
	bridge.Bind("getUser", func(id int) User {
		return User{ID: id, Username: "biyonik", CreatedAt: time.Now()}
	})

	// 5. Hata dönebilen fonksiyon
	bridge.Bind("divide", func(a, b float64) (float64, error) {
		log.Printf("[Go] divide called with: %.2f, %.2f", a, b)
		if b == 0 {
			return 0, fmt.Errorf("sıfıra bölme hatası")
		}
		return a / b, nil
	})

	// 6. Uzun süren işlem (simülasyon)
	bridge.Bind("longTask", func(seconds int) string {
		log.Printf("[Go] longTask called, will take %d seconds", seconds)
		time.Sleep(time.Duration(seconds) * time.Second)
		return fmt.Sprintf("%d saniye sonra tamamlandı!", seconds)
	})

	// HTML içeriği
	html := `
<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GOMAD Hello World</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            color: #fff;
            min-height: 100vh;
            padding: 20px;
        }
        
        .container {
            max-width: 800px;
            margin: 0 auto;
        }
        
        h1 {
            text-align: center;
            margin-bottom: 10px;
            font-size: 2.5em;
            background: linear-gradient(to right, #e94560, #ff6b6b);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        
        .subtitle {
            text-align: center;
            color: #aaa;
            margin-bottom: 30px;
        }
        
        .card {
            background: rgba(255,255,255,0.1);
            border-radius: 10px;
            padding: 20px;
            margin-bottom: 20px;
            backdrop-filter: blur(10px);
        }
        
        .card h2 {
            color: #e94560;
            margin-bottom: 15px;
            font-size: 1.2em;
        }
        
        .test-row {
            display: flex;
            gap: 10px;
            margin-bottom: 10px;
            flex-wrap: wrap;
        }
        
        input {
            flex: 1;
            padding: 10px;
            border: none;
            border-radius: 5px;
            background: rgba(255,255,255,0.2);
            color: white;
            font-size: 14px;
            min-width: 100px;
        }
        
        input::placeholder {
            color: rgba(255,255,255,0.5);
        }
        
        button {
            padding: 10px 20px;
            border: none;
            border-radius: 5px;
            background: linear-gradient(to right, #e94560, #ff6b6b);
            color: white;
            cursor: pointer;
            font-size: 14px;
            font-weight: bold;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 20px rgba(233, 69, 96, 0.4);
        }
        
        button:active {
            transform: translateY(0);
        }
        
        button:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }
        
        .result {
            margin-top: 10px;
            padding: 10px;
            background: rgba(0,0,0,0.3);
            border-radius: 5px;
            font-family: monospace;
            white-space: pre-wrap;
            word-break: break-all;
        }
        
        .result.success {
            border-left: 3px solid #4caf50;
        }
        
        .result.error {
            border-left: 3px solid #f44336;
        }
        
        .result.pending {
            border-left: 3px solid #ff9800;
        }
        
        .events-log {
            max-height: 150px;
            overflow-y: auto;
        }
        
        .event-item {
            padding: 5px;
            margin: 2px 0;
            background: rgba(0,0,0,0.2);
            border-radius: 3px;
            font-size: 12px;
        }
        
        .footer {
            text-align: center;
            margin-top: 20px;
            color: #666;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 GOMAD</h1>
        <p class="subtitle">Go + Angular Desktop Framework - Bridge Test</p>
        
        <!-- Test 1: Get Version -->
        <div class="card">
            <h2>📋 1. Versiyon Al (Argümansız)</h2>
            <button onclick="testVersion()">getVersion()</button>
            <div id="result-version" class="result">Sonuç burada görünecek...</div>
        </div>
        
        <!-- Test 2: Greet -->
        <div class="card">
            <h2>👋 2. Selamlama (String Argüman)</h2>
            <div class="test-row">
                <input type="text" id="name-input" placeholder="İsminizi yazın..." value="Ahmet">
                <button onclick="testGreet()">greet(name)</button>
            </div>
            <div id="result-greet" class="result">Sonuç burada görünecek...</div>
        </div>
        
        <!-- Test 3: Add -->
        <div class="card">
            <h2>➕ 3. Toplama (İki Int Argüman)</h2>
            <div class="test-row">
                <input type="number" id="num-a" placeholder="a" value="5">
                <input type="number" id="num-b" placeholder="b" value="3">
                <button onclick="testAdd()">add(a, b)</button>
            </div>
            <div id="result-add" class="result">Sonuç burada görünecek...</div>
        </div>
        
        <!-- Test 4: Get User -->
        <div class="card">
            <h2>👤 4. Kullanıcı Al (Object Dönüşü)</h2>
            <div class="test-row">
                <input type="number" id="user-id" placeholder="User ID" value="1">
                <button onclick="testGetUser()">getUser(id)</button>
            </div>
            <div id="result-user" class="result">Sonuç burada görünecek...</div>
        </div>
        
        <!-- Test 5: Divide -->
        <div class="card">
            <h2>➗ 5. Bölme (Hata Dönebilir)</h2>
            <div class="test-row">
                <input type="number" id="div-a" placeholder="a" value="10">
                <input type="number" id="div-b" placeholder="b" value="2">
                <button onclick="testDivide()">divide(a, b)</button>
            </div>
            <p style="font-size:12px; color:#aaa; margin-top:5px;">💡 b=0 yaparak hata testi yapabilirsiniz</p>
            <div id="result-divide" class="result">Sonuç burada görünecek...</div>
        </div>
        
        <!-- Test 6: Long Task -->
        <div class="card">
            <h2>⏳ 6. Uzun İşlem (Async Test)</h2>
            <div class="test-row">
                <input type="number" id="task-seconds" placeholder="Saniye" value="2">
                <button id="long-task-btn" onclick="testLongTask()">longTask(seconds)</button>
            </div>
            <div id="result-long" class="result">Sonuç burada görünecek...</div>
        </div>
        
        <!-- Events -->
        <div class="card">
            <h2>📡 Go'dan Gelen Eventler</h2>
            <div id="events-log" class="events-log result">
                Event bekleniyor...
            </div>
        </div>
        
        <div class="footer">
            GOMAD Framework - Phase 2: Bridge Layer Test<br>
            Press F12 for DevTools
        </div>
    </div>
    
    <script>
        // Helper: Sonucu göster
        function showResult(elementId, data, isError = false) {
            const el = document.getElementById(elementId);
            el.className = 'result ' + (isError ? 'error' : 'success');
            el.textContent = typeof data === 'object' ? JSON.stringify(data, null, 2) : data;
        }
        
        function showPending(elementId) {
            const el = document.getElementById(elementId);
            el.className = 'result pending';
            el.textContent = '⏳ İşleniyor...';
        }
        
        // Test 1: Version
        async function testVersion() {
            showPending('result-version');
            try {
                const result = await window.gomad.call('getVersion');
                showResult('result-version', result);
            } catch (e) {
                showResult('result-version', 'Hata: ' + e.message, true);
            }
        }
        
        // Test 2: Greet
        async function testGreet() {
            const name = document.getElementById('name-input').value;
            showPending('result-greet');
            try {
                const result = await window.gomad.call('greet', name);
                showResult('result-greet', result);
            } catch (e) {
                showResult('result-greet', 'Hata: ' + e.message, true);
            }
        }
        
        // Test 3: Add
        async function testAdd() {
            const a = parseInt(document.getElementById('num-a').value);
            const b = parseInt(document.getElementById('num-b').value);
            showPending('result-add');
            try {
                const result = await window.gomad.call('add', a, b);
                showResult('result-add', a + ' + ' + b + ' = ' + result);
            } catch (e) {
                showResult('result-add', 'Hata: ' + e.message, true);
            }
        }
        
        // Test 4: Get User
        async function testGetUser() {
            const id = parseInt(document.getElementById('user-id').value);
            showPending('result-user');
            try {
                const result = await window.gomad.call('getUser', id);
                showResult('result-user', result);
            } catch (e) {
                showResult('result-user', 'Hata: ' + e.message, true);
            }
        }
        
        // Test 5: Divide
        async function testDivide() {
            const a = parseFloat(document.getElementById('div-a').value);
            const b = parseFloat(document.getElementById('div-b').value);
            showPending('result-divide');
            try {
                const result = await window.gomad.call('divide', a, b);
                showResult('result-divide', a + ' / ' + b + ' = ' + result);
            } catch (e) {
                showResult('result-divide', '❌ Hata: ' + e.message, true);
            }
        }
        
        // Test 6: Long Task
        async function testLongTask() {
            const seconds = parseInt(document.getElementById('task-seconds').value);
            const btn = document.getElementById('long-task-btn');
            btn.disabled = true;
            btn.textContent = '⏳ Çalışıyor...';
            showPending('result-long');
            try {
                const result = await window.gomad.call('longTask', seconds);
                showResult('result-long', '✅ ' + result);
            } catch (e) {
                showResult('result-long', 'Hata: ' + e.message, true);
            } finally {
                btn.disabled = false;
                btn.textContent = 'longTask(seconds)';
            }
        }
        
        // Event listener
        if (window.gomad) {
            window.gomad.on('app:notification', (data) => {
                const log = document.getElementById('events-log');
                const time = new Date().toLocaleTimeString();
                log.innerHTML = '<div class="event-item">📨 [' + time + '] ' + JSON.stringify(data) + '</div>' + log.innerHTML;
            });
            
            // Test amaçlı: Her 10 saniyede bir event gönder
            console.log('GOMAD Bridge ready!');
        }
    </script>
</body>
</html>
`

	// HTML'i yükle
	wv.SetHTML(html)

	log.Println("WebView created, starting event loop...")
	log.Println("Press F12 to open DevTools")
	log.Println("")

	// Birkaç saniye sonra test event'i gönder
	go func() {
		time.Sleep(3 * time.Second)
		log.Println("[Go] Sending test event...")
		wv.Emit("app:notification", map[string]interface{}{
			"message": "GOMAD Backend hazır!",
			"time":    time.Now().Format("15:04:05"),
		})
	}()

	// TypeScript çıktısını oluştur
	// Bunu production build'de değil, dev modda çalıştırın.
	if err := bridge.GenerateTSDefinitions("gomad.d.ts"); err != nil {
		log.Printf("Failed to generate TS definitions: %v", err)
	} else {
		log.Println("TypeScript definitions generated: gomad.d.ts ✅")
	}

	// Event loop başlat
	wv.Run()

	log.Println("Application closed")
}
