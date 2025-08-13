package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// setupWebViewEnvironment sets up WebView2 environment variables before starting the app
func setupWebViewEnvironment(config *Config) {
	// Set WebView2 additional browser arguments for proxy
	if config.ProxyServer != "" {
		args := fmt.Sprintf("--proxy-server=%s", config.ProxyServer)
		os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", args)
	}
	
	// Note: User data directory in Wails v2 is handled automatically
	// The app data is stored in standard locations based on the OS
}

func main() {
	// Load configuration first to set up webview environment
	config := loadConfig()
	
	// Set up WebView2 environment before starting the app
	setupWebViewEnvironment(config)
	
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "miproject",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewUserDataPath: "miproject",
		},
		})

	if err != nil {
		println("Error:", err.Error())
	}
}
