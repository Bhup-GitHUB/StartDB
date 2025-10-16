# StartDB Usage Guide

## Table of Contents

1. [Quick Start](#quick-start)
2. [Installation](#installation)
3. [Interactive Shell Mode](#interactive-shell-mode)
4. [Command Line Mode](#command-line-mode)
5. [Storage Options](#storage-options)
6. [Command Reference](#command-reference)
7. [Examples](#examples)
8. [Troubleshooting](#troubleshooting)

---

## Quick Start

StartDB is an AI-powered database with two main usage modes:

- **Interactive Shell**: Stay in the database and run multiple commands
- **Command Line**: Run single commands and exit

### Build StartDB

```bash
go build -o bin/startdb.exe ./cmd/startdb
```

---

## Installation

### Prerequisites

- Go 1.21 or higher
- Windows, macOS, or Linux

### Build from Source

```bash
# Clone the repository
git clone https://github.com/Bhup-GitHUB/startdb.git
cd startdb

# Build the executable
go build -o bin/startdb.exe ./cmd/startdb

# Run StartDB
./bin/startdb.exe --help
```

---

## Interactive Shell Mode

The interactive shell allows you to stay connected to StartDB and run multiple commands in a session.

### Starting the Shell

#### Memory Storage (Default)

```bash
./bin/startdb.exe shell
```

#### Disk Storage (Persistent)

```bash
./bin/startdb.exe --storage=disk shell
```

#### Custom Database File

```bash
./bin/startdb.exe --storage=disk --data=my_database.json shell
```

### Shell Interface

When you start the shell, you'll see:

```
StartDB Interactive Shell (Storage: disk)
Data file: startdb.json
Type 'help' for commands, 'quit' to exit

startdb>
```

### Shell Commands

| Command             | Description             | Example                 |
| ------------------- | ----------------------- | ----------------------- |
| `set <key> <value>` | Store a key-value pair  | `set user:1 "John Doe"` |
| `get <key>`         | Retrieve a value by key | `get user:1`            |
| `delete <key>`      | Remove a key-value pair | `delete user:1`         |
| `exists <key>`      | Check if a key exists   | `exists user:1`         |
| `list`              | List all keys           | `list`                  |
| `clear`             | Clear the screen        | `clear`                 |
| `help`              | Show available commands | `help`                  |
| `quit` or `exit`    | Exit the shell          | `quit`                  |

### Example Shell Session

```bash
$ ./bin/startdb.exe --storage=disk shell

StartDB Interactive Shell (Storage: disk)
Data file: startdb.json
Type 'help' for commands, 'quit' to exit

startdb> set name "Bhupesh Kumar"
OK
startdb> set age "25"
OK
startdb> set role "Backend Developer"
OK
startdb> list
Found 3 key(s):
1. age
2. name
3. role
startdb> get name
Value: Bhupesh Kumar
startdb> exists email
Exists: false
startdb> set email "bhupesh@example.com"
OK
startdb> delete age
OK
startdb> list
Found 3 key(s):
1. email
2. name
3. role
startdb> quit
Goodbye!
```

---

## Command Line Mode

Run individual commands without entering the interactive shell.

### Basic Syntax

```bash
./bin/startdb.exe [flags] <command> [arguments]
```

### Storage Flags

- `--storage=memory` - In-memory storage (default, temporary)
- `--storage=disk` - Disk storage (persistent)
- `--data=filename.json` - Custom data file path

### Examples

#### Store Data

```bash
# Memory storage
./bin/startdb.exe set user:1 "John Doe"

# Disk storage
./bin/startdb.exe --storage=disk set user:1 "John Doe"

# Custom file
./bin/startdb.exe --storage=disk --data=users.json set user:1 "John Doe"
```

#### Retrieve Data

```bash
# Get value
./bin/startdb.exe --storage=disk get user:1

# Check existence
./bin/startdb.exe --storage=disk exists user:1

# List all keys
./bin/startdb.exe --storage=disk list
```

#### Delete Data

```bash
./bin/startdb.exe --storage=disk delete user:1
```

---

## Storage Options

### Memory Storage

- **Type**: In-memory only
- **Persistence**: Data lost when process exits
- **Use Case**: Temporary data, testing, caching
- **Performance**: Fastest access
- **Command**: `--storage=memory` (default)

```bash
./bin/startdb.exe shell
# or
./bin/startdb.exe set key "value"
```

### Disk Storage

- **Type**: Persistent file-based storage
- **Persistence**: Data survives restarts
- **Use Case**: Production data, long-term storage
- **Performance**: Slightly slower due to I/O
- **Command**: `--storage=disk`

```bash
./bin/startdb.exe --storage=disk shell
# or
./bin/startdb.exe --storage=disk set key "value"
```

### Custom Data Files

- **Multiple Databases**: Use different files for different datasets
- **File Format**: JSON
- **Command**: `--data=filename.json`

```bash
# Users database
./bin/startdb.exe --storage=disk --data=users.json set user:1 "John"

# Products database
./bin/startdb.exe --storage=disk --data=products.json set product:1 "Laptop"

# Orders database
./bin/startdb.exe --storage=disk --data=orders.json set order:1 "Order details"
```

---

## Command Reference

### Global Commands

| Command   | Description              | Flags                 |
| --------- | ------------------------ | --------------------- |
| `shell`   | Start interactive shell  | `--storage`, `--data` |
| `version` | Show version information | None                  |
| `help`    | Show help information    | None                  |

### Data Commands

| Command  | Description           | Arguments       | Example             |
| -------- | --------------------- | --------------- | ------------------- |
| `set`    | Store key-value pair  | `<key> <value>` | `set user:1 "John"` |
| `get`    | Retrieve value        | `<key>`         | `get user:1`        |
| `delete` | Remove key-value pair | `<key>`         | `delete user:1`     |
| `exists` | Check key existence   | `<key>`         | `exists user:1`     |
| `list`   | List all keys         | None            | `list`              |

### Global Flags

| Flag        | Short | Description                     | Default      |
| ----------- | ----- | ------------------------------- | ------------ |
| `--storage` | `-s`  | Storage type (memory/disk)      | memory       |
| `--data`    | `-d`  | Data file path for disk storage | startdb.json |
| `--help`    | `-h`  | Show help                       | -            |
| `--version` | `-v`  | Show version                    | -            |

---

## Examples

### Example 1: User Management System

```bash
# Start with disk storage
./bin/startdb.exe --storage=disk --data=users.json shell

startdb> set user:1 "John Doe"
OK
startdb> set user:2 "Jane Smith"
OK
startdb> set user:3 "Bob Johnson"
OK
startdb> list
Found 3 key(s):
1. user:1
2. user:2
3. user:3
startdb> get user:2
Value: Jane Smith
startdb> delete user:3
OK
startdb> quit
Goodbye!

# Data persists after restart
./bin/startdb.exe --storage=disk --data=users.json get user:1
Value: John Doe
```

### Example 2: Configuration Management

```bash
# Store application configuration
./bin/startdb.exe --storage=disk --data=config.json set database_url "postgres://localhost:5432/mydb"
./bin/startdb.exe --storage=disk --data=config.json set api_key "abc123xyz"
./bin/startdb.exe --storage=disk --data=config.json set debug_mode "true"

# Retrieve configuration
./bin/startdb.exe --storage=disk --data=config.json get database_url
Value: postgres://localhost:5432/mydb
```

### Example 3: Session Management

```bash
# Temporary session data (memory)
./bin/startdb.exe shell

startdb> set session:abc123 "user_id:456"
OK
startdb> set session:abc123:expires "2024-12-31"
OK
startdb> get session:abc123
Value: user_id:456
startdb> quit
Goodbye!

# Session data is lost (memory storage)
./bin/startdb.exe get session:abc123
Error: key not found
```

### Example 4: Multi-Database Setup

```bash
# Users database
./bin/startdb.exe --storage=disk --data=users.json set admin "admin@example.com"
./bin/startdb.exe --storage=disk --data=users.json set guest "guest@example.com"

# Products database
./bin/startdb.exe --storage=disk --data=products.json set laptop "MacBook Pro"
./bin/startdb.exe --storage=disk --data=products.json set phone "iPhone 15"

# Orders database
./bin/startdb.exe --storage=disk --data=orders.json set order:001 "laptop,phone"
./bin/startdb.exe --storage=disk --data=orders.json set order:002 "laptop"

# List all databases
ls *.json
# users.json, products.json, orders.json
```

---

## Troubleshooting

### Common Issues

#### 1. "key not found" Error

```bash
startdb> get nonexistent_key
Error: key not found
```

**Solution**: Check if the key exists using `exists <key>` or `list` commands.

#### 2. "invalid storage type" Error

```bash
./bin/startdb.exe --storage=invalid set key "value"
Error: invalid storage type: invalid (use 'memory' or 'disk')
```

**Solution**: Use `--storage=memory` or `--storage=disk`.

#### 3. "failed to initialize disk storage" Error

```bash
./bin/startdb.exe --storage=disk --data=/invalid/path/file.json set key "value"
Error: failed to initialize disk storage: open /invalid/path/file.json: no such file or directory
```

**Solution**: Ensure the directory exists or use a valid path.

#### 4. "corrupted data file" Error

```bash
./bin/startdb.exe --storage=disk get key
Error: corrupted data file: invalid character 'x' looking for beginning of value
```

**Solution**: Delete the corrupted file and start fresh, or restore from backup.

### Performance Tips

1. **Use Memory Storage for Temporary Data**

   ```bash
   ./bin/startdb.exe shell  # Default is memory
   ```

2. **Use Disk Storage for Persistent Data**

   ```bash
   ./bin/startdb.exe --storage=disk shell
   ```

3. **Use Custom Files for Organization**

   ```bash
   ./bin/startdb.exe --storage=disk --data=users.json shell
   ```

4. **Batch Operations in Shell Mode**
   ```bash
   # More efficient than multiple command-line calls
   ./bin/startdb.exe --storage=disk shell
   startdb> set key1 "value1"
   startdb> set key2 "value2"
   startdb> set key3 "value3"
   ```

### File Locations

- **Default Data File**: `startdb.json` (in current directory)
- **Custom Data Files**: As specified with `--data` flag
- **Binary Location**: `bin/startdb.exe`

### Data Format

StartDB stores data in JSON format:

```json
{
  "data": {
    "key1": "dmFsdWUx", // Base64 encoded
    "key2": "dmFsdWUy"
  }
}
```

---

## Advanced Usage

### Environment Variables

```bash
# Set default storage type
export STARTDB_STORAGE=disk

# Set default data file
export STARTDB_DATA=my_database.json
```

### Scripting

```bash
#!/bin/bash
# backup.sh - Backup all keys from a database

./bin/startdb.exe --storage=disk --data=users.json list > backup.txt
echo "Backup completed: backup.txt"
```

### Integration with Other Tools

```bash
# Export to CSV
./bin/startdb.exe --storage=disk list | while read key; do
    value=$(./bin/startdb.exe --storage=disk get "$key")
    echo "$key,$value"
done > export.csv
```

---

## Getting Help

- **Command Help**: `./bin/startdb.exe --help`
- **Shell Help**: Type `help` in the interactive shell
- **Version Info**: `./bin/startdb.exe version`
- **Specific Command Help**: `./bin/startdb.exe <command> --help`

---

_For more information, visit the [StartDB GitHub Repository](https://github.com/Bhup-GitHUB/startdb)_
