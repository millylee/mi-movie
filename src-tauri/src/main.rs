// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod settings;

use settings::{AppSettings, SettingsManager};
use std::{
    fs,
    io::Write,
    path::{Path, PathBuf},
    sync::Mutex,
};
use tauri::{
    menu::{Menu, MenuItem},
    tray::{MouseButton, MouseButtonState, TrayIcon, TrayIconBuilder, TrayIconEvent},
    Manager, RunEvent, State, WebviewUrl, WebviewWindowBuilder, WindowEvent,
};

// 移除旧的 AppState 定义，直接使用 AppStateV2 并重命名为 AppState
struct AppState {
    settings_manager: Mutex<SettingsManager>,
    tray_icon: Mutex<Option<TrayIcon>>,
    is_reloading: Mutex<bool>,
}

#[tauri::command]
fn get_settings(state: State<AppState>) -> Result<AppSettings, String> {
    let manager = state.settings_manager.lock().map_err(|e| e.to_string())?;
    manager.load().map_err(|e| e.to_string())
}

#[tauri::command]
fn save_settings(settings: AppSettings, state: State<AppState>) -> Result<(), String> {
    let manager = state.settings_manager.lock().map_err(|e| e.to_string())?;
    manager.save(&settings).map_err(|e| e.to_string())
}

#[tauri::command]
async fn open_settings(app: tauri::AppHandle) -> Result<(), String> {
    // 如果设置窗口已存在，直接显示
    if let Some(settings_window) = app.get_webview_window("settings") {
        settings_window.show().map_err(|e| e.to_string())?;
        settings_window.set_focus().map_err(|e| e.to_string())?;
        return Ok(());
    }

    // 创建设置窗口
    create_settings_window(&app).map_err(|e| e.to_string())?;

    Ok(())
}

fn create_settings_window(app: &tauri::AppHandle) -> tauri::Result<tauri::WebviewWindow> {
    if let Some(window) = app.get_webview_window("settings") {
        return Ok(window);
    }
    WebviewWindowBuilder::new(app, "settings", WebviewUrl::App("index.html".into()))
        .title("MiMovie 设置")
        .inner_size(500.0, 650.0)
        .min_inner_size(400.0, 550.0)
        .resizable(true)
        .center()
        .visible(true)
        .decorations(true)
        .focused(true)
        .build()
}

fn app_data_dir() -> PathBuf {
    directories::ProjectDirs::from("", "", "MiMovie")
        .map(|dirs| {
            let data_dir = dirs.data_dir().to_path_buf();
            data_dir.parent().unwrap_or(&data_dir).to_path_buf()
        })
        .unwrap_or_else(|| PathBuf::from("."))
}

fn extensions_root() -> PathBuf {
    app_data_dir().join("extension")
}

fn extensions_log_path() -> PathBuf {
    app_data_dir().join("log").join("extension.log")
}

fn log_to_file(message: &str) {
    let log_path = extensions_log_path();
    if let Some(parent) = log_path.parent() {
        if let Err(e) = fs::create_dir_all(parent) {
            println!("Failed to create log directory: {}", e);
            return;
        }
    }
    let mut file = match fs::OpenOptions::new()
        .create(true)
        .append(true)
        .open(&log_path)
    {
        Ok(file) => file,
        Err(e) => {
            println!("Failed to open log file: {}", e);
            return;
        }
    };
    let _ = writeln!(file, "{}", message);
}

fn discover_extension_dirs(root: &Path) -> Vec<PathBuf> {
    let mut result = Vec::new();
    if !root.exists() {
        log_to_file(&format!(
            "Extension root does not exist: {}",
            root.to_string_lossy()
        ));
        return result;
    }
    let entries = match fs::read_dir(root) {
        Ok(entries) => entries,
        Err(e) => {
            log_to_file(&format!(
                "Failed to read extension root {}: {}",
                root.to_string_lossy(),
                e
            ));
            return result;
        }
    };
    for entry in entries.flatten() {
        let path = entry.path();
        if !path.is_dir() {
            continue;
        }
        if path.join("manifest.json").is_file() {
            result.push(path);
        }
    }
    log_to_file(&format!(
        "Discovered {} extension(s) under {}",
        result.len(),
        root.to_string_lossy()
    ));
    result
}

fn format_extension_arg(paths: &[PathBuf]) -> String {
    paths
        .iter()
        .map(|p| format!("\"{}\"", p.to_string_lossy()))
        .collect::<Vec<_>>()
        .join(",")
}

