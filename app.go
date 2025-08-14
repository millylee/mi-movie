package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	config *Config
}

// NewApp creates a new App application struct
func NewApp() *App {
	// 强制重新加载配置以确保获取最新的配置
	config := ForceReloadConfig()
	fmt.Printf("NewApp - Loaded config: ProxyServer=%s, HomePage=%s, UserData=%s\n", 
		config.ProxyServer, config.HomePage, config.UserData)
	return &App{
		config: config,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetProxyServer returns the configured proxy server
func (a *App) GetProxyServer() string {
	return a.config.ProxyServer
}

// GetHomePage returns the configured home page
func (a *App) GetHomePage() string {
	return a.config.HomePage
}

// GetUserData returns the configured user data directory
func (a *App) GetUserData() string {
	return a.config.UserData
}

// GetUserAgent returns the configured user agent
func (a *App) GetUserAgent() string {
	return a.config.UserAgent
}

// SetProxyServer sets the proxy server configuration
func (a *App) SetProxyServer(proxy string) error {
	a.config.ProxyServer = proxy
	return saveConfig(a.config)
}

// SetHomePage sets the home page configuration
func (a *App) SetHomePage(homepage string) error {
	a.config.HomePage = homepage
	return saveConfig(a.config)
}

// SetUserData sets the user data directory configuration
func (a *App) SetUserData(userData string) error {
	a.config.UserData = userData
	return saveConfig(a.config)
}

// SetUserAgent sets the user agent configuration
func (a *App) SetUserAgent(userAgent string) error {
	a.config.UserAgent = userAgent
	return saveConfig(a.config)
}

// GetAllConfig returns all configuration
func (a *App) GetAllConfig() *Config {
	return a.config
}

// GetWebViewConfig returns webview-specific configuration for debugging
func (a *App) GetWebViewConfig() map[string]string {
	config := make(map[string]string)
	
	if a.config.ProxyServer != "" {
		config["proxyServer"] = a.config.ProxyServer
	}
	
	if a.config.UserData != "" {
		config["userData"] = a.config.UserData
	}
	
	if a.config.HomePage != "" {
		config["homePage"] = a.config.HomePage
	}
	
	return config
}

// GetConfigInfo returns detailed configuration information including file path
func (a *App) GetConfigInfo() map[string]string {
	config := make(map[string]string)
	
	config["proxyServer"] = a.config.ProxyServer
	config["homePage"] = a.config.HomePage
	config["userData"] = a.config.UserData
	config["userAgent"] = a.config.UserAgent
	config["configPath"] = GetConfigPath()
	
	fmt.Printf("GetConfigInfo called - returning: %+v\n", config)
	return config
}

// ReloadConfig reloads configuration from file and updates the app config
func (a *App) ReloadConfig() error {
	newConfig := loadConfig()
	a.config = newConfig
	
	// Re-setup WebView environment if proxy changed
	setupWebViewEnvironment(a.config)
	
	return nil
}

// NavigateToHome navigates to the configured home page
func (a *App) NavigateToHome() error {
	if a.config.HomePage == "" {
		return fmt.Errorf("home page not configured")
	}
	
	// 在Wails v2中，我们使用runtime的方法来导航
	// 这里我们返回URL，让前端处理导航
	return nil
}

// OpenURLInBrowser opens a URL in the system default browser
func (a *App) OpenURLInBrowser(url string) error {
	if a.ctx == nil {
		return fmt.Errorf("context not available")
	}
	
	// Open URL in system default browser
	runtime.BrowserOpenURL(a.ctx, url)
	
	return nil
}

// OpenConfigDialog opens the configuration dialog
func (a *App) OpenConfigDialog() error {
	if a.ctx == nil {
		return fmt.Errorf("context not available")
	}
	
	// Open a dialog using Wails runtime
	return nil
}

// UpdateConfig updates the configuration with new values
func (a *App) UpdateConfig(newConfig map[string]string) error {
	updated := false
	
	if newConfig["proxyServer"] != a.config.ProxyServer {
		a.config.ProxyServer = newConfig["proxyServer"]
		updated = true
	}
	
	if newConfig["homePage"] != a.config.HomePage {
		a.config.HomePage = newConfig["homePage"]
		updated = true
	}
	
	if newConfig["userData"] != a.config.UserData {
		a.config.UserData = newConfig["userData"]
		updated = true
	}
	
	if updated {
		err := saveConfig(a.config)
		if err != nil {
			return err
		}
		
		// Re-setup WebView environment
		setupWebViewEnvironment(a.config)
		
		fmt.Printf("Configuration updated: %+v\n", a.config)
	}
	
	return nil
}

// OpenSettingsWindow opens settings dialog within the app
func (a *App) OpenSettingsWindow() error {
	if a.ctx == nil {
		return fmt.Errorf("context not available")
	}
	
	// 通过runtime显示自定义设置对话框
	// 我们创建一个简单的输入对话框序列来获取设置
	
	// 获取代理设置
	proxyResult, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   "代理设置",
		Message: fmt.Sprintf("当前代理: %s\n\n是否要修改代理设置？", a.config.ProxyServer),
		Buttons: []string{"修改", "跳过"},
	})
	if err != nil {
		return err
	}
	
	newProxy := a.config.ProxyServer
	if proxyResult == "修改" {
		// 使用简单的输入对话框
		newProxy, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.InfoDialog,
			Title:   "输入代理地址",
			Message: "请在弹出的窗口中输入新的代理地址\n\n当前: " + a.config.ProxyServer + "\n\n例如: http://127.0.0.1:7890",
		})
		// 这里我们需要一个更好的输入方式，暂时使用消息框
	}
	
	// 获取主页设置
	homeResult, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   "主页设置",
		Message: fmt.Sprintf("当前主页: %s\n\n是否要修改主页？", a.config.HomePage),
		Buttons: []string{"修改", "跳过"},
	})
	if err != nil {
		return err
	}
	
	newHomePage := a.config.HomePage
	if homeResult == "修改" {
		// 简单的主页选择
		pageChoice, _ := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.QuestionDialog,
			Title:   "选择主页",
			Message: "请选择要设置的主页：",
			Buttons: []string{"https://www.iyf.tv", "https://www.youtube.com", "https://www.google.com", "自定义"},
		})
		
		switch pageChoice {
		case "https://www.iyf.tv":
			newHomePage = "https://www.iyf.tv"
		case "https://www.youtube.com":
			newHomePage = "https://www.youtube.com"
		case "https://www.google.com":
			newHomePage = "https://www.google.com"
		case "自定义":
			runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
				Type:    runtime.InfoDialog,
				Title:   "自定义主页",
				Message: "自定义主页功能需要在配置文件中手动设置\n\n配置文件位置: " + GetConfigPath(),
			})
		}
	}
	
	// 反检测设置
	antiDetectionResult, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.QuestionDialog,
		Title:   "反检测设置",
		Message: fmt.Sprintf("当前反检测模式: %s\n\n是否要切换反检测模式？", 
			func() string { if a.config.AntiDetection { return "已启用" } else { return "已禁用" } }()),
		Buttons: []string{"切换", "保持"},
	})
	if err != nil {
		return err
	}
	
	newAntiDetection := a.config.AntiDetection
	if antiDetectionResult == "切换" {
		newAntiDetection = !a.config.AntiDetection
	}
	
	// 应用设置
	changed := false
	if newProxy != a.config.ProxyServer {
		a.config.ProxyServer = newProxy
		changed = true
	}
	if newHomePage != a.config.HomePage {
		a.config.HomePage = newHomePage
		changed = true
	}
	if newAntiDetection != a.config.AntiDetection {
		a.config.AntiDetection = newAntiDetection
		changed = true
	}
	
	if changed {
		// 保存配置
		err = saveConfig(a.config)
		if err != nil {
			runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
				Type:    runtime.ErrorDialog,
				Title:   "保存失败",
				Message: "保存配置失败: " + err.Error(),
			})
			return err
		}
		
		// 询问是否重启
		restartResult, _ := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.QuestionDialog,
			Title:   "设置已保存",
			Message: "配置已保存成功！\n\n是否立即重启应用以使配置生效？",
			Buttons: []string{"立即重启", "稍后手动重启"},
		})
		
		if restartResult == "立即重启" {
			return a.RestartApplication()
		}
	} else {
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.InfoDialog,
			Title:   "设置完成",
			Message: "没有修改任何设置。",
		})
	}
	
	return nil
}

