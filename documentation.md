# StartDB Local Testing Guide

## Quick Start

### 1. Build the Database

```bash
go build -o bin/startdb ./cmd/startdb
```

### 2. Test Basic Operations

```bash
# Set a value
./bin/startdb set user:1 "John Doe"

# Get a value
./bin/startdb get user:1

# List all keys
./bin/startdb list

# Delete a key
./bin/startdb delete user:1
```

## Interactive Shell Testing

### Start the Shell

```bash
./bin/startdb shell
```

### Basic Commands in Shell

```
startdb> set user:1 "John Doe"
startdb> get user:1
startdb> list
startdb> exists user:1
startdb> delete user:1
startdb> help
startdb> quit
```

## Transaction Testing

### Command Line Transactions

```bash
# Begin transaction
./bin/startdb begin

# Perform operations
./bin/startdb set user:1 "John Doe"
./bin/startdb set user:2 "Jane Smith"

# Commit transaction
./bin/startdb commit

# Or rollback
./bin/startdb rollback
```

### Shell Transactions

```
startdb> begin
startdb> set user:1 "John Doe"
startdb> set user:2 "Jane Smith"
startdb> status
startdb> commit
```

## SQL Testing

### Create Tables

```bash
./bin/startdb sql "CREATE TABLE users (id INTEGER, name TEXT, email TEXT)"
```

### Insert Data

```bash
./bin/startdb sql "INSERT INTO users VALUES (1, 'John', 'john@example.com')"
./bin/startdb sql "INSERT INTO users VALUES (2, 'Jane', 'jane@example.com')"
```

### Query Data

```bash
./bin/startdb sql "SELECT * FROM users"
./bin/startdb sql "SELECT * FROM users WHERE id = 1"
```

### Update Data

```bash
./bin/startdb sql "UPDATE users SET name = 'John Updated' WHERE id = 1"
```

### Delete Data

```bash
./bin/startdb sql "DELETE FROM users WHERE id = 2"
```

### Shell SQL

```
startdb> sql CREATE TABLE products (id INTEGER, name TEXT)
startdb> sql INSERT INTO products VALUES (1, 'Laptop')
startdb> sql SELECT * FROM products
```

## Storage Options

### Memory Storage (Default)

```bash
./bin/startdb set key1 "value1"
./bin/startdb get key1
```

### Disk Storage (Persistent)

```bash
./bin/startdb --storage=disk set user:1 "John Doe"
./bin/startdb --storage=disk get user:1
```

### Custom Data File

```bash
./bin/startdb --storage=disk --data=my_db.json set key1 "value1"
./bin/startdb --storage=disk --data=my_db.json get key1
```

## WAL (Write-Ahead Logging) Testing

### Enable WAL

```bash
./bin/startdb --storage=disk --wal set user:1 "John Doe"
./bin/startdb --storage=disk --wal get user:1
```

### Crash Recovery

```bash
# Create data with WAL
./bin/startdb --storage=disk --wal --data=test.json set user:1 "John Doe"

# Simulate crash (delete data file)
rm test.json

# Recover from WAL
./bin/startdb --storage=disk --wal --data=test.json recover
./bin/startdb --storage=disk --wal --data=test.json get user:1
```

### Checkpoint

```bash
./bin/startdb --storage=disk --wal --data=test.json checkpoint
```

## Complete Test Script

Create `test_script.txt`:

```
set user:1 "John Doe"
set user:2 "Jane Smith"
list
begin
set user:3 "Bob Johnson"
status
commit
sql CREATE TABLE users (id INTEGER, name TEXT)
sql INSERT INTO users VALUES (1, 'John')
sql SELECT * FROM users
quit
```

Run the script:

```bash
Get-Content test_script.txt | ./bin/startdb shell
```

## All Available Commands

### Basic Operations

- `set <key> <value>` - Store a key-value pair
- `get <key>` - Retrieve a value by key
- `delete <key>` - Remove a key-value pair
- `exists <key>` - Check if a key exists
- `list` - List all keys

### Transactions

- `begin` - Start a new transaction
- `commit` - Commit the current transaction
- `rollback` - Rollback the current transaction
- `status` - Show transaction status

### SQL

- `sql <query>` - Execute a SQL query

### Storage

- `checkpoint` - Create a checkpoint (WAL only)
- `recover` - Recover from crash (WAL only)

### System

- `version` - Show version information
- `help` - Show help (in shell)

## Command Line Options

- `--storage=memory` - Use memory storage (default)
- `--storage=disk` - Use disk storage
- `--data=filename.json` - Custom data file
- `--wal` - Enable Write-Ahead Logging
- `--wal-file=filename.wal` - Custom WAL file

## Quick Test Checklist

- [ ] Basic CRUD operations work
- [ ] Transactions work (begin/commit/rollback)
- [ ] SQL queries work (CREATE/INSERT/SELECT)
- [ ] Disk storage persists data
- [ ] WAL recovery works
- [ ] Shell commands work
- [ ] Help system works

## Troubleshooting

**Command not found**: Make sure you built the database with `go build`

**Permission denied**: Check file permissions for data files

**SQL errors**: Ensure table exists before inserting data

**WAL errors**: Make sure `--wal` flag is used consistently

## Clean Up

Remove test files:

```bash
rm -f test*.json test*.wal
```

That's it! You now know how to test StartDB locally. 🚀
