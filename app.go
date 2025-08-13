package main

import (
	"context"
	"fmt"
)

// App struct
type App struct {
	ctx    context.Context
	config *Config
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		config: GetConfig(),
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
	config["configPath"] = GetConfigPath()
	
	return config
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
