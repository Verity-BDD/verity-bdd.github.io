---
title: Ability Examples
description: Real-world examples of custom Ability implementations
---

Examples of Abilities for various testing scenarios.

## Database Ability (PostgreSQL)

### Interface

```go
package database

import (
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"

    "github.com/nchursin/verity-bdd/verity_abilities"
)

type DatabaseAbility interface {
    abilities.Ability

    Connect(dsn string) error
    Disconnect() error
    Ping() error

    Query(query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(query string, args ...interface{}) *sql.Row
    Execute(query string, args ...interface{}) (sql.Result, error)

    BeginTx() (*sql.Tx, error)

    LastQuery() string
    LastError() error
    IsConnected() bool
}
```

### Implementation

```go
type databaseAbility struct {
    db        *sql.DB
    lastQuery string
    lastError error
    dsn       string
    mutex     sync.RWMutex
}

func ConnectToPostgreSQL(dsn string) DatabaseAbility {
    return &databaseAbility{dsn: dsn}
}

func (d *databaseAbility) Connect(dsn string) error {
    d.mutex.Lock()
    defer d.mutex.Unlock()

    if dsn != "" {
        d.dsn = dsn
    }

    db, err := sql.Open("postgres", d.dsn)
    if err != nil {
        d.lastError = fmt.Errorf("failed to open database: %w", err)
        return d.lastError
    }

    if err := db.Ping(); err != nil {
        d.lastError = fmt.Errorf("failed to ping database: %w", err)
        db.Close()
        return d.lastError
    }

    d.db = db
    d.lastError = nil
    return nil
}

func (d *databaseAbility) Query(query string, args ...interface{}) (*sql.Rows, error) {
    d.mutex.Lock()
    defer d.mutex.Unlock()

    if d.db == nil {
        err := fmt.Errorf("database not connected")
        d.lastError = err
        d.lastQuery = query
        return nil, err
    }

    d.lastQuery = query
    rows, err := d.db.Query(query, args...)
    d.lastError = err
    return rows, err
}

func (d *databaseAbility) Execute(query string, args ...interface{}) (sql.Result, error) {
    d.mutex.Lock()
    defer d.mutex.Unlock()

    if d.db == nil {
        err := fmt.Errorf("database not connected")
        d.lastError = err
        d.lastQuery = query
        return nil, err
    }

    d.lastQuery = query
    result, err := d.db.Exec(query, args...)
    d.lastError = err
    return result, err
}
```

### Activities

```go
type CreateTableActivity struct {
    tableName string
    schema    string
}

func CreateTable(tableName, schema string) *CreateTableActivity {
    return &CreateTableActivity{tableName: tableName, schema: schema}
}

func (c *CreateTableActivity) PerformAs(actor core.Actor) error {
    ability, err := actor.AbilityTo(&databaseAbility{})
    if err != nil {
        return fmt.Errorf("actor does not have database ability: %w", err)
    }

    db := ability.(DatabaseAbility)
    _, err = db.Execute(c.schema)
    return err
}

func (c *CreateTableActivity) Description() string {
    return fmt.Sprintf("creates table: %s", c.tableName)
}
```

### Usage

```go
func TestDatabaseOperations(t *testing.T) {
    test := verity.NewVerityTest(t, verity.Scene{})

    actor := test.ActorCalled("DBAdmin").WhoCan(
        database.ConnectToPostgreSQL("postgres://user:pass@localhost/testdb?sslmode=disable"),
    )

    actor.AttemptsTo(
        core.Do("connects to database", func(actor core.Actor) error {
            ability, _ := actor.AbilityTo(&databaseAbility{})
            return ability.(DatabaseAbility).Connect("")
        }),
        database.CreateTable("users", `
            CREATE TABLE users (
                id SERIAL PRIMARY KEY,
                name VARCHAR(100) NOT NULL,
                email VARCHAR(255) UNIQUE NOT NULL,
                created_at TIMESTAMP DEFAULT NOW()
            )
        `),
        database.InsertInto("users", map[string]interface{}{
            "name":  "John Doe",
            "email": "john@example.com",
        }),
        ensure.That(database.RowCount("users"), expectations.Equals(1)),
    )
}
```

---

## FileSystem Ability

### Interface

```go
type FileSystemAbility interface {
    abilities.Ability

    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm fs.FileMode) error
    DeleteFile(path string) error
    Exists(path string) bool

    CreateDir(path string, perm fs.FileMode) error
    ListDir(path string) ([]fs.DirEntry, error)

    BackupFile(path string) (string, error)
    RestoreFile(backupPath string) error
    GetFileSize(path string) int64
    GetFileModTime(path string) time.Time

    SetWorkingDirectory(dir string) error
    GetWorkingDirectory() string

    LastOperation() string
}
```

### Implementation

