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
  const [userDataPath, setUserDataPath] = useState(initialSettings?.userDataPath || "");

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
      {/* 自定义标题栏 */}
      <div className="h-8 bg-card border-b border-border flex items-center justify-between px-3 select-none">
        <div data-tauri-drag-region className="flex-1 flex items-center">
          <span className="text-sm font-medium text-foreground">MiMovie 设置</span>
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={handleMinimize}
            className="w-8 h-6 flex items-center justify-center hover:bg-accent rounded transition-colors"
          >
            <svg className="w-3 h-3" viewBox="0 0 12 12" fill="currentColor">
              <rect x="0" y="5" width="12" height="2" />
            </svg>
          </button>
          <button
            onClick={handleClose}
            className="w-8 h-6 flex items-center justify-center hover:bg-destructive hover:text-destructive-foreground rounded transition-colors"
          >
            <svg className="w-3 h-3" viewBox="0 0 12 12" fill="currentColor">
              <path d="M11.25 1.81L10.19 0.75L6 4.94L1.81 0.75L0.75 1.81L4.94 6L0.75 10.19L1.81 11.25L6 7.06L10.19 11.25L11.25 10.19L7.06 6L11.25 1.81Z" />
            </svg>
          </button>
        </div>
      </div>

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
                <Button type="button" variant="outline" onClick={handleSelectFolder}>
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
