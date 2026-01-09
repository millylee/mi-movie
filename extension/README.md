# 扩展目录

将 Chrome 扩展放置在此目录下，应用启动时会自动加载。

## 目录结构示例

```
extension/
  ├── example-extension/
  │   ├── manifest.json
  │   ├── background.js
  │   ├── content.js
  │   └── icons/
  │       └── icon.png
```

## 注意事项

- 每个扩展必须包含有效的 `manifest.json` 文件
- 支持 Chrome Extension Manifest V2 和 V3
- 扩展需要是解压后的文件夹格式，不支持 .crx 文件
