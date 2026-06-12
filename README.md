# MangaMetaManager (MMM)

MangaMetaManager is a modern, local manga metadata management tool inspired by tinyMediaManager. It helps you organize your manga collection, fetch metadata from various sources, and embed it directly into your archive files (`.cbz`, `.zip`).

## 🌟 Features

- **Resource Management**: Scan and manage multiple folders for your manga collection.
- **Metadata Fetching**: Built-in support for multiple providers including Amazon, FANZA, and Bangumi.
- **ComicInfo.xml Support**:
  - Reads and writes standard metadata to your archives safely.
  - **Smart Pathing**: Automatically detects the best location for `ComicInfo.xml` inside archives, supporting nested folder structures (e.g., placing it alongside images in a subfolder).
  - Maintains existing metadata locations when updating.
- **Modern Web Interface**: Beautiful, high-performance UI built with React 19, Vite, and Tailwind CSS.
- **Powerful CLI**: Full control from the terminal for automation and advanced management.
- **Global Proxy**: Integrated proxy support (HTTP, HTTPS, SOCKS5) with provider-level overrides and connectivity testing.

## 🛠 Tech Stack

- **Backend**: Go (Gin, GORM, Cobra, Viper)
- **Frontend**: React (TypeScript, Vite, Tailwind CSS, Lucide Icons)
- **Database**: SQLite

## 🚀 Getting Started

### Prerequisites

- **Go**: 1.25 or higher
- **Node.js**: 22 or higher
- **npm**: (included with Node.js)

### Build from Source

1. **Build the Frontend**:
   ```bash
   cd web
   npm install
   npm run build
   cd ..
   ```

2. **Build the Backend**:
   ```bash
   # This will create a binary named 'mmm' in the root directory
   go build -o mmm main.go
   ```

### Run the Application

Start the web server to access the GUI:
```bash
./mmm serve --port 8080
```
Then open [http://localhost:8080](http://localhost:8080) in your browser.

## 💻 CLI Usage

### Proxy Configuration
Manage your global network settings:
```bash
./mmm proxy set --host 127.0.0.1 --port 7890 --type socks5
./mmm proxy show
./mmm proxy test
```

### Library Operations
Manage your collection folders and scan for changes:
```bash
./mmm library add /path/to/manga
./mmm scan
```

## ⚙️ Configuration

A `config.yaml` file is created on first run (or you can use `config.yaml.example`). It allows you to customize server ports, database paths, and initial proxy settings.

```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  path: "mmm.db"
```

## 🔒 Global Proxy Policy

MMM follows a consistent outbound network policy:
- **Default**: All outbound requests use the global proxy if enabled.
- **Granular Control**: Each provider (Amazon, Bangumi, etc.) can be configured to:
  - `inherit`: Use global proxy.
  - `disabled`: Bypass proxy for this specific source.
  - `custom`: Use a dedicated proxy for this specific source.
- **Bypass**: Requests to `localhost`, `127.0.0.1`, and local network ranges always bypass the proxy.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
