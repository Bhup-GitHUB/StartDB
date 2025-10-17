# StartDB Testing Documentation

## Overview

This document provides comprehensive testing steps for StartDB's Write-Ahead Logging (WAL) functionality. The WAL implementation ensures data durability and crash recovery capabilities.

## Prerequisites

- Go 1.21 or higher
- Windows, macOS, or Linux
- StartDB built and ready (`go build -o bin/startdb.exe ./cmd/startdb`)

## Test Categories

### 1. Basic WAL Functionality Tests

#### Test 1.1: Enable WAL with Disk Storage

```bash
# Test basic WAL operations
./bin/startdb.exe --storage=disk --wal --data=test_basic.json set user:1 "John Doe"
./bin/startdb.exe --storage=disk --wal --data=test_basic.json get user:1
./bin/startdb.exe --storage=disk --wal --data=test_basic.json list

# Expected: Should show user:1 with value "John Doe"
```

#### Test 1.2: WAL File Creation

```bash
# Check if WAL file is created
ls -la test_basic.json.wal

# Expected: WAL file should exist and contain binary data
```

#### Test 1.3: Multiple Operations

```bash
# Perform multiple operations
./bin/startdb.exe --storage=disk --wal --data=test_multi.json set user:1 "John Doe"
./bin/startdb.exe --storage=disk --wal --data=test_multi.json set user:2 "Jane Smith"
./bin/startdb.exe --storage=disk --wal --data=test_multi.json delete user:1
./bin/startdb.exe --storage=disk --wal --data=test_multi.json list

# Expected: Only user:2 should remain
```

### 2. Crash Recovery Tests

#### Test 2.1: Simulate Crash and Recovery

```bash
# Step 1: Create data with WAL
./bin/startdb.exe --storage=disk --wal --data=test_crash.json set user:1 "John Doe"
./bin/startdb.exe --storage=disk --wal --data=test_crash.json set user:2 "Jane Smith"

# Step 2: Simulate crash by deleting data file (keep WAL)
rm test_crash.json

# Step 3: Recover using WAL
./bin/startdb.exe --storage=disk --wal --data=test_crash.json recover

# Step 4: Verify data recovery
./bin/startdb.exe --storage=disk --wal --data=test_crash.json list

# Expected: Both users should be recovered
```

#### Test 2.2: Partial Crash Recovery

```bash
# Step 1: Create initial data
./bin/startdb.exe --storage=disk --wal --data=test_partial.json set user:1 "John Doe"

# Step 2: Add more data
./bin/startdb.exe --storage=disk --wal --data=test_partial.json set user:2 "Jane Smith"
./bin/startdb.exe --storage=disk --wal --data=test_partial.json delete user:1

# Step 3: Simulate crash
rm test_partial.json

# Step 4: Recover
./bin/startdb.exe --storage=disk --wal --data=test_partial.json recover

# Step 5: Verify
./bin/startdb.exe --storage=disk --wal --data=test_partial.json get user:2

# Expected: user:2 should exist, user:1 should not
```

### 3. Checkpoint Tests

#### Test 3.1: Create Checkpoint

```bash
# Step 1: Create data
./bin/startdb.exe --storage=disk --wal --data=test_checkpoint.json set user:1 "John Doe"
./bin/startdb.exe --storage=disk --wal --data=test_checkpoint.json set user:2 "Jane Smith"

# Step 2: Check WAL file size before checkpoint
ls -la test_checkpoint.json.wal

# Step 3: Create checkpoint
./bin/startdb.exe --storage=disk --wal --data=test_checkpoint.json checkpoint

# Step 4: Check WAL file size after checkpoint
ls -la test_checkpoint.json.wal

# Expected: WAL file should be smaller or empty after checkpoint
```

#### Test 3.2: Data Persistence After Checkpoint

```bash
# Step 1: Create data and checkpoint
./bin/startdb.exe --storage=disk --wal --data=test_persist.json set user:1 "John Doe"
./bin/startdb.exe --storage=disk --wal --data=test_persist.json checkpoint

# Step 2: Verify data still exists
./bin/startdb.exe --storage=disk --wal --data=test_persist.json get user:1

# Expected: Data should still be accessible
```

### 4. Interactive Shell Tests

#### Test 4.1: WAL Shell Commands

```bash
# Start interactive shell with WAL
./bin/startdb.exe --storage=disk --wal --data=test_shell.json shell

# In the shell, run these commands:
# set user:1 "John Doe"
# wal-status
# list
# checkpoint
# wal-status
# quit

# Expected: WAL status should show enabled, checkpoint should work
```

#### Test 4.2: Shell WAL Recovery

