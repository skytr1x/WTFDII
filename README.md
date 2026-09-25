# What the FUCK did i install?

Imagine situation where you installed just one package

```bash
npm install package
```

And then you figuring out that there is 20 more packages, about 4 of them you didn't even knew.
What the FUCK did i install helps you out.

What matters, that What the FUCK did i install made on Go. That's makes it a lot faster than internal tools of npm, pip and any of supported package managers.

## What it does?

What the FUCK did I Install analyzes your project's dependencies and shows you:

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

## Supported package managers

**Currently supported:**

- npm
- pnpm
- Yarn

**Future plans:**

- pip
- Poetry
- Cargo
- Go modules
- Maven
- Gradle
- NuGet

## Installation

Currently you can only build it from source:

```bash
git clone https://github.com/skytr1x/wtfdii
cd wtfdii
go mod tidy
go build -o wtfdii ./cmd/wtfdii

sudo mv wtfdii /usr/local/bin/
```

## Usage

Show a summary of your project's dependencies:

```bash
> wtfdii
```

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

---

You can also find out **why** package is installed:

```bash
> wtfdii why lodash
```

```text
lodash

Installed because:

your-app
└── webpack
    └── some-plugin
        └── lodash
```

---

You can check size of every package:

```bash
> wtfdii size
```

```text
────────────────────────────────────────────────────────────
Package                                  Size
────────────────────────────────────────────────────────────
puppeteer                                82.4 MB
typescript                               24.1 MB
@swc/core                                18.7 MB
eslint                                    9.2 MB
lodash                                    4.8 MB
```

---

Feel unsecure with your packages? You can check it out:

```bash
> wtfdii security
```

```text
Security

🔴 2 high severity vulnerabilities
🟡 4 moderate vulnerabilities

Affected packages:

high minimist 1.2.5
  Prototype Pollution in minimist
  https://npmjs.com/advisories/1179
```

---

Wanna check for outdated packages?

```bash
> wtfdii stale
```

```text
Potentially stale dependencies

package                                  current         latest
───────────────────────────────────────────────────────────────────────────
lodash                                   4.17.15         4.18.1

🟡 1 outdated packages found
```

## Configuration

You can create `.wtfdiirc` in your project root:

```json
{
    "stale_threshold_months": 24,
    "ignore_packages": ["some-internal-package"],
    "custom_rules": {}
}
```

- `stale_threshold_months` < How many months old a package should be to be considered stale (default: 24)
- `ignore_packages` - Array of package names to ignore in analysis
- `custom_rules` - Custom rules for analysis (still in plans tho)
