import { useState } from "react";
import { open } from "@tauri-apps/plugin-dialog";
import { getCurrentWindow } from "@tauri-apps/api/window";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import type { AppSettings } from "../App";

interface SettingsProps {
  initialSettings: AppSettings | null;
  onSave: (settings: AppSettings) => void;
  onCancel: () => void;
}

export default function Settings({ initialSettings, onSave }: SettingsProps) {
  const [proxy, setProxy] = useState(initialSettings?.proxy || "");
  const [targetUrl, setTargetUrl] = useState(initialSettings?.targetUrl || "");
  const [userDataPath, setUserDataPath] = useState(
    initialSettings?.userDataPath || ""
  );

  const handleSelectFolder = async () => {
    try {
      const selected = await open({
        directory: true,
        multiple: false,
        title: "选择用户数据目录",
      });

      if (selected && typeof selected === "string") {
        setUserDataPath(selected);
      }
    } catch (error) {
      console.error("Failed to select folder:", error);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!targetUrl) {
      alert("请输入目标网站 URL");
      return;
    }

    onSave({
      proxy,
      targetUrl,
      userDataPath,
    });
  };

  const handleMinimize = async () => {
    try {
      console.log("handleMinimize called");
      const window = getCurrentWindow();
      console.log("Current window label:", window.label);
      await window.hide();
      console.log("Window hidden successfully");
    } catch (error) {
      console.error("Failed to minimize window:", error);
    }
  };

  const handleClose = async () => {
    try {
      console.log("handleClose called");
      const window = getCurrentWindow();
      console.log("Current window label:", window.label);
      await window.hide();
      console.log("Window hidden successfully");
    } catch (error) {
      console.error("Failed to close window:", error);
    }
  };

  return (
    <div className="flex flex-col h-screen bg-background">
      {/* 设置内容 */}
      <div className="flex-1 flex items-center justify-center p-4 overflow-auto">
        <div className="w-full max-w-md space-y-6 bg-card p-6 rounded-lg border border-border shadow-lg">
          <div className="space-y-2 text-center">
            <h1 className="text-2xl font-bold text-foreground">应用设置</h1>
            <p className="text-sm text-muted-foreground">配置应用程序参数</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="targetUrl">目标网站 URL *</Label>
              <Input
                id="targetUrl"
                type="url"
                placeholder="https://example.com"
                value={targetUrl}
                onChange={(e) => setTargetUrl(e.target.value)}
                required
              />
              <p className="text-xs text-muted-foreground">
                点击保存后将在新窗口中打开此网站
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="proxy">HTTP 代理</Label>
              <Input
                id="proxy"
                type="text"
                placeholder="http://127.0.0.1:7890"
                value={proxy}
                onChange={(e) => setProxy(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">
                格式: http://host:port (可选，WebView 支持代理)
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="userDataPath">用户数据目录</Label>
              <div className="flex gap-2">
                <Input
                  id="userDataPath"
                  type="text"
                  placeholder="留空使用默认路径"
                  value={userDataPath}
                  onChange={(e) => setUserDataPath(e.target.value)}
                />
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleSelectFolder}
                >
                  浏览
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                浏览器用户数据存储路径 (可选，WebView 模式不生效)
              </p>
            </div>

            <div className="pt-4">
              <Button type="submit" className="w-full">
                保存并启动
              </Button>
            </div>

            {initialSettings?.targetUrl && (
              <p className="text-xs text-center text-muted-foreground">
                关闭窗口将最小化到系统托盘
              </p>
            )}
          </form>
        </div>
      </div>
    </div>
  );
}