```bash
# Start shell and create data
./bin/startdb.exe --storage=disk --wal --data=test_shell_recovery.json shell

# In shell:
# set user:1 "John Doe"
# set user:2 "Jane Smith"
# quit

# Simulate crash
rm test_shell_recovery.json

# Start shell again (should auto-recover)
./bin/startdb.exe --storage=disk --wal --data=test_shell_recovery.json shell

# In shell:
# list
# quit

# Expected: Both users should be recovered automatically
```

### 5. Memory Engine WAL Tests

#### Test 5.1: Memory Engine with WAL

```bash
# Test memory engine with WAL
./bin/startdb.exe --storage=memory --wal --wal-file=test_memory.wal set user:1 "John Doe"
./bin/startdb.exe --storage=memory --wal --wal-file=test_memory.wal get user:1

# Expected: Should work with memory storage
```

#### Test 5.2: Memory Engine Recovery

```bash
# Create data in memory with WAL
./bin/startdb.exe --storage=memory --wal --wal-file=test_memory_recovery.wal set user:1 "John Doe"

# Restart and recover
./bin/startdb.exe --storage=memory --wal --wal-file=test_memory_recovery.wal recover
./bin/startdb.exe --storage=memory --wal --wal-file=test_memory_recovery.wal get user:1

# Expected: Data should be recovered from WAL
```

### 6. Error Handling Tests

#### Test 6.1: Invalid WAL File

```bash
# Create corrupted WAL file
echo "corrupted data" > test_corrupted.wal

# Try to use corrupted WAL
./bin/startdb.exe --storage=disk --wal --wal-file=test_corrupted.wal --data=test_corrupted.json get user:1

# Expected: Should handle corruption gracefully
```

#### Test 6.2: WAL Without Storage

```bash
# Try to use WAL commands without WAL enabled
./bin/startdb.exe --storage=disk --data=test_no_wal.json checkpoint

# Expected: Should show error about WAL not being enabled
```

### 7. Performance Tests

#### Test 7.1: Large Dataset WAL

```bash
# Create large dataset with WAL
for i in {1..100}; do
    ./bin/startdb.exe --storage=disk --wal --data=test_large.json set "key$i" "value$i"
done

# Check WAL file size
ls -la test_large.json.wal

# Create checkpoint
./bin/startdb.exe --storage=disk --wal --data=test_large.json checkpoint

# Verify all data exists
./bin/startdb.exe --storage=disk --wal --data=test_large.json list | wc -l

# Expected: Should show 100 keys
```

### 8. Concurrent Access Tests

#### Test 8.1: Multiple Processes

```bash
# Terminal 1: Start shell
./bin/startdb.exe --storage=disk --wal --data=test_concurrent.json shell

# Terminal 2: Use command line
./bin/startdb.exe --storage=disk --wal --data=test_concurrent.json set user:1 "John Doe"

# Expected: Both should work without conflicts
```

## Test Cleanup

After testing, clean up test files:

```bash
# Remove all test files
rm -f test_*.json test_*.wal
```

## Expected Results Summary

| Test Category     | Expected Outcome                      |
| ----------------- | ------------------------------------- |
| Basic WAL         | All operations logged and recoverable |
| Crash Recovery    | Data fully recovered from WAL         |
| Checkpoint        | WAL truncated, data preserved         |
| Interactive Shell | WAL commands work in shell            |
| Memory Engine     | WAL works with memory storage         |
| Error Handling    | Graceful error handling               |
| Performance       | Efficient WAL operations              |
| Concurrent Access | No data corruption                    |

## Troubleshooting

### Common Issues

1. **WAL file not created**: Ensure `--wal` flag is used
2. **Recovery fails**: Check WAL file permissions and corruption
3. **Checkpoint fails**: Ensure WAL is enabled
4. **Shell commands not working**: Verify WAL is enabled in shell

### Debug Commands

```bash
# Check WAL file contents (binary)
hexdump -C test_file.wal

# Check file permissions
ls -la test_file.wal

# Verify data file
cat test_file.json
```

## Test Automation

For automated testing, create a test script:

```bash
#!/bin/bash
# test_wal.sh

echo "Starting WAL tests..."

# Run all tests
echo "Test 1: Basic WAL functionality"
./bin/startdb.exe --storage=disk --wal --data=test1.json set user:1 "John Doe"
./bin/startdb.exe --storage=disk --wal --data=test1.json get user:1

echo "Test 2: Crash recovery"
rm test1.json
./bin/startdb.exe --storage=disk --wal --data=test1.json recover
./bin/startdb.exe --storage=disk --wal --data=test1.json get user:1

echo "Test 3: Checkpoint"
./bin/startdb.exe --storage=disk --wal --data=test1.json checkpoint

# Cleanup
rm -f test1.json test1.json.wal

echo "All tests completed!"
```

This documentation provides comprehensive testing coverage for all WAL functionality in StartDB.
