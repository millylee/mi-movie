package main

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
)

type Config struct {
	ProxyServer string `json:"proxyServer"`
	HomePage    string `json:"homePage"`
	UserData    string `json:"userData"`
}

var appConfig *Config

func init() {
	appConfig = &Config{
		ProxyServer: "",
		HomePage:    "https://www.iyf.tv",
		UserData:    "",
	}
}

func loadConfig() *Config {
	// 解析命令行参数
	var proxyServer = flag.String("proxyServer", "", "Proxy server URL")
	var homePage = flag.String("homePage", "", "Home page URL")
	var userData = flag.String("userData", "", "User data directory")
	flag.Parse()
	
	// 确定用户数据目录
	dataDir := *userData
	if dataDir == "" {
		userConfigDir, err := os.UserConfigDir()
		if err == nil {
			dataDir = filepath.Join(userConfigDir, "miproject")
		} else {
			// 如果获取用户配置目录失败，使用当前目录
			dataDir = "."
		}
	}
	
	// 确保数据目录存在
	os.MkdirAll(dataDir, 0755)
	
	// 配置文件路径
	configPath := filepath.Join(dataDir, "config.json")
	
	// 默认配置
	config := &Config{
		ProxyServer: "",
		HomePage:    "https://www.iyf.tv",
		UserData:    dataDir,
	}
	
	// 尝试从配置文件加载
	configFileExists := false
	if _, err := os.Stat(configPath); err == nil {
		configFileExists = true
		data, err := os.ReadFile(configPath)
		if err == nil {
			var fileConfig Config
			if err := json.Unmarshal(data, &fileConfig); err == nil {
				// 从文件加载配置
				config.ProxyServer = fileConfig.ProxyServer
				config.HomePage = fileConfig.HomePage
				config.UserData = fileConfig.UserData
			}
		}
	}
	
	// 命令行参数优先级最高（如果提供了的话）
	configModified := false
	if *proxyServer != "" {
		config.ProxyServer = *proxyServer
		configModified = true
	}
	if *homePage != "" {
		config.HomePage = *homePage
		configModified = true
	}
	if *userData != "" {
		config.UserData = *userData
		configModified = true
	}
	
	// 确保数据目录存在（最终确认）
	os.MkdirAll(config.UserData, 0755)
	
	// 只有在命令行参数修改了配置或者配置文件不存在时才保存
	if configModified || !configFileExists {
		saveConfig(config)
	}
	
	return config
}

func saveConfig(config *Config) error {
	if config.UserData == "" {
		return nil
	}
	
	// 确保目录存在
	if err := os.MkdirAll(config.UserData, 0755); err != nil {
		return err
	}
	
	configPath := filepath.Join(config.UserData, "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(configPath, data, 0644)
}

func GetConfig() *Config {
	if appConfig == nil {
		appConfig = loadConfig()
	}
	return appConfig
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	config := GetConfig()
	if config.UserData == "" {
		return ""
	}
	return filepath.Join(config.UserData, "config.json")
}