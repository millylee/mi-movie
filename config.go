package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ProxyServer       string `json:"proxyServer"`
	HomePage          string `json:"homePage"`
	UserData          string `json:"userData"`
	UserAgent         string `json:"userAgent"`
	AntiDetection     bool   `json:"antiDetection"`
}

var appConfig *Config

func init() {
	appConfig = &Config{
		ProxyServer:   "",
		HomePage:      "https://www.iyf.tv",
		UserData:      "",
		UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36 Edg/139.0.0.0",
		AntiDetection: true, // 默认启用反检测
	}
}

var (
	flagsParsed      = false
	proxyServerFlag  *string
	homePageFlag     *string
	userDataFlag     *string
	userAgentFlag    *string
	antiDetectionFlag *bool
)

func initFlags() {
	if !flagsParsed {
		proxyServerFlag = flag.String("proxyServer", "", "Proxy server URL")
		homePageFlag = flag.String("homePage", "", "Home page URL")
		userDataFlag = flag.String("userData", "", "User data directory")
		userAgentFlag = flag.String("userAgent", "", "User agent string")
		antiDetectionFlag = flag.Bool("antiDetection", true, "Enable anti-detection mode")
		flag.Parse()
		flagsParsed = true
	}
}

func loadConfig() *Config {
	// 初始化命令行参数（只执行一次）
	initFlags()
	
	// 确定用户数据目录
	dataDir := *userDataFlag
	if dataDir == "" {
		// 首先尝试使用可执行文件同目录下的 miproject 文件夹
		exePath, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exePath)
			dataDir = filepath.Join(exeDir, "miproject")
		} else {
			// 如果获取可执行文件路径失败，使用系统用户配置目录
			userConfigDir, err := os.UserConfigDir()
			if err == nil {
				dataDir = filepath.Join(userConfigDir, "miproject")
			} else {
				// 最后兜底，使用当前目录
				dataDir = "."
			}
		}
	}
	
	// 确保数据目录存在
	os.MkdirAll(dataDir, 0755)
	
	// 配置文件路径
	configPath := filepath.Join(dataDir, "config.json")
	
	// 调试输出
	fmt.Printf("Configuration directory: %s\n", dataDir)
	fmt.Printf("Configuration file path: %s\n", configPath)
	
	// 默认配置
	config := &Config{
		ProxyServer:   "",
		HomePage:      "https://www.iyf.tv",
		UserData:      dataDir,
		UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36 Edg/139.0.0.0",
		AntiDetection: true,
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
				if fileConfig.UserAgent != "" {
					config.UserAgent = fileConfig.UserAgent
				}
				fmt.Printf("Loaded configuration from: %s\n", configPath)
			}
		}
	} else {
		// 如果新位置没有配置文件，尝试从旧位置（系统配置目录）迁移
		if userConfigDir, err := os.UserConfigDir(); err == nil {
			oldConfigPath := filepath.Join(userConfigDir, "miproject", "config.json")
			if _, err := os.Stat(oldConfigPath); err == nil {
				data, err := os.ReadFile(oldConfigPath)
				if err == nil {
					var fileConfig Config
					if err := json.Unmarshal(data, &fileConfig); err == nil {
						config.ProxyServer = fileConfig.ProxyServer
						config.HomePage = fileConfig.HomePage
						config.UserData = fileConfig.UserData
						if fileConfig.UserAgent != "" {
							config.UserAgent = fileConfig.UserAgent
						}
						configFileExists = true
						fmt.Printf("Migrated configuration from: %s to %s\n", oldConfigPath, configPath)
					}
				}
			}
		}
	}
	
	// 命令行参数优先级最高（如果提供了的话）
	configModified := false
	if *proxyServerFlag != "" {
		config.ProxyServer = *proxyServerFlag
		configModified = true
	}
	if *homePageFlag != "" {
		config.HomePage = *homePageFlag
		configModified = true
	}
	if *userDataFlag != "" {
		config.UserData = *userDataFlag
		configModified = true
	}
	if *userAgentFlag != "" {
		config.UserAgent = *userAgentFlag
		configModified = true
	}
	if antiDetectionFlag != nil {
		config.AntiDetection = *antiDetectionFlag
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

// ForceReloadConfig forces a reload of the configuration
func ForceReloadConfig() *Config {
	appConfig = loadConfig()
	return appConfig
}