# SQL Server Service

This package provides a service for connecting to and executing queries against Microsoft SQL Server.

## Usage

### Initialization

```go
import "idongivaflyinfa/service"
import "idongivaflyinfa/config"

cfg := config.SQLServerConfig{
    Server:   "localhost",
    Port:     "1433",
    Database: "mydb",
    UserID:   "sa",
    Password: "password",
    Encrypt:  true,
}

sqlService, err := service.NewSQLServerService(cfg)
if err != nil {
    log.Fatal(err)
}
defer sqlService.Close()
```

### Execute Query (SELECT)

```go
result, err := sqlService.ExecuteQuery("SELECT * FROM users WHERE id = 1")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Columns:", result.Columns)
for _, row := range result.Rows {
    fmt.Println("Row:", row)
}
```

### Execute Non-Query (INSERT, UPDATE, DELETE)

```go
rowsAffected, err := sqlService.ExecuteNonQuery("UPDATE users SET name = 'John' WHERE id = 1")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Rows affected:", rowsAffected)
```

### Check Connection

```go
if sqlService.IsConnected() {
    fmt.Println("SQL Server is connected")
}
```

## Methods

- `NewSQLServerService(cfg SQLServerConfig) (*SQLServerService, error)` - Create a new SQL Server service
- `Close() error` - Close the database connection
- `ExecuteQuery(query string) (*models.SQLResult, error)` - Execute a SELECT query and return results
- `ExecuteNonQuery(query string) (int64, error)` - Execute INSERT/UPDATE/DELETE and return rows affected
- `IsConnected() bool` - Check if the connection is active

## Connection String Format

The service builds connection strings in the following format:
- `server=hostname;port=1433;database=dbname;user id=username;password=pass;encrypt=true`
- For Windows Authentication: `server=hostname;port=1433;database=dbname;trusted_connection=true;encrypt=true`

