import { useState } from "react";
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

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!targetUrl) {
      alert("请输入目标网站 URL");
      return;
    }

    onSave({
      proxy,
      targetUrl,
    });
  };

  return (
    <div className="flex flex-col h-screen bg-background">
      {/* 设置内容 */}
      <div className="flex-1 flex items-center justify-center p-4 overflow-auto">
        <div className="w-full max-w-md space-y-6 p-6">
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

            <div className="pt-4">
              <Button type="submit" className="w-full">
                保存并启动
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
