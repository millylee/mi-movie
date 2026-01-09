// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod settings;

use settings::{AppSettings, SettingsManager};
use std::sync::Mutex;
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
    println!("Saving settings: {:?}", settings);
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
    WebviewWindowBuilder::new(
        app,
        "settings",
        WebviewUrl::App("index.html".into()),
    )
    .title("MiMovie 设置")
    .inner_size(500.0, 650.0)
    .min_inner_size(400.0, 550.0)
    .resizable(true)
    .center()
    .visible(true)
    .decorations(false)
    .focused(true)
    .build()
}

fn create_main_window(app: &tauri::AppHandle, settings: &AppSettings) -> tauri::Result<tauri::WebviewWindow> {
    println!("Creating main window with URL: {}", settings.target_url);
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
    .inner_size(1400.0, 800.0)
    .min_inner_size(800.0, 600.0)
    .resizable(true)
    .center()
    .visible(true)
    .decorations(true)
    .focused(true);

    // 如果设置了代理，应用代理配置
    if !settings.proxy.is_empty() {
        println!("Applying proxy settings: {}", settings.proxy);
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
    println!(">>> reload_main_window called");
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
        println!("Closing existing main window for reload...");
        // 这次 close 会被 on_window_event 放行，因为检测到了 is_reloading
        if let Err(e) = main_window.close() {
            println!("Error closing main window: {}", e);
        }
        
        // 等待窗口完全销毁
        let mut retries = 0;
        while app.get_webview_window("main").is_some() {
            if retries > 50 { // 5秒超时
                println!("Timeout waiting for window close");
                break;
            }
            std::thread::sleep(std::time::Duration::from_millis(100));
            retries += 1;
        }
    }

    // 重置标志
    if let Ok(mut reloading) = state.is_reloading.lock() {
        *reloading = false;
    }

    // 创建新窗口
    println!("Attempting to create new main window...");
    match create_main_window(&app, &settings) {
        Ok(_) => {
            println!("Main window created successfully");
            Ok(())
        },
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
                    println!("Tray menu event: {}", event.id().as_ref());
                    match event.id().as_ref() {
                        "show" => {
                            if let Some(main_window) = app.get_webview_window("main") {
                                let _ = main_window.show();
                                let _ = main_window.set_focus();
                            } else if let Some(settings_window) = app.get_webview_window("settings") {
                                // 如果没有主窗口，显示设置窗口
                                let _ = settings_window.show();
                                let _ = settings_window.set_focus();
                            }
                        }
                        "settings" => {
                            // 无论是显示现有还是创建新设置窗口
                            // 注意：open_settings 是 async 命令函数，不能在 sync closure 直接调用其逻辑
                            // 但我们可以重用 open_settings 的逻辑，或者通过 app handle 直接操作
                            if let Some(w) = app.get_webview_window("settings") {
                                let _ = w.show();
                                let _ = w.set_focus();
                            } else {
                                let _ = create_settings_window(app);
                            }
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
                    } = event {
                        println!("Tray left click");
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
            let settings = state.settings_manager.lock().unwrap().load().unwrap_or_default();
            println!("Loaded settings on startup: {:?}", settings);
            
            if !settings.target_url.is_empty() {
                println!("Startup: Opening main window");
                // 有配置，启动主窗口
                create_main_window(app.handle(), &settings)?;
            } else {
                println!("Startup: Opening settings window");
                // 无配置，启动设置窗口
                create_settings_window(app.handle())?;
            }

            Ok(())
        })
        .on_window_event(|window, event| {
            if let WindowEvent::CloseRequested { api, .. } = event {
                let label = window.label();
                println!("CloseRequested: {}", label);
                
                // 检查是否正在重载
                let app_handle = window.app_handle();
                let state = app_handle.state::<AppState>();
                let is_reloading = state.is_reloading.lock().map(|b| *b).unwrap_or(false);

                if label == "main" && is_reloading {
                    println!("Allowing close for reload: {}", label);
                    // 如果正在重载，允许关闭
                    return;
                }

                // 默认行为：拦截关闭，改为隐藏
                if label == "main" || label == "settings" {
                    println!("Preventing close, hiding window: {}", label);
                    api.prevent_close();
                    let _ = window.hide();
                }
            }
        })
        .build(tauri::generate_context!())
        .expect("error while building tauri application")
        .run(|_app_handle, _event| {
            // 空闭包，不做额外处理
        });
}
