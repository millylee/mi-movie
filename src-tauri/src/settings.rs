use serde::{Deserialize, Serialize};
use std::fs;
use std::io;
use std::path::PathBuf;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AppSettings {
    pub proxy: String,
    pub target_url: String,
    pub user_data_path: String,
}

impl Default for AppSettings {
    fn default() -> Self {
        Self {
            proxy: String::new(),
            target_url: String::new(),
            user_data_path: String::new(),
        }
    }
}

pub struct SettingsManager {
    config_path: PathBuf,
}

impl SettingsManager {
    pub fn new() -> Self {
        let config_dir = directories::ProjectDirs::from("", "", "MiMovie")
            .map(|dirs| dirs.config_dir().to_path_buf())
            .unwrap_or_else(|| PathBuf::from("."));

        fs::create_dir_all(&config_dir).ok();

        Self {
            config_path: config_dir.join("settings.json"),
        }
    }

    pub fn load(&self) -> io::Result<AppSettings> {
        if !self.config_path.exists() {
            return Ok(AppSettings::default());
        }

        let content = fs::read_to_string(&self.config_path)?;
        serde_json::from_str(&content).map_err(|e| io::Error::new(io::ErrorKind::InvalidData, e))
    }

    pub fn save(&self, settings: &AppSettings) -> io::Result<()> {
        let content = serde_json::to_string_pretty(settings)
            .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, e))?;
        fs::write(&self.config_path, content)
    }
}
