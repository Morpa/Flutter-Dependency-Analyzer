# Flutter Dependency Analyzer

A Go-based command-line tool that analyzes Flutter/Dart projects to identify unused dependencies and provide insights into your project's package management.

## Features

- Scans your Flutter/Dart project for dependency usage
- Identifies potentially unused packages
- Categorizes dependencies by source (hosted vs Git)
- Provides a clean, color-coded summary of your dependencies
- Helps maintain a leaner project by highlighting unnecessary packages

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/flutter-dependency-analyzer
cd flutter-dependency-analyzer

# Build the tool
go build -o dependency-analyzer

# Optional: Move to a directory in your PATH
mv dependency-analyzer /usr/local/bin/
```

## Usage

Navigate to your Flutter/Dart project root directory (where your `pubspec.yaml` file is located) and run:

```bash
dependency-analyzer
```

## How It Works

The tool:
1. Reads and parses your `pubspec.yaml` and `pubspec.lock` files
2. Scans your `lib` directory recursively for import statements
3. Cross-references declared dependencies against actual imports
4. Identifies packages that are declared but not directly imported
5. Categorizes dependencies by type (hosted or Git-based)

## Output Example

```
🔍 Checking dependencies...
🔍 Analyzing dependencies...
📌 package_name_1: 1.0.0 (Git)
📦 package_name_2: 2.1.3
❌ package_name_3: 1.5.7 (unused)

📊 Summary:
   Total dependencies: 10
   Git dependencies: 1
   Hosted dependencies: 8
   Packages possibly unused: 1
⚠️  The following packages seem to not be used in direct imports:
   - package_name_3
   You may consider removing them, but check if they are not used indirectly.
```

## Important Notes

- The tool identifies packages that are not directly imported but doesn't account for:
  - Transitive dependencies (packages required by other packages)
  - Runtime-only dependencies
  - Packages used via reflection or code generation
- Always verify before removing any dependency from your project
- Some packages like `flutter`, `flutter_test`, `flutter_localizations`, and `cupertino_icons` are automatically excluded from the unused check

## Requirements

- Go 1.16 or higher
- A valid Flutter/Dart project with `pubspec.yaml` and `pubspec.lock` files

## License

[MIT License](LICENSE)