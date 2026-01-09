import { useEffect, useState, useRef } from "react";
import { invoke } from "@tauri-apps/api/core";
import { Button } from "./ui/button";
import type { AppSettings } from "../App";

interface WebViewContainerProps {
  settings: AppSettings;
  onOpenSettings: () => void;
}

export default function WebViewContainer({ settings, onOpenSettings }: WebViewContainerProps) {
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const hasLaunchedRef = useRef(false);

  useEffect(() => {
    // 防止重复调用
    if (hasLaunchedRef.current) {
      return;
    }

    const initializeWebView = async () => {
      try {
        setIsLoading(true);
        setError(null);

        // 调用后端命令启动 WebView 窗口
        await invoke("launch_webview", { settings });
        hasLaunchedRef.current = true;
        setIsLoading(false);
      } catch (err) {
        console.error("Failed to launch webview:", err);
        setError(String(err));
        setIsLoading(false);
        hasLaunchedRef.current = false;
      }
    };

    initializeWebView();
  }, []); // 空依赖数组，只在组件挂载时执行一次

  const handleRetry = async () => {
    hasLaunchedRef.current = false;
    setIsLoading(true);
    setError(null);

    try {
      await invoke("launch_webview", { settings });
      hasLaunchedRef.current = true;
      setIsLoading(false);
    } catch (err) {
      console.error("Failed to launch webview:", err);
      setError(String(err));
      setIsLoading(false);
    }
  };

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-screen bg-background p-4">
        <div className="max-w-md space-y-4 text-center">
          <h2 className="text-2xl font-bold text-destructive">启动失败</h2>
          <p className="text-muted-foreground">{error}</p>
          <div className="flex gap-2 justify-center">
            <Button onClick={handleRetry}>重试</Button>
            <Button variant="outline" onClick={onOpenSettings}>
              打开设置
            </Button>
          </div>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center h-screen bg-background">
        <div className="space-y-4 text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="text-foreground">正在打开网页...</p>
          <p className="text-sm text-muted-foreground">{settings.targetUrl}</p>
          <p className="text-xs text-muted-foreground mt-2">
            新窗口将在几秒钟内打开
          </p>
        </div>
      </div>
    );
  }

  // WebView 窗口已启动，主窗口将被隐藏
  return null;
}
