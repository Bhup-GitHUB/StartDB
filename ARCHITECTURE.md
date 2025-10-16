# StartDB System Architecture

## 📋 Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Design Principles](#design-principles)
3. [Layer-by-Layer Deep Dive](#layer-by-layer-deep-dive)
4. [Data Flow Patterns](#data-flow-patterns)
5. [Communication Protocols](#communication-protocols)
6. [Scalability & Performance](#scalability--performance)
7. [Failure Handling](#failure-handling)
8. [Security Considerations](#security-considerations)

---

## Architecture Overview

StartDB follows a **layered microservices architecture** with clear separation between:

- **Core Database Engine** (Go) - Low-latency OLTP operations
- **AI Optimization Service** (TypeScript) - Asynchronous pattern analysis
- **Observability Layer** - Monitoring and alerting

### Why This Architecture?

**Problem**: Traditional monolithic databases tightly couple query processing with optimization logic, making it difficult to:

- Scale components independently
- Experiment with ML models without affecting production
- Hot-swap optimization strategies

**Solution**: Separate the deterministic (database operations) from the probabilistic (AI predictions).

---

## Design Principles

### 1. **Separation of Concerns**

Each layer has a single, well-defined responsibility:

- **Client Layer**: User interface and protocols
- **Gateway Layer**: Security, routing, load balancing
- **Query Layer**: SQL parsing and optimization
- **Transaction Layer**: ACID guarantees
- **Storage Layer**: Data persistence
- **AI Layer**: Pattern learning and recommendations

### 2. **Loose Coupling via Interfaces**

Components communicate through well-defined APIs:

```go
// Storage Engine Interface
type StorageEngine interface {
    Get(key string) ([]byte, error)
    Put(key string, value []byte) error
    Delete(key string) error
    Scan(prefix string) Iterator
}
```

This allows swapping implementations (e.g., BoltDB → RocksDB) without changing other layers.

### 3. **Asynchronous AI Feedback Loop**

Critical decision: **AI analysis must NOT block database operations**.

```
Query Execution → [Async Queue] → AI Service → [Recommendations Queue] → Apply in Background
```

This ensures:

- Zero latency impact on queries
- AI service failures don't affect database availability
- Recommendations can be validated before application

### 4. **Observability by Design**

Every component emits:

- **Metrics**: Latency, throughput, error rates (Prometheus format)
- **Logs**: Structured JSON logs for aggregation
- **Traces**: Distributed tracing (OpenTelemetry compatible)

---

## Layer-by-Layer Deep Dive

### Layer 1: Client Layer

**Purpose**: Provide multiple interfaces for database interaction

**Components**:

#### 1.1 CLI Interface

```bash
startdb> INSERT INTO users (id, name) VALUES (1, 'Alice');
OK (2.3ms)
```

**Implementation**:

- Written in Go using `cobra` library
- Maintains persistent connection to backend
- Supports scripting and batch operations

#### 1.2 REST API

```
POST /api/v1/query
Content-Type: application/json

{
  "query": "SELECT * FROM users WHERE age > 21",
  "params": []
}
```

**Why REST?**:

- Universal compatibility
- Easy debugging with curl/Postman
- Language-agnostic

#### 1.3 Client SDKs

```python
# Python SDK
from startdb import Client

db = Client("localhost:8080")
result = db.query("SELECT * FROM users WHERE id = ?", [123])
```

**Planned SDKs**: Python, JavaScript, Go, Java

---

### Layer 2: API Gateway & Load Balancer

**Purpose**: Single entry point for all client requests

**Components**:

#### 2.1 Authentication Module

**Mechanism**: JWT-based stateless auth

```go
type AuthMiddleware struct {
    secretKey []byte
}

func (a *AuthMiddleware) Authenticate(token string) (*User, error) {
    // Validate JWT signature
    // Check expiration
    // Return user context
}
```

**Features**:

- API key management
- Role-based access control (RBAC)
- Rate limiting per user/API key

#### 2.2 Request Router

**Responsibility**: Route requests to backend instances

**Strategy**: Consistent hashing for connection affinity

```
hash(client_id) % num_instances → instance_id
```

**Why?**: Ensures same client connects to same instance, improving cache hit rate.

#### 2.3 Connection Pool

**Problem**: Creating TCP connections is expensive
**Solution**: Maintain pool of persistent connections

```go
type ConnectionPool struct {
    connections chan *Connection
    maxSize     int
}
```

**Configuration**:

- Min connections: 10
- Max connections: 1000
- Idle timeout: 5 minutes

---

### Layer 3: Query Coordinator Layer

**Purpose**: Parse, plan, and execute queries

#### 3.1 SQL Parser

**Input**: Raw SQL string
**Output**: Abstract Syntax Tree (AST)

**Example**:

```sql
SELECT name, email FROM users WHERE age > 21 ORDER BY name LIMIT 10;
```

**AST Structure**:

```go
type SelectStatement struct {
    Columns   []string          // ["name", "email"]
    Table     string            // "users"
    Where     *WhereClause      // age > 21
    OrderBy   []OrderByClause   // name ASC
    Limit     int               // 10
}
```

**Implementation**: Hand-written recursive descent parser (Phase 1), later ANTLR-based (Phase 3).

#### 3.2 Query Planner

**Responsibility**: Generate optimal execution plan

**Process**:

1. **Cost estimation** for each possible plan
2. **Index selection** based on WHERE clauses
3. **Join order optimization** (for multi-table queries)

**Example**:

```
Query: SELECT * FROM users WHERE age > 21 AND city = 'NYC'

Plan Options:
Option A: Full table scan → Filter age → Filter city (Cost: 1000 units)
Option B: Index scan on age → Filter city (Cost: 100 units)
Option C: Index scan on city → Filter age (Cost: 80 units) ✅ CHOSEN
```

**Statistics Used**:

- Table cardinality (row count)
- Index selectivity
- Column distribution histograms

#### 3.3 Query Executor

**Responsibility**: Execute the chosen plan

**Execution Models**:

- **Volcano Model**: Iterator-based, one row at a time
- **Vectorized**: Process batches (Phase 2 enhancement)

**Pseudocode**:

```go
func (e *Executor) Execute(plan *Plan) (*ResultSet, error) {
    iterator := plan.GetIterator()
    results := []Row{}

    for iterator.Next() {
        row := iterator.Current()
        results = append(results, row)
    }

    return &ResultSet{rows: results}, nil
}
```

---

### Layer 4: Transaction & Concurrency Layer

**Purpose**: Ensure ACID properties

#### 4.1 Transaction Manager

**Isolation Levels Supported**:

1. Read Uncommitted
2. Read Committed ✅ (Default)
3. Repeatable Read
4. Serializable

**Implementation**: MVCC (Multi-Version Concurrency Control)

**How MVCC Works**:

```
Row: { id: 1, name: "Alice", version: 5, txn_id: 100 }

Transaction 101 reads → Sees version 5
Transaction 102 writes → Creates version 6
Transaction 101 still sees version 5 (snapshot isolation)
```

**Benefits**:

- Readers don't block writers
- Writers don't block readers
- High concurrency

#### 4.2 Lock Manager

**Lock Granularity**:

- **Row-level locks** (default)
- **Table-level locks** (for DDL operations)

**Lock Types**:

- **Shared Lock (S)**: Multiple readers allowed
- **Exclusive Lock (X)**: Single writer, no readers

**Deadlock Detection**:

- **Wait-for graph** approach
- Check for cycles every 1 second
- Abort youngest transaction on deadlock

#### 4.3 WAL Manager

**Purpose**: Ensure durability even on crash

**WAL Protocol**:

```
1. Write operation to WAL (disk flush)
2. Update in-memory data structures
3. Return success to client
4. Asynchronously checkpoint to data files
```

**WAL Entry Format**:

```
[LSN][TXN_ID][OPERATION][TABLE][KEY][OLD_VALUE][NEW_VALUE][CHECKSUM]
```

**Recovery Process**:

```go
func RecoverFromWAL() {
    lastCheckpoint := findLastCheckpoint()

    for entry := range readWAL(lastCheckpoint) {
        if entry.TxnID in committedTransactions {
            replayEntry(entry)  // Redo
        } else {
            // Transaction was incomplete, ignore (Undo not needed with MVCC)
        }
    }
}
```

---

### Layer 5: Storage Engine Layer

**Purpose**: Manage in-memory and disk data structures

#### 5.1 Buffer Pool Manager

**Concept**: In-memory cache of disk pages

**Structure**:

```go
type BufferPool struct {
    pages       map[PageID]*Page
    freeList    []PageID
    dirtyPages  map[PageID]bool
    evictionLRU *LRUCache
    mutex       sync.RWMutex
}
```

**Operations**:

- **Pin(pageID)**: Load page into memory, increment reference count
- **Unpin(pageID)**: Decrement reference count
- **FlushDirty()**: Write modified pages to disk

**LRU Eviction Policy**:

- Only evict pages with refCount == 0
- Prioritize clean pages (dirty pages require disk write)

**Size Configuration**:

- Default: 25% of available RAM
- Adjustable via config file

#### 5.2 Index Manager

**Index Types**:

**1. B-Tree Index** (Default for range queries)

```
Structure:
         [50]
        /    \
    [25]      [75]
   /   \      /   \
[10] [30]  [60] [90]
```

**Use Cases**:

- Range queries: `WHERE age BETWEEN 20 AND 30`
- Sorted access: `ORDER BY created_at`

**2. Hash Index** (For equality lookups)

```
hash("alice@example.com") → Bucket 42 → [Row Pointer]
```

**Use Cases**:

- Exact match: `WHERE email = 'alice@example.com'`
- Join operations on unique keys

**Index Creation**:

```sql
CREATE INDEX idx_users_email ON users(email) USING HASH;
```

**Automatic Index Recommendation** (AI-driven):

```
AI detects: 80% of queries filter on "city" column
→ Recommends: CREATE INDEX idx_users_city ON users(city)
→ DBA approves or auto-applies
```

#### 5.3 Cache Manager

**Purpose**: Cache query results

**Cache Layers**:

1. **Query Result Cache**: Full result sets
2. **Parsed Query Cache**: Reuse ASTs
3. **Execution Plan Cache**: Avoid re-planning

**Invalidation Strategy**:

```go
type CacheEntry struct {
    query       string
    result      *ResultSet
    dependencies []string  // ["users", "orders"]
    createdAt   time.Time
}

// Invalidate on write
func (c *CacheManager) OnWrite(table string) {
    for key, entry := range c.entries {
        if contains(entry.dependencies, table) {
            c.Delete(key)
        }
    }
}
```

**TTL-based Expiration**:

- Configurable per-query
- Default: 5 minutes

---

### Layer 6: Persistence Layer

**Purpose**: Durable storage on disk

#### 6.1 Data Files

**File Format**:

```
[Header: 4KB]
  - Magic number
  - Version
  - Checksum
  - Page size

[Pages: 4KB each]
  Page 0: Meta page (table schema, index locations)
  Page 1-N: Data pages
```

**Page Structure**:

```
[Page Header: 64 bytes]
  - Page ID
  - Page type (data/index)
  - Free space pointer
  - Slot count

[Slot Array]
  [Slot 0: offset, length]
  [Slot 1: offset, length]

[Free Space]

[Tuples (stored bottom-up)]
  [Tuple N]
  [Tuple N-1]
```

**Why Bottom-Up?**: Allows page to fill efficiently without fragmentation.

#### 6.2 WAL Log Files

**Append-Only Design**:

```
wal_000001.log (0-100MB)
wal_000002.log (100-200MB)
...
```

**Rotation Policy**:

- New file every 100MB
- Keep last 10 files
- Archive to cold storage

#### 6.3 Index Files

**Separate Files per Index**:

```
users_idx_email.btree
users_idx_age.btree
orders_idx_user_id.hash
```

**Why Separate?**: Easier to drop indexes without affecting data files.

---

## AI Optimization Service (Separate Process)

**Language**: TypeScript/Node.js
**Why TypeScript?**:

- Rich ecosystem for ML (TensorFlow.js, Brain.js)
- Excellent async handling for batch processing
- Fast prototyping for ML experiments

### Components:

#### 1. Query Log Collector

**Data Source**: Real-time stream from Go backend

**Metrics Collected**:

```typescript
interface QueryLog {
  query: string;
  executionTime: number;
  timestamp: Date;
  rowsScanned: number;
  rowsReturned: number;
  indexesUsed: string[];
  cacheHit: boolean;
}
```

**Storage**: Time-series database (InfluxDB or Prometheus)

#### 2. Pattern Analyzer

**Statistical Analysis**:

```typescript
class PatternAnalyzer {
  // Detect most frequent queries
  findHotQueries(): Query[] {
    // Count queries by normalized form
    // Return top 10% by frequency
  }

  // Identify slow queries
  findSlowQueries(): Query[] {
    // Return queries > 100ms 95th percentile
  }

  // Detect sequential scans
  findMissingIndexes(): IndexRecommendation[] {
    // Look for queries with rowsScanned >> rowsReturned
  }
}
```

#### 3. ML Predictor

**Model**: LSTM for time-series prediction

**Training Data**:

- Historical query patterns (last 7 days)
- Day of week, hour of day features
- Workload seasonality

**Prediction**:

```typescript
// Predict query volume for next hour
const prediction = await model.predict({
  hour: 14,
  dayOfWeek: 3,
  historicalLoad: lastWeekData,
});

// Pre-warm cache if spike predicted
if (prediction.volume > threshold) {
  recommendCacheWarmup();
}
```

#### 4. Recommendation Engine

**Recommendation Types**:

1. **Index Recommendations**:

```typescript
{
  type: "CREATE_INDEX",
  table: "users",
  column: "email",
  reasoning: "80% of queries filter on this column",
  estimatedSpeedup: "50x",
  storageOverhead: "10MB"
}
```

2. **Cache Policy Recommendations**:

```typescript
{
  type: "ADJUST_CACHE_TTL",
  query: "SELECT * FROM products WHERE category = ?",
  newTTL: 600, // seconds
  reasoning: "High frequency, low mutation rate"
}
```

3. **Schema Recommendations**:

```typescript
{
  type: "ADD_COLUMN",
  table: "orders",
  suggestion: "Add denormalized user_email to avoid joins",
  estimatedSpeedup: "30%"
}
```

---

## Data Flow Patterns

### Pattern 1: Read Query Flow

```
1. Client sends SELECT query
   ↓
2. API Gateway authenticates & routes
   ↓
3. Parser converts to AST
   ↓
4. Planner generates execution plan
   ↓
5. Check Query Result Cache
   ├─ HIT → Return cached result
   └─ MISS → Continue
   ↓
6. Executor checks Buffer Pool
   ├─ Data in memory → Read
   └─ Data on disk → Load page to buffer pool
   ↓
7. Check indexes for efficient access
   ├─ Index exists → Use B-Tree/Hash lookup
   └─ No index → Sequential scan
   ↓
8. Return result to client
   ↓
9. [Async] Log query metrics to AI service
```

### Pattern 2: Write Query Flow

```
1. Client sends INSERT/UPDATE/DELETE
   ↓
2. Begin transaction (get TXN_ID)
   ↓
3. Acquire locks on affected rows
   ↓
4. Write to WAL (MUST complete before next step)
   ↓ [DISK FLUSH - Durability guaranteed]
5. Update in-memory buffer pool
   ↓
6. Mark pages as dirty
   ↓
7. Update indexes
   ↓
8. Release locks
   ↓
9. Commit transaction
   ↓
10. Return success to client
    ↓
11. [Background] Checkpoint dirty pages to disk
    ↓
12. [Async] Invalidate affected cache entries
    ↓
13. [Async] Log metrics to AI service
```

### Pattern 3: AI Optimization Flow

```
1. AI Service continuously reads query logs
   ↓
2. Aggregate metrics every 5 minutes
   ↓
3. Pattern Analyzer detects:
   - Frequent queries
   - Slow queries
   - Missing indexes
   ↓
4. ML Predictor forecasts workload
   ↓
5. Recommendation Engine generates suggestions
   ↓
6. Send recommendations to Go backend via gRPC
   ↓
7. Backend evaluates recommendations:
   ├─ Auto-apply safe changes (cache policies)
   └─ Queue manual approval (schema changes)
   ↓
8. Apply approved recommendations
   ↓
9. Monitor performance impact
   ↓
10. Feedback loop: Report results to AI service
```

---

## Communication Protocols

### Between Client ↔ Backend

**Protocol**: HTTP/REST or native TCP protocol

**REST Endpoints**:

```
POST   /api/v1/query          - Execute SQL query
POST   /api/v1/transaction    - Begin transaction
GET    /api/v1/status         - Health check
GET    /api/v1/metrics        - Prometheus metrics
```

### Between Backend ↔ AI Service

**Protocol**: gRPC (high performance, typed contracts)

**Service Definition**:

```protobuf
service AIOptimizer {
  rpc SubmitQueryLog(QueryLogBatch) returns (Ack);
  rpc GetRecommendations(Empty) returns (stream Recommendation);
  rpc ReportApplicationResult(Result) returns (Ack);
}
```

**Why gRPC?**:

- Binary protocol (faster than JSON)
- Streaming support (real-time logs)
- Strong typing (contract validation)

### Between Backend ↔ Monitoring

**Protocol**: Prometheus pull model

**Metrics Exposed**:

```
# Query performance
startdb_query_duration_seconds{type="select"} 0.023
startdb_query_duration_seconds{type="insert"} 0.045

# Cache statistics
startdb_cache_hit_ratio 0.85
startdb_cache_entries 15234

# Storage stats
startdb_disk_reads_total 45123
startdb_disk_writes_total 12456
```

---

## Scalability & Performance

### Horizontal Scaling Strategy

#### Read Replicas

```
           [Load Balancer]
                 │
     ┌───────────┼───────────┐
     ▼           ▼           ▼
 [Primary]   [Replica 1] [Replica 2]
     │           │           │
     └──── Async Replication ──┘
```

**Replication Strategy**:

- **Primary**: Handles all writes
- **Replicas**: Handle read queries
- **Replication lag**: Target < 100ms

#### Sharding (Future Phase)

```
Users with ID 1-1000     → Shard 1
Users with ID 1001-2000  → Shard 2
Users with ID 2001-3000  → Shard 3
```

**Sharding Key Selection**:

- Must be present in most queries
- Evenly distributed
- Example: user_id (good), timestamp (bad - hotspots)

### Performance Targets

| Metric              | Target   | Notes               |
| ------------------- | -------- | ------------------- |
| Query Latency (p50) | < 5ms    | Indexed reads       |
| Query Latency (p99) | < 50ms   | Complex queries     |
| Throughput          | 100K qps | Per instance        |
| Write Latency       | < 10ms   | Including WAL flush |
| Recovery Time       | < 30s    | From crash          |

---

## Failure Handling

### Failure Scenarios & Solutions

#### 1. Backend Crash

**Detection**: Health check failure (3 consecutive)
**Recovery**:

```go
1. Load last checkpoint
2. Replay WAL from checkpoint to crash point
3. Mark instance as healthy
4. Resume serving traffic
```

#### 2. AI Service Crash

**Impact**: Zero (backend continues normally)
**Recovery**: Recommendations paused, catch up when restarted

#### 3. Disk Failure

**Solution**: Replicate WAL to remote storage (S3)
**Recovery**: Restore from replica + remote WAL

#### 4. Network Partition

**Handling**:

- Primary can't reach replicas → Pause writes (avoid split-brain)
- Use consensus protocol (Raft) for leader election

---

## Security Considerations

### 1. Authentication

- JWT tokens with short expiration (15 min)
- Refresh tokens for long sessions
- API keys for programmatic access

### 2. Authorization

- Role-based access control (RBAC)
- Row-level security (future)

### 3. SQL Injection Prevention

- Parameterized queries only
- AST validation before execution

### 4. Encryption

- TLS for all network communication
- At-rest encryption for data files (AES-256)

### 5. Audit Logging

- Log all DDL operations
- Log authentication failures
- Retention: 90 days

---

## Technology Choices Justification

### Why Go for Core Engine?

✅ **Compiled** - Native performance
✅ **Goroutines** - Lightweight concurrency (1M+ connections possible)
✅ **Memory Safety** - Garbage collection without manual management
✅ **Simple Deployment** - Single binary
❌ **Con**: GC pauses (mitigated by tuning)

### Why TypeScript for AI Service?

✅ **Rich ML Ecosystem** - TensorFlow.js, ML.js, Brain.js
✅ **Async by Default** - Perfect for I/O-heavy ML tasks
✅ **Fast Iteration** - Quick experimentation
✅ **Type Safety** - Prevents runtime errors
❌ **Con**: Slower than Go (acceptable for non-critical path)

### Why BoltDB for Storage?

✅ **Embedded** - No separate process
✅ **ACID** - Built-in transaction support
✅ **Pure Go** - Easy integration
✅ **B-Tree based** - Good read/write balance
❌ **Con**: Single writer (future: move to RocksDB for multi-writer)

---

## Monitoring & Observability

### Metrics Dashboard

**Key Metrics**:

1. **Query Performance**

   - Latency percentiles (p50, p95, p99)
   - Throughput (qps)
   - Error rate

2. **Resource Utilization**

   - CPU usage per component
   - Memory usage
   - Disk I/O

3. **Cache Efficiency**

   - Hit/miss ratio
   - Eviction rate

4. **AI Insights**
   - Recommendations generated
   - Recommendations applied
   - Performance improvement

**Visualization**: Grafana dashboards

---

## Future Enhancements

### Phase 6: Distributed Database

- **Consensus**: Raft for leader election
- **Data Distribution**: Range-based sharding
- **Cross-shard Transactions**: Two-phase commit

### Phase 7: Advanced AI

- **NLP Query Interface**: "Show me all orders from last week"
- **Anomaly Detection**: Detect unusual query patterns
- **Auto-scaling**: Predict load and scale preemptively

### Phase 8: Compliance

- **GDPR**: Right to be forgotten
- **Audit Trails**: Complete operation history
- **Backup/Restore**: Point-in-time recovery

---

## Conclusion

StartDB's architecture balances:

- **Performance**: Go's speed + intelligent caching
- **Intelligence**: AI-driven optimization
- **Scalability**: Microservices + cloud-native design
- **Reliability**: WAL + MVCC + replication

This design allows StartDB to learn and adapt while maintaining the performance and reliability expected from a production database.

---

**Last Updated**: January 2025
**Version**: 1.0
**Author**: Bhupesh Kumar