fn resolve_webview_data_dir(settings: &AppSettings) -> PathBuf {
    if !settings.user_data_path.is_empty() {
        PathBuf::from(&settings.user_data_path)
    } else {
        app_data_dir().join("webview")
    }
}

fn create_main_window(
    app: &tauri::AppHandle,
    settings: &AppSettings,
) -> tauri::Result<tauri::WebviewWindow> {
    let webview_data_dir = resolve_webview_data_dir(settings);
    if let Err(e) = fs::create_dir_all(&webview_data_dir) {
        println!("Failed to create webview data directory: {}", e);
    }

    let extension_dirs = discover_extension_dirs(&extensions_root());

    if extension_dirs.is_empty() {
        log_to_file("No extensions found to load.");
    } else {
        let list = extension_dirs
            .iter()
            .map(|p| p.to_string_lossy().to_string())
            .collect::<Vec<_>>()
            .join(", ");
        log_to_file(&format!("Extensions to load: {}", list));
    }

    let mut builder = WebviewWindowBuilder::new(
        app,
        "main",
        WebviewUrl::External(
            settings
                .target_url
                .parse::<url::Url>()
                .map_err(|e| tauri::Error::AssetNotFound(format!("Invalid URL: {}", e)))?,
        ),
    )
    .title("MiMovie")
    .inner_size(1280.0, 800.0)
    .min_inner_size(800.0, 600.0)
    .resizable(true)
    .center()
    .visible(true)
    .decorations(true)
    .focused(true);

    builder = builder.data_directory(webview_data_dir);

    #[cfg(target_os = "windows")]
    if !extension_dirs.is_empty() {
        builder = builder.browser_extensions_enabled(true);
        let extensions_arg = format_extension_arg(&extension_dirs);
        let args = format!(
            "--disable-extensions-except={} --load-extension={}",
            extensions_arg, extensions_arg
        );
        log_to_file(&format!("WebView2 extension args: {}", args));
        builder = builder.additional_browser_args(&args);
    }

    // 如果设置了代理，应用代理配置
    if !settings.proxy.is_empty() {
        builder = builder.proxy_url(
            settings
                .proxy
                .parse::<url::Url>()
                .map_err(|e| tauri::Error::AssetNotFound(format!("Invalid proxy URL: {}", e)))?,
        );
    }

    builder.build()
}