// RestartApplication restarts the current application
func (a *App) RestartApplication() error {
	if a.ctx == nil {
		return fmt.Errorf("context not available")
	}
	
	// 显示重启确认对话框
	result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "重启应用",
		Message:       "确定要重启应用吗？\n\n注意：重启后配置更改才会生效。",
		Buttons:       []string{"重启", "取消"},
		DefaultButton: "重启",
	})
	
	if err != nil {
		return err
	}
	
	if result == "重启" {
		// 获取当前可执行文件路径
		executable, err := os.Executable()
		if err != nil {
			runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
				Type:    runtime.ErrorDialog,
				Title:   "重启失败",
				Message: "无法获取可执行文件路径: " + err.Error(),
			})
			return err
		}
		
		// 获取当前工作目录
		workDir, err := os.Getwd()
		if err != nil {
			workDir = filepath.Dir(executable)
		}
		
		// 启动新进程
		cmd := exec.Command(executable)
		cmd.Dir = workDir
		cmd.Env = os.Environ() // 继承环境变量
		
		// 在Windows上，使用CREATE_NEW_PROCESS_GROUP避免继承控制台
		err = cmd.Start()
		if err != nil {
			runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
				Type:    runtime.ErrorDialog,
				Title:   "重启失败",
				Message: "无法启动新进程: " + err.Error() + "\n\n请手动重启应用。",
			})
			return err
		}
		
		// 显示重启提示
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.InfoDialog,
			Title:   "重启中",
			Message: "新进程已启动，当前应用即将关闭...",
		})
		
		// 稍微延迟后退出当前应用，确保新进程已启动
		go func() {
			// 等待一秒确保新进程启动
			// time.Sleep(1 * time.Second)
			runtime.Quit(a.ctx)
		}()
	}
	
	return nil
}
