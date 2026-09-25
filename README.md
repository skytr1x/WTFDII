# WTF Did I Install? 🤔

> CLI tool that explains **what's actually in your project's dependencies**.

You install one package:

```bash
npm install some-package
```

And it pulls in 17 more packages, 4 of which you've never seen before.

**WTF Did I Install?** helps you figure it out.

---

## 🎯 What It Does

WTF Did I Install analyzes your project's dependencies and shows you:

- How many dependencies are installed
- Which dependencies are transitive
- Which packages haven't been updated in a while
- Which dependencies are no longer used
- Which packages take up the most space
- Potential security issues
- Duplicate dependencies
- Dependency chains (why a package is installed)
- Package licenses
- Dependencies that can be replaced or removed

## 📦 Supported Package Managers

**Current (Phase 1 - MVP):**
- npm
- pnpm
- Yarn

**Planned (Future phases):**
- pip
- Poetry
- Cargo
- Go modules
- Maven
- Gradle
- NuGet

## 🚀 Installation

### From Source

```bash
git clone https://github.com/skytr1x/wtfdii
cd wtfdii
go build -o wtfdii ./cmd/wtfdii
```

Then move the binary to your PATH:

```bash
sudo mv wtfdii /usr/local/bin/
```

### Using Go Install (Coming Soon)

```bash
go install github.com/skytr1x/wtfdii/cmd/wtfdii@latest
```

## 📖 Usage

### Overview

Show a summary of your project's dependencies:

```bash
wtfdii
```

Output:

```text
WTF Did I Install?

Project: my-awesome-app
Package manager: npm

Dependencies
────────────────────────────────
Direct dependencies       23
Transitive dependencies   187
Total packages            210

Disk usage
────────────────────────────────
node_modules              184 MB

⚠ Things you might want to know

  🔴 3 packages have security advisories
  Run 'wtfdii security' for security audit
  Run 'wtfdii stale' for outdated packages
  Run 'wtfdii unused' for unused dependencies
```

### Why is this package installed?

Find out why a specific package is in your project:

```bash
wtfdii why lodash
```

Output:

```text
lodash

Installed because:

your-app
└── webpack
    └── some-plugin
        └── lodash
```

### Show largest packages

See which packages take up the most disk space:

```bash
wtfdii size
```

Output:

```text
Largest packages
────────────────────────────────────────────────────────────
Package                                  Size
────────────────────────────────────────────────────────────
puppeteer                                82.4 MB
typescript                               24.1 MB
@swc/core                                18.7 MB
eslint                                    9.2 MB
lodash                                    4.8 MB
```

You can limit the number of results:

```bash
wtfdii size --limit 5
```

### Security audit

Check for known security vulnerabilities:

```bash
wtfdii security
```

Output:

```text
Security

🔴 2 high severity vulnerabilities
🟡 4 moderate vulnerabilities

Affected packages:

high minimist 1.2.5
  Prototype Pollution in minimist
  https://npmjs.com/advisories/1179
```

## ⚙️ Configuration

Create a `.wtfdiirc` file in your project root:

```json
{
  "stale_threshold_months": 24,
  "ignore_packages": [
    "some-internal-package"
  ],
  "custom_rules": {}
}
```

### Configuration Options

- `stale_threshold_months` (default: 24) - How many months old a package should be to be considered stale
- `ignore_packages` - Array of package names to ignore in analysis
- `custom_rules` - Custom rules for analysis (coming soon)

## 🎨 Options

### Global Flags

- `--path, -p <path>` - Project path to analyze (default: current directory)
- `--no-color` - Disable colored output

### Commands

- `wtfdii` - Show overview of dependencies
- `wtfdii why <package>` - Show why a package is installed
- `wtfdii size` - Show largest packages
- `wtfdii security` - Run security audit
- `wtfdii stale` - Show outdated packages (coming in Phase 2)
- `wtfdii unused` - Show potentially unused dependencies (coming in Phase 3)

## 🏗️ Architecture

The project uses a modular architecture with adapters for different package managers:

```
wtfdii/
├── cmd/wtfdii/           # CLI entry point
├── internal/
│   ├── analyzer/         # Core analysis logic
│   ├── packagemanager/   # Package manager adapters
│   │   ├── manager.go    # Interface
│   │   ├── npm.go        # NPM adapter
│   │   ├── pnpm.go       # PNPM adapter
│   │   └── yarn.go       # Yarn adapter
│   ├── models/           # Data structures
│   ├── output/           # Output formatting
│   └── config/           # Configuration management
└── docs/                 # Documentation
```

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Node.js (for testing with npm/pnpm/yarn projects)

### Building

```bash
go build -o wtfdii ./cmd/wtfdii
```

### Running Tests

```bash
go test ./...
```

## 🗺️ Roadmap

### Phase 1 (MVP) - ✅ Complete
- [x] Basic dependency overview
- [x] `wtfdii why <package>` command
- [x] `wtfdii size` command
- [x] Support for npm, pnpm, yarn
- [x] Configuration file support

### Phase 2
- [ ] `wtfdii stale` - Show outdated packages
- [ ] `wtfdii security` - Enhanced security analysis
- [ ] Better dependency chain visualization
- [ ] Duplicate dependencies detection

### Phase 3
- [ ] `wtfdii unused` - Detect unused dependencies
- [ ] License analysis
- [ ] Package replacement suggestions
- [ ] Support for pip, Poetry, Cargo

## 📝 License

MIT

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 💡 Inspiration

This tool was inspired by the common frustration of not knowing what's actually installed in `node_modules` and why.

---

Made with ❤️ by skytr1x
# WTFDII
