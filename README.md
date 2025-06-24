# GoBrowser
## Project Structure
```text
gobrowser/
├── cmd/
│   └── main.go             # app entrypoint: initialize Fyne, config, CLI flags
├── internal/
│   ├── browser/                # core browser logic: URL parsing, HTTP, tab management
│   │   ├── engine.go
│   │   ├── tab.go
│   │   └── renderer.go
│   ├── ui/                     # Fyne UI layer: windows, widgets, dialogs, bindings
│   │   ├── main_window.go
│   │   ├── toolbar.go
│   │   └── tabview.go
│   └── config/                 # loading config/profile, bookmarks, history
│       └── settings.go
├── pkg/                        # reusable components (if open-sourcing or importing)
│   └── utils/                  # e.g. network helpers, URL normalization
│       └── httpclient.go
├── assets/                     # icons, CSS themes, static resources
├── configs/                    # default config files / templates
│   └── default.yaml
├── tests/                      # end‑to‑end or external test cases
├── scripts/                    # build/package automation (e.g. desktop packaging scripts)
└── go.mod
```