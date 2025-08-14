package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// 全局应用实例，用于菜单回调
var globalApp *App

// createAppMenu 创建应用菜单
func createAppMenu() *menu.Menu {
	appMenu := menu.NewMenu()
	
	// 文件菜单
	fileMenu := appMenu.AddSubmenu("文件")
	fileMenu.AddText("设置", keys.CmdOrCtrl("s"), func(_ *menu.CallbackData) {
		if globalApp != nil {
			globalApp.OpenSettingsWindow()
		}
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("重启应用", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		if globalApp != nil {
			globalApp.RestartApplication()
		}
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("退出", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		if globalApp != nil && globalApp.ctx != nil {
			runtime.Quit(globalApp.ctx)
		}
	})
	
	// 查看菜单
	viewMenu := appMenu.AddSubmenu("查看")
	viewMenu.AddText("刷新", keys.Key("f5"), func(_ *menu.CallbackData) {
		if globalApp != nil && globalApp.ctx != nil {
			runtime.WindowReload(globalApp.ctx)
		}
	})
	viewMenu.AddText("全屏", keys.Key("f11"), func(_ *menu.CallbackData) {
		if globalApp != nil && globalApp.ctx != nil {
			runtime.WindowToggleMaximise(globalApp.ctx)
		}
	})
	
	// 帮助菜单
	helpMenu := appMenu.AddSubmenu("帮助")
	helpMenu.AddText("关于", nil, func(_ *menu.CallbackData) {
		if globalApp != nil && globalApp.ctx != nil {
			runtime.MessageDialog(globalApp.ctx, runtime.MessageDialogOptions{
				Type:    runtime.InfoDialog,
				Title:   "关于 MiProject",
				Message: "MiProject - 基于Wails v2的WebView应用\n\n支持代理、自定义User-Agent等功能",
			})
		}
	})
	
	return appMenu
}

// setupWebViewEnvironment sets up WebView2 environment variables before starting the app
func setupWebViewEnvironment(config *Config) {
	var args []string
	
	// Set WebView2 additional browser arguments for proxy
	if config.ProxyServer != "" {
		args = append(args, fmt.Sprintf("--proxy-server=%s", config.ProxyServer))
		fmt.Printf("Setting WebView2 proxy: %s\n", config.ProxyServer)
	}
	
	// Set User-Agent
	if config.UserAgent != "" {
		args = append(args, fmt.Sprintf("--user-agent=%s", config.UserAgent))
		fmt.Printf("Setting WebView2 user agent: %s\n", config.UserAgent)
	}
	
	// Add anti-detection arguments to bypass Cloudflare and other bot detection
	if config.AntiDetection {
		antiDetectionArgs := []string{
			"--disable-blink-features=AutomationControlled",  // 禁用自动化控制标识
			"--disable-features=VizDisplayCompositor",         // 禁用某些检测特征
			"--disable-dev-shm-usage",                         // 禁用开发共享内存
			"--disable-software-rasterizer",                   // 禁用软件光栅化
			"--disable-background-timer-throttling",           // 禁用后台定时器限制
			"--disable-backgrounding-occluded-windows",        // 禁用被遮挡窗口的后台处理
			"--disable-renderer-backgrounding",                // 禁用渲染器后台处理
			"--disable-field-trial-config",                    // 禁用字段试验配置
			"--disable-back-forward-cache",                    // 禁用前进后退缓存
			"--disable-ipc-flooding-protection",               // 禁用IPC洪水保护
			"--disable-hang-monitor",                          // 禁用挂起监视器
			"--disable-prompt-on-repost",                      // 禁用重新提交提示
			"--disable-default-apps",                          // 禁用默认应用
			"--disable-component-update",                      // 禁用组件更新
			"--disable-background-networking",                 // 禁用后台网络
			"--disable-sync",                                  // 禁用同步
			"--metrics-recording-only",                        // 仅记录指标
			"--no-default-browser-check",                      // 不检查默认浏览器
			"--no-first-run",                                  // 不显示首次运行
			"--password-store=basic",                          // 基本密码存储
			"--use-mock-keychain",                             // 使用模拟密钥链
		}
		
		args = append(args, antiDetectionArgs...)
		fmt.Printf("Anti-detection mode enabled\n")
	}
	
	// Set all browser arguments
	if len(args) > 0 {
		allArgs := fmt.Sprintf("%s", args[0])
		for i := 1; i < len(args); i++ {
			allArgs += " " + args[i]
		}
		os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", allArgs)
		fmt.Printf("Setting WebView2 arguments with anti-detection: %s\n", allArgs)
	}
	
	// Always set the user data folder for WebView2 if specified
	if config.UserData != "" {
		os.Setenv("WEBVIEW2_USER_DATA_FOLDER", config.UserData)
		fmt.Printf("Setting WebView2 user data folder: %s\n", config.UserData)
	}
}

func main() {
	// Load configuration first to set up webview environment
	config := loadConfig()
	
	// Set up WebView2 environment before starting the app
	setupWebViewEnvironment(config)
	
	// Create an instance of the app structure
	app := NewApp()
	
	// Set global app instance for menu callbacks
	globalApp = app
	
	// Create application menu
	appMenu := createAppMenu()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "miproject",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Menu:             appMenu,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewUserDataPath: config.UserData,
		},
		})

	if err != nil {
		println("Error:", err.Error())
	}
}