#[tauri::command(rename_all = "snake_case", name = "reload_main_window")]
async fn reload_main_window(settings: AppSettings, app: tauri::AppHandle) -> Result<(), String> {
    let state = app.state::<AppState>();

    // 0. 先隐藏设置窗口，避免视觉干扰
    // 保存设置窗口的引用，以便出错时恢复显示
    let settings_window = app.get_webview_window("settings");
    if let Some(ref win) = settings_window {
        let _ = win.hide();
    }

    // 设置正在重载标志
    if let Ok(mut reloading) = state.is_reloading.lock() {
        *reloading = true;
    }

    if let Some(main_window) = app.get_webview_window("main") {
        // 这次 close 会被 on_window_event 放行，因为检测到了 is_reloading
        if let Err(e) = main_window.close() {
            println!("Error closing main window: {}", e);
        }

        // 等待窗口完全销毁
        let mut retries = 0;
        while app.get_webview_window("main").is_some() {
            if retries > 50 {
                // 5秒超时
                break;
            }
            std::thread::sleep(std::time::Duration::from_millis(100));
            retries += 1;
        }
        // 确保 Webview2 进程完全释放资源
        std::thread::sleep(std::time::Duration::from_millis(500));
    }

    // 重置标志
    if let Ok(mut reloading) = state.is_reloading.lock() {
        *reloading = false;
    }

    // 创建新窗口
    match create_main_window(&app, &settings) {
        Ok(_) => Ok(()),
        Err(e) => {
            println!("Failed to create main window: {}", e);
            // 失败时恢复显示设置窗口
            if let Some(win) = settings_window {
                let _ = win.show();
                let _ = win.set_focus();
            }
            Err(e.to_string())
        }
    }
}

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_dialog::init())
        .manage(AppState {
            settings_manager: Mutex::new(SettingsManager::new()),
            tray_icon: Mutex::new(None),
            is_reloading: Mutex::new(false),
        })
        .invoke_handler(tauri::generate_handler![
            get_settings,
            save_settings,
            open_settings,
            reload_main_window
        ])
        .setup(|app| {
            // 托盘菜单
            let show_item = MenuItem::with_id(app, "show", "显示主窗口", true, None::<&str>)?;
            let settings_item = MenuItem::with_id(app, "settings", "设置", true, None::<&str>)?;
            let quit_item = MenuItem::with_id(app, "quit", "退出", true, None::<&str>)?;
            let menu = Menu::with_items(app, &[&show_item, &settings_item, &quit_item])?;

            // 托盘图标
            let tray = TrayIconBuilder::new()
                .icon(app.default_window_icon().unwrap().clone())
                .menu(&menu)
                .on_menu_event(|app, event| {
                    match event.id().as_ref() {
                        "show" => {
                            if let Some(main_window) = app.get_webview_window("main") {
                                let _ = main_window.show();
                                let _ = main_window.set_focus();
                            } else if let Some(settings_window) = app.get_webview_window("settings")
                            {
                                // 如果没有主窗口，显示设置窗口
                                let _ = settings_window.show();
                                let _ = settings_window.set_focus();
                            }
                        }
                        "settings" => {
                            let app = app.clone();
                            tauri::async_runtime::spawn(async move {
                                let state = app.state::<AppState>();

                                // 1. 标记正在切换窗口（允许关闭）
                                if let Ok(mut reloading) = state.is_reloading.lock() {
                                    *reloading = true;
                                }

                                // 2. 如果有主窗口，关闭它（而不是隐藏）
                                if let Some(main_window) = app.get_webview_window("main") {
                                    if let Err(e) = main_window.close() {
                                        println!("Error closing main window: {}", e);
                                    }

                                    // 等待关闭
                                    let mut retries = 0;
                                    while app.get_webview_window("main").is_some() {
                                        if retries > 50 {
                                            break;
                                        }
                                        std::thread::sleep(std::time::Duration::from_millis(100));
                                        retries += 1;
                                    }

                                    // 关键：增加缓冲时间，确保 Webview2 进程完全释放资源（User Data 锁）
                                    std::thread::sleep(std::time::Duration::from_millis(500));
                                }

                                // 3. 标记结束切换（恢复拦截）
                                if let Ok(mut reloading) = state.is_reloading.lock() {
                                    *reloading = false;
                                }

                                // 4. 显示或创建设置窗口
                                if let Some(w) = app.get_webview_window("settings") {
                                    let _ = w.show();
                                    let _ = w.set_focus();
                                } else {
                                    if let Err(e) = create_settings_window(&app) {
                                        println!("Failed to create settings window: {}", e);
                                    }
                                }
                            });
                        }
                        "quit" => app.exit(0),
                        _ => {}
                    }
                })
                .on_tray_icon_event(|tray, event| {
                    if let TrayIconEvent::Click {
                        button: MouseButton::Left,
                        button_state: MouseButtonState::Up,
                        ..
                    } = event
                    {
                        let app = tray.app_handle();
                        // 优先显示主窗口
                        if let Some(main_window) = app.get_webview_window("main") {
                            let _ = main_window.show();
                            let _ = main_window.set_focus();
                        } else {
                            // 否则显示设置窗口
                            if let Some(settings_window) = app.get_webview_window("settings") {
                                let _ = settings_window.show();
                                let _ = settings_window.set_focus();
                            } else {
                                // 都没有，创建设置窗口
                                let _ = create_settings_window(app);
                            }
                        }
                    }
                })
                .build(app)?;

            // 保存托盘实例
            let state = app.state::<AppState>();
            *state.tray_icon.lock().unwrap() = Some(tray);

            // 启动逻辑
            let settings = state
                .settings_manager
                .lock()
                .unwrap()
                .load()
                .unwrap_or_default();

            if !settings.target_url.is_empty() {
                // 有配置，启动主窗口
                create_main_window(app.handle(), &settings)?;
            } else {
                // 无配置，启动设置窗口
                create_settings_window(app.handle())?;
            }

            Ok(())
        })
        .on_window_event(|window, event| {
            if let WindowEvent::CloseRequested { api, .. } = event {
                let label = window.label();

                // 检查是否正在重载
                let app_handle = window.app_handle();
                let state = app_handle.state::<AppState>();
                let is_reloading = state.is_reloading.lock().map(|b| *b).unwrap_or(false);

                if label == "main" && is_reloading {
                    // 如果正在重载，允许关闭
                    return;
                }

                // 默认行为：拦截关闭，改为隐藏
                if label == "main" || label == "settings" {
                    api.prevent_close();
                    let _ = window.hide();
                }
            }
        })
        .build(tauri::generate_context!())
        .expect("error while building tauri application")
        .run(|app_handle, event| {
            if let RunEvent::ExitRequested { api, .. } = event {
                let state = app_handle.state::<AppState>();
                let is_reloading = state.is_reloading.lock().map(|b| *b).unwrap_or(false);

                if is_reloading {
                    api.prevent_exit();
                }
            }
        });
}
