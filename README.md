# MiMovie

一个基于 Tauri 2 的跨平台应用，在应用内嵌入 WebView 显示网站，支持代理设置和最小化到托盘。

## 功能特性

- ✅ **内嵌 WebView**：在应用窗口内显示目标网站
- ✅ **自定义 URL**：设置启动时默认打开的网站
- ✅ **设置管理**：保存和加载应用配置
- ✅ **跨平台**：支持 Windows、macOS 和 Linux
- ✅ **系统托盘**：支持最小化到托盘，托盘菜单控制

## 扩展自动加载（Windows）
应用启动时会自动扫描数据目录下的扩展并加载（仅 Windows / WebView2 支持）。

**扩展目录：**
- `%APPDATA%\MiMovie\extension`

**目录结构示例：**
```
extension/
  ├── example-extension/
  │   ├── manifest.json
  │   ├── background.js
  │   ├── content.js
  │   └── icons/
  │       └── icon.png
```

**日志：**
- 扩展加载日志保存在 `%APPDATA%\MiMovie\log\extension.log`

> 说明：仅支持解压后的扩展目录，不支持 `.crx`。macOS/Linux 底层 WebView 不支持 Chrome 扩展。

## 需求与行为规约 (Design Specification)

为了确保应用行为的一致性和稳定性，以下是应用的核心行为定义。后续的重构和修复将严格遵循此规约。

### 1. 窗口管理 (Window Management)

应用主要包含两个类型的窗口：

1.  **设置窗口 (Settings Window)**:
    - **Label**: `settings`
    - **用途**: 用于首次配置或修改 URL、代理等参数。
    - **行为**: 只有在未配置 URL 或用户主动通过托盘菜单请求时显示。
2.  **主窗口 (Main Window)**:
    - **Label**: `main`
    - **用途**: 加载用户配置的目标 URL。
    - **行为**: 启动时如果已有配置，自动创建并显示。

### 2. 启动流程 (Startup Flow)

1.  应用启动。
2.  初始化系统托盘。
3.  读取本地配置文件 (`settings.json`)。
4.  **判断逻辑**:
    - **情况 A (无配置)**: 如果 `target_url` 为空 -> 创建并显示 **设置窗口**。
    - **情况 B (有配置)**: 如果 `target_url` 存在 -> 创建并显示 **主窗口**，加载该 URL。

### 3. 关闭行为 (Closing Behavior)

- **点击窗口关闭按钮 (X)**:
  - **不退出程序**。
  - **隐藏当前窗口** (无论是设置窗口还是主窗口)。
  - 程序继续在后台运行，可以通过系统托盘再次唤醒。
- **退出程序**:
  - 只能通过系统托盘右键菜单中的 "退出" 选项来彻底终结程序进程。

### 4. 托盘交互 (System Tray Interaction)

- **左键点击 (Left Click)**:
  - 如果主窗口存在 -> 显示并聚焦主窗口。
  - 如果主窗口不存在（未配置）且设置窗口存在 -> 显示并聚焦设置窗口。
  - 如果窗口已显示 -> (可选) 将其置顶或什么都不做。
- **右键点击 (Right Click)**:
  - 弹出原生系统菜单。
  - **菜单项**:
    1.  **显示主窗口**: 显示主窗口（如果没有配置则提示或显示设置）。
    2.  **设置**: 打开设置窗口。
    3.  **退出**: 彻底关闭应用。

### 5. 设置变更流程 (Settings Update Flow)

当用户在设置界面点击 "保存并启动"：

1.  保存配置到文件。
2.  **强制关闭**当前已存在的 **主窗口** (如果存在)。
    - _注意_：这里必须彻底销毁旧窗口，以便重新创建并应用新的 WebView 配置（如代理）。
3.  等待旧窗口完全关闭。
4.  使用新配置创建并显示新的 **主窗口**。
5.  隐藏 **设置窗口**。

---

## 依赖与开发

### 安装依赖

```bash
pnpm install
```

### 开发

```bash
pnpm tauri dev
```

### 构建

```bash
pnpm tauri build
```

构建完成后，可执行文件将位于 `src-tauri/target/release` 目录。

## 项目结构

```
mi-movie/
├── src/                    # React 前端源码
│   ├── components/        # UI 组件
│   │   ├── ui/           # shadcn/ui 基础组件
│   │   ├── Settings.tsx  # 设置页面
│   │   └── WebViewContainer.tsx
│   ├── lib/              # 工具函数
│   ├── App.tsx           # 主应用组件
│   └── main.tsx          # 入口文件
├── src-tauri/             # Tauri 后端 (Rust)
│   ├── src/
│   │   ├── main.rs       # 主程序入口
│   │   ├── settings.rs   # 设置管理
│   │   └── webview.rs    # 浏览器启动逻辑 (如存在)
│   ├── icons/            # 应用图标
│   ├── Cargo.toml        # Rust 依赖配置
│   └── tauri.conf.json   # Tauri 配置
├── extension/             # Chrome 扩展目录（开发时）
└── icon/                  # 原始图标资源
```