```go
func ManageFileSystemIn(directory string) FileSystemAbility {
    abs, _ := filepath.Abs(directory)
    return &fileSystemAbility{
        workingDir: abs,
        backupDir:  filepath.Join(abs, ".backups"),
        backups:    make(map[string]string),
    }
}

func (f *fileSystemAbility) BackupFile(path string) (string, error) {
    f.mutex.Lock()
    defer f.mutex.Unlock()

    fullPath := filepath.Join(f.workingDir, path)

    if err := os.MkdirAll(f.backupDir, 0755); err != nil {
        return "", fmt.Errorf("failed to create backup directory: %w", err)
    }

    timestamp := time.Now().Format("20060102-150405")
    backupName := fmt.Sprintf("%s_%s", filepath.Base(path), timestamp)
    backupPath := filepath.Join(f.backupDir, backupName)

    if err := copyFile(fullPath, backupPath); err != nil {
        return "", fmt.Errorf("failed to backup file: %w", err)
    }

    f.backups[path] = backupPath
    return backupPath, nil
}
```

### Usage

```go
func TestFileSystemWithBackup(t *testing.T) {
    test := verity.NewVerityTest(t, verity.Scene{})
    tempDir := t.TempDir()

    actor := test.ActorCalled("FileEditor").WhoCan(
        filesystem.ManageFileSystemIn(tempDir),
    )

    actor.AttemptsTo(
        core.Do("creates original file", func(actor core.Actor) error {
            ability, _ := actor.AbilityTo(&fileSystemAbility{})
            return ability.(FileSystemAbility).WriteFile("important.txt", []byte("original content"), 0644)
        }),
        filesystem.BackupUpFile("important.txt"),
        core.Do("modifies file", func(actor core.Actor) error {
            ability, _ := actor.AbilityTo(&fileSystemAbility{})
            return ability.(FileSystemAbility).WriteFile("important.txt", []byte("modified content"), 0644)
        }),
        ensure.That(filesystem.FileContent("important.txt"), expectations.Equals("modified content")),
        filesystem.RestoreLastBackup("important.txt"),
        ensure.That(filesystem.FileContent("important.txt"), expectations.Equals("original content")),
    )
}
```

---

## WebSocket Ability

### Interface

```go
type WebSocketAbility interface {
    abilities.Ability

    Connect(url string, header http.Header) error
    Disconnect() error
    IsConnected() bool

    Send(message []byte) error
    SendJSON(v interface{}) error
    Receive(timeout time.Duration) ([]byte, error)
    ReceiveJSON(v interface{}, timeout time.Duration) error

    LastMessage() []byte
    ConnectionDuration() time.Duration
    MessageCount() int
}
```

### Usage

```go
func TestWebSocketChat(t *testing.T) {
    // Start a test WebSocket echo server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        upgrader := websocket.Upgrader{}
        conn, _ := upgrader.Upgrade(w, r, nil)
        defer conn.Close()

        for {
            messageType, message, err := conn.ReadMessage()
            if err != nil {
                break
            }
            conn.WriteMessage(messageType, message)
        }
    }))
    defer server.Close()

    wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

    actor := test.ActorCalled("WebSocketClient").WhoCan(
        websocket.ConnectToWebSocket(),
    )

    actor.AttemptsTo(
        core.Do("connects to websocket", func(actor core.Actor) error {
            ability, _ := actor.AbilityTo(&webSocketAbility{})
            return ability.(WebSocketAbility).Connect(wsURL, nil)
        }),
        core.Do("sends message", func(actor core.Actor) error {
            ability, _ := actor.AbilityTo(&webSocketAbility{})
            return ability.(WebSocketAbility).Send([]byte("Hello WebSocket!"))
        }),
        core.Do("receives response", func(actor core.Actor) error {
            ability, _ := actor.AbilityTo(&webSocketAbility{})
            _, err := ability.(WebSocketAbility).Receive(5 * time.Second)
            return err
        }),
        ensure.That(websocket.LastMessage(), expectations.Equals([]byte("Hello WebSocket!"))),
    )
}
```

---

## Redis Ability

### Interface

```go
type RedisAbility interface {
    abilities.Ability

    Connect(addr string, options *redis.Options) error
    Disconnect() error
    Ping() error

    Set(key string, value interface{}, expiration time.Duration) error
    Get(key string) (string, error)
    Del(keys ...string) error
    Exists(keys ...string) (int64, error)

    HSet(key string, values ...interface{}) error
    HGet(key, field string) (string, error)
    HGetAll(key string) (map[string]string, error)

    LPush(key string, values ...interface{}) error
    RPop(key string) (string, error)
    LRange(key string, start, stop int64) ([]string, error)
}
```

### Usage

```go
func TestRedisOperations(t *testing.T) {
    test := verity.NewVerityTest(t, verity.Scene{})

    actor := test.ActorCalled("RedisUser").WhoCan(
        redis.ConnectToRedis("localhost:6379"),
    )

    actor.AttemptsTo(
        core.Do("sets key-value", func(actor core.Actor) error {
            ability, _ := actor.AbilityTo(&redisAbility{})
            return ability.(RedisAbility).Set("test:key", "test-value", 0)
        }),
        ensure.That(redis.KeyExists("test:key"), expectations.IsTrue()),
        ensure.That(redis.StringValue("test:key"), expectations.Equals("test-value")),
        core.Do("deletes key", func(actor core.Actor) error {
            ability, _ := actor.AbilityTo(&redisAbility{})
            return ability.(RedisAbility).Del("test:key")
        }),
        ensure.That(redis.KeyExists("test:key"), expectations.IsFalse()),
    )
}
```
