import { useState, useEffect } from "react";
import { invoke } from "@tauri-apps/api/core";
import Settings from "./components/Settings";

export interface AppSettings {
  proxy: string;
  targetUrl: string;
  userDataPath: string;
}

function App() {
  const [settings, setSettings] = useState<AppSettings | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    console.log("App mounted, loading settings...");
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      console.log("Calling get_settings...");
      const loadedSettings = await invoke<AppSettings>("get_settings");
      console.log("Settings loaded:", loadedSettings);
      setSettings(loadedSettings);
    } catch (error) {
      console.error("Failed to load settings:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSaveSettings = async (newSettings: AppSettings) => {
    try {
      await invoke("save_settings", { settings: newSettings });
      setSettings(newSettings);

      // 保存后创建主窗口加载目标 URL
      if (newSettings.targetUrl) {
        await invoke("reload_main_window", { settings: newSettings });
      }
    } catch (error) {
      console.error("Failed to save settings:", error);
      alert("保存设置失败: " + error);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen bg-background">
        <div className="text-foreground">加载中...</div>
      </div>
    );
  }

  // 设置页面
  return (
    <Settings
      initialSettings={settings}
      onSave={handleSaveSettings}
      onCancel={() => {}}
    />
  );
}

export default App;
