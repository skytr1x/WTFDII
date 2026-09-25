# WTF Did I Install?

> CLI-инструмент, который объясняет, **что на самом деле находится в зависимостях твоего проекта**.

Ты устанавливаешь одну библиотеку:

```bash
npm install some-package
```

А она тянет за собой ещё 17 пакетов, 4 из которых ты никогда не видел.

**WTF Did I Install?** помогает разобраться.

---

## 🎯 Идея

Инструмент анализирует зависимости проекта и показывает:

- сколько зависимостей установлено;
- какие зависимости являются транзитивными;
- какие пакеты давно не обновлялись;
- какие зависимости больше не используются;
- какие пакеты занимают больше всего места;
- потенциальные security-проблемы;
- дублирующиеся зависимости;
- зависимости с подозрительно большой dependency tree;
- лицензии пакетов;
- зависимости, которые можно заменить или удалить.

Пример:

```text
$ wtf

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
  🟡 14 packages haven't been updated for 2+ years
  🟡 7 packages are duplicated
  🟢 4 dependencies appear to be unused

Largest packages
────────────────────────────────
  puppeteer                82 MB
  typescript               24 MB
  eslint                    9 MB
  lodash                    4 MB

Dependency tree

your-app
├── next
│   ├── react
│   ├── react-dom
│   └── ...
├── prisma
│   └── ...
└── zod
```

---

# ✨ Features

## 1. Dependency overview

Показывает общую картину проекта:

```text
Direct dependencies:       23
Transitive dependencies: 187
Total:                    210
```

Это позволяет быстро понять реальный размер dependency tree.

---

## 2. Почему этот пакет установлен?

Одна из главных функций.

Например:

```bash
wtf why lodash
```

Результат:

```text
lodash

Installed because:

your-app
└── webpack
    └── some-plugin
        └── lodash
```

То есть `lodash` может вообще не находиться в `package.json`, но всё равно присутствовать в проекте.

---

## 3. Кто занимает место?

```bash
wtf size
```

Пример:

```text
Package                     Size
────────────────────────────────────
puppeteer                   82.4 MB
typescript                  24.1 MB
@swc/core                   18.7 MB
eslint                       9.2 MB
lodash                      4.8 MB
```

Полезно для поиска неожиданных тяжёлых зависимостей.

---

## 4. Старые зависимости

```bash
wtf stale
```

Пример:

```text
Potentially stale dependencies

package              current       latest       age
────────────────────────────────────────────────────
lodash               4.17.15       4.17.21      4 years
moment               2.29.1        2.30.1       3 years
some-package         1.2.0         4.1.0        2 years
```

> Инструмент не считает старую версию автоматически уязвимостью. Старость и security-проблема — разные вещи.

---

## 5. Security audit

```bash
wtf security
```

Пример:

```text
Security

🔴 2 high severity vulnerabilities
🟡 4 moderate vulnerabilities

Affected packages:

minimist 1.2.5
└── used by some-package
```

В будущем можно подключить несколько vulnerability databases.

---

## 6. Неиспользуемые зависимости

```bash
wtf unused
```

Пример:

```text
Potentially unused dependencies

  axios
  moment
  lodash
```

Важно: это **эвристика**, а не гарантия.

Например, пакет может использоваться через:

- dynamic imports;
- configuration files;
- plugins;
- scripts;
- runtime loading.

Поэтому программа должна писать:

```text
⚠ Potentially unused
```

а не:

```text
❌ Definitely unused
```

---

# 📦 Поддерживаемые package managers

Первая версия:

- npm
- pnpm
- Yarn

В будущем:

- pip
- Poetry
- Cargo
- Go modules
- Maven
- Gradle
- NuGet

Архитектура должна позволять добавлять новые package managers через отдельные adapters.

Например:

```text
PackageManager
      │
      ├── NpmAdapter
      ├── PnpmAdapter
      ├── YarnAdapter
      ├── PipAdapter
      └── CargoAdapter
```

---

# 🚀 Installation

Пока проект находится в разработке:

```bash
git clone https://github.com/USERNAME/wtf-did-i-install
cd wtf-did-i-install

npm install
npm run build
```

П
