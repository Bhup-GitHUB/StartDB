# StartDB Quick Reference

## 🚀 Quick Start

```bash
# Build
go build -o bin/startdb.exe ./cmd/startdb

# Interactive Shell (Recommended)
./bin/startdb.exe shell                    # Memory storage
./bin/startdb.exe --storage=disk shell     # Disk storage

# Single Commands
./bin/startdb.exe set key "value"          # Memory storage
./bin/startdb.exe --storage=disk set key "value"  # Disk storage
```

## 📋 Commands

| Command             | Description     | Example             |
| ------------------- | --------------- | ------------------- |
| `set <key> <value>` | Store data      | `set user:1 "John"` |
| `get <key>`         | Get data        | `get user:1`        |
| `delete <key>`      | Delete data     | `delete user:1`     |
| `exists <key>`      | Check existence | `exists user:1`     |
| `list`              | List all keys   | `list`              |
| `help`              | Show help       | `help`              |
| `quit`              | Exit shell      | `quit`              |

## ⚙️ Storage Options

| Option             | Description         | Use Case                |
| ------------------ | ------------------- | ----------------------- |
| `--storage=memory` | In-memory (default) | Temporary data, testing |
| `--storage=disk`   | Persistent file     | Production data         |
| `--data=file.json` | Custom file         | Multiple databases      |

## 💡 Examples

### User Management

```bash
./bin/startdb.exe --storage=disk --data=users.json shell
startdb> set user:1 "John Doe"
startdb> set user:2 "Jane Smith"
startdb> list
startdb> get user:1
startdb> quit
```

### Configuration

```bash
./bin/startdb.exe --storage=disk --data=config.json set api_key "abc123"
./bin/startdb.exe --storage=disk --data=config.json get api_key
```

### Multiple Databases

```bash
# Users
./bin/startdb.exe --storage=disk --data=users.json set admin "admin@example.com"

# Products
./bin/startdb.exe --storage=disk --data=products.json set laptop "MacBook Pro"

# Orders
./bin/startdb.exe --storage=disk --data=orders.json set order:1 "laptop,phone"
```

## 🔧 Troubleshooting

| Error                               | Solution                                   |
| ----------------------------------- | ------------------------------------------ |
| `key not found`                     | Use `exists <key>` or `list` to check      |
| `invalid storage type`              | Use `--storage=memory` or `--storage=disk` |
| `failed to initialize disk storage` | Check file path exists                     |
| `corrupted data file`               | Delete file and start fresh                |

## 📁 File Locations

- **Binary**: `bin/startdb.exe`
- **Default Data**: `startdb.json`
- **Custom Data**: As specified with `--data`

---

**Full Documentation**: [docs/USAGE.md](USAGE.md)
