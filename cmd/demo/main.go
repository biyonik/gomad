package main

import webview "github.com/webview/webview_go"

func main() {
	w := webview.New(true)
	defer w.Destroy()

	w.SetTitle("GOMAD WebView Testi")
	w.SetSize(800, 600, webview.HintNone)

	w.SetHtml(`
        <!DOCTYPE html>
        <html>
        <head>
            <style>
                body {
                    font-family: Arial, sans-serif;
                    display: flex;
                    justify-content: center;
                    align-items: center;
                    height: 100vh;
                    margin: 0;
                    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                    color: white;
                }
                h1 { font-size: 3em; }
            </style>
        </head>
        <body>
            <h1>🚀 GOMAD WebView!</h1>
        </body>
        </html>
    `)

	w.Run()
}
