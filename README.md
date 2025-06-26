# GoBrowser

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.23.6+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/)
[![Gio](https://img.shields.io/badge/Fyne-v2.6.1-007ACC?style=for-the-badge&logo=go&logoColor=white)](https://gioui.org/)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen?style=for-the-badge)](Makefile)

**A modern web browser built from scratch in Go using Fyne**

_Learning browser internals by implementing core web technologies_

[🚀 Quick Start](#-quick-start) •
[📖 Documentation](#-documentation) •
[🎯 Features](#-features) •
[🛠️ Development](#️-development) •
[🤝 Contributing](#-contributing)

</div>

---

## 📋 Overview

GoBrowser is an educational web browser implementation built from the ground up using **Go** and **Gio** GUI framework. This project follows the comprehensive [Browser Engineering]:

- 🌐 **HTTP/HTTPS protocol handling**
- 📝 **HTML parsing and DOM tree construction**
- 🎨 **CSS rendering and layout engines**
- ⚡ **JavaScript execution and event handling**
- 🔄 **Multi-tab browsing experience**
- 📱 **Cross-platform desktop support**

> _"Every browser has thousands of unfixed bugs, from the smallest of mistakes to myriad mix ups and mismatches. Every browser must be endlessly tuned and optimized to squeeze out that last bit of performance."_ - Browser Engineering

## 🎯 Features

### ✅ Currently Implemented

- [x] Basic Gio GUI application structure
- [x] Main window and UI components
- [x] Project architecture and build system

### 🚧 In Development

- [ ] **Browser Layout**

  - [ ] Multi-tab interface with `DocTabs`
  - [ ] Navigation toolbar (back, forward, refresh, address bar)
  - [ ] Bookmark management system
  - [ ] Settings and preferences panel

- [ ] **Core Browser Functionality**
  - [ ] HTML parser and tokenizer
  - [ ] CSS parser and styling engine
  - [ ] DOM tree construction and manipulation
  - [ ] JavaScript execution environment
  - [ ] Task scheduling and threading
  - [ ] Network request handling

### 🎯 Planned Features

- [ ] Developer tools and inspector
- [ ] Extensions and plugin system
- [ ] Advanced security features
- [ ] Performance optimization
- [ ] Accessibility support

## 🚀 Quick Start

### Prerequisites

- **Go 1.23.6+** - [Download here](https://golang.org/dl/)
- **Git** - [Download here](https://git-scm.com/)

### Installation

```bash
# Clone the repository
git clone https://github.com/ducnd58233/gobrowser.git
cd gobrowser

# Install dependencies
go mod download

# Install development tools
make install-tools

# Build and run the application
make run-app
```

### Using Pre-built Binary

```bash
# Build the application
make build-app

# Run the built binary
make run-build
```

## 🏗️ Project Structure

```
gobrowser/
├── 📁 cmd/
│   └── main.go                 # Application entry point
├── 📁 internal/
│   ├── 📁 browser/            # Core browser logic
│   │   ├── engine.go          # Browser engine
│   │   ├── tab.go             # Tab management
│   │   ├── const.go           # Constants
│   │   └── utils.go           # Utility functions
│   └── 📁 ui/                 # Fyne UI components
│       ├── main_window.go     # Main window
│       ├── toolbar.go         # Navigation toolbar
│       ├── tabview.go         # Tab interface
│       └── const.go           # UI constants
├── 📁 build/                  # Build artifacts
├── 📄 Makefile               # Build automation
├── 📄 go.mod                 # Go module definition
└── 📄 README.md              # Project documentation
```

## 🛠️ Development

### Build Commands

```bash
# Run the application in development mode
make run-build

# Run tests with coverage
make test
make coverage

# Lint the code
make lint

# Fix linting issues automatically
make fix-lint
```

### Development Tools

The project uses several development tools to ensure code quality:

- **golangci-lint** - Comprehensive Go linter
- **Gio tools** - GUI development utilities
- **Go testing** - Built-in testing framework

### Architecture Principles

Following the browser-engineering methodology, this project emphasizes:

1. **Incremental Development** - Each component builds upon the previous
2. **Educational Value** - Code clarity over performance optimization
3. **Standards Compliance** - Following web standards where applicable
4. **Modular Design** - Clean separation of concerns

## 📖 Documentation

### Learning Resources

- 📚 **[Browser Engineering Book](https://browser.engineering/)** - Primary learning resource
- 🎯 **[Gio Documentation](https://gioui.org/doc/learn/get-started)** - GUI framework documentation
- 🔧 **[Go Documentation](https://golang.org/doc/)** - Go language reference

### Browser Components

| Component   | Description                       | Status         |
| ----------- | --------------------------------- | -------------- |
| HTML Parser | Tokenization and DOM construction | 🚧 Planning    |
| CSS Engine  | Styling and layout computation    | 🚧 Planning    |
| JavaScript  | V8-like execution environment     | 🚧 Planning    |
| Networking  | HTTP/HTTPS request handling       | 🚧 Planning    |
| UI Layer    | Fyne-based user interface         | ✅ In Progress |

## 🤝 Contributing

### How to Contribute

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Write comprehensive tests for new features
- Update documentation for any API changes
- Use meaningful commit messages
