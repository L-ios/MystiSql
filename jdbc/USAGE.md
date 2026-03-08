# MystiSql JDBC Driver Usage Guide

Complete guide for using MystiSql JDBC Driver in your Java applications.

## Table of Contents

1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [Connection URL Format](#connection-url-format)
4. [Authentication](#authentication)
5. [Basic Operations](#basic-operations)
6. [PreparedStatement](#preparedstatement)
7. [Connection Pooling](#connection-pooling)
8. [IDE Tool Configuration](#ide-tool-configuration)
9. [Error Handling](#error-handling)
10. [Performance Optimization](#performance-optimization)
11. [Troubleshooting](#troubleshooting)

## Installation

### Option 1: Download JAR

Download the latest JAR from [Releases](https://github.com/mystisql/mystisql/releases) and add to your classpath.

### Option 2: Maven

```xml
<dependency>
    <groupId>io.github.mystisql</groupId>
    <artifactId>mystisql-jdbc</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Option 3: Gradle

```groovy
implementation 'io.github.mystisql:mystisql-jdbc:1.0.0'
```

## Quick Start

```java
import java.sql.*;

public class QuickStart {
    public static void main(String[] args) {
        String url = "jdbc:mystisql://gateway.example.com:8080/production-mysql";
        String user = "your-username";
        String password = "your-token";
        
        try (Connection conn = DriverManager.getConnection(url, user, password);
             Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery("SELECT * FROM users LIMIT 10")) {
            
            while (rs.next()) {
                System.out.println(rs.getString("name"));
            }
            
        } catch (SQLException e) {
            e.printStackTrace();
        }
    }
}
```

## Connection URL Format

```
jdbc:mystisql://gateway-host:port/instance-name?param1=value1&param2=value2
```

### Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `timeout` | int | 30 | Query timeout in seconds |
| `ssl` | boolean | false | Enable HTTPS |
| `verifySsl` | boolean | true | Verify SSL certificate |
| `token` | string | - | Authentication token |
| `maxConnections` | int | 20 | HTTP connection pool size |

### Examples

```java
// Basic connection
String url = "jdbc:mystisql://gateway:8080/instance";

// With SSL
String url = "jdbc:mystisql://gateway:8080/instance?ssl=true";

// With timeout and token
String url = "jdbc:mystisql://gateway:8080/instance?timeout=60&token=abc123";

// Disable SSL verification (development only!)
String url = "jdbc:mystisql://gateway:8080/instance?ssl=true&verifySsl=false";

// Custom connection pool size
String url = "jdbc:mystisql://gateway:8080/instance?maxConnections=50";
```

## Authentication

### Method 1: Token in URL

```java
String url = "jdbc:mystisql://gateway:8080/instance?token=your-api-token";
Connection conn = DriverManager.getConnection(url);
```

### Method 2: Password as Token

```java
String url = "jdbc:mystisql://gateway:8080/instance";
String user = "username";  // Optional
String password = "your-api-token";  // Token goes here

Connection conn = DriverManager.getConnection(url, user, password);
```

### Method 3: Properties

```java
String url = "jdbc:mystisql://gateway:8080/instance";
Properties props = new Properties();
props.setProperty("user", "username");
props.setProperty("password", "your-api-token");
props.setProperty("ssl", "true");

Connection conn = DriverManager.getConnection(url, props);
```

## Basic Operations

### SELECT Query

```java
try (Connection conn = DriverManager.getConnection(url, user, password);
     Statement stmt = conn.createStatement();
     ResultSet rs = stmt.executeQuery("SELECT id, name, email FROM users")) {
    
    while (rs.next()) {
        int id = rs.getInt("id");
        String name = rs.getString("name");
        String email = rs.getString("email");
        
        System.out.printf("ID: %d, Name: %s, Email: %s%n", id, name, email);
    }
}
```

### INSERT/UPDATE/DELETE

```java
try (Connection conn = DriverManager.getConnection(url, user, password);
     Statement stmt = conn.createStatement()) {
    
    // INSERT
    int rowsInserted = stmt.executeUpdate(
        "INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com')"
    );
    System.out.println("Rows inserted: " + rowsInserted);
    
    // UPDATE
    int rowsUpdated = stmt.executeUpdate(
        "UPDATE users SET email = 'newemail@example.com' WHERE id = 1"
    );
    System.out.println("Rows updated: " + rowsUpdated);
    
    // DELETE
    int rowsDeleted = stmt.executeUpdate("DELETE FROM users WHERE id = 1");
    System.out.println("Rows deleted: " + rowsDeleted);
}
```

### Get Generated Keys

```java
String sql = "INSERT INTO users (name, email) VALUES ('Bob', 'bob@example.com')";

try (Connection conn = DriverManager.getConnection(url, user, password);
     Statement stmt = conn.createStatement(Statement.RETURN_GENERATED_KEYS)) {
    
    stmt.executeUpdate(sql);
    
    ResultSet keys = stmt.getGeneratedKeys();
    if (keys.next()) {
        long id = keys.getLong(1);
        System.out.println("Generated ID: " + id);
    }
}
```

## PreparedStatement

**ALWAYS use PreparedStatement for user input to prevent SQL injection!**

### SELECT with Parameters

```java
String sql = "SELECT * FROM users WHERE age > ? AND status = ?";

try (Connection conn = DriverManager.getConnection(url, user, password);
     PreparedStatement stmt = conn.prepareStatement(sql)) {
    
    stmt.setInt(1, 18);
    stmt.setString(2, "active");
    
    ResultSet rs = stmt.executeQuery();
    
    while (rs.next()) {
        System.out.println(rs.getString("name"));
    }
    rs.close();
}
```

### INSERT with Parameters

```java
String sql = "INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)";

try (Connection conn = DriverManager.getConnection(url, user, password);
     PreparedStatement stmt = conn.prepareStatement(sql, Statement.RETURN_GENERATED_KEYS)) {
    
    stmt.setString(1, "John Doe");
    stmt.setString(2, "john@example.com");
    stmt.setInt(3, 25);
    stmt.setString(4, "active");
    
    int rowsInserted = stmt.executeUpdate();
    
    ResultSet keys = stmt.getGeneratedKeys();
    if (keys.next()) {
        System.out.println("Generated ID: " + keys.getLong(1));
    }
}
```

### Various Data Types

```java
PreparedStatement stmt = conn.prepareStatement(
    "INSERT INTO events (user_id, event_time, score, active, notes) VALUES (?, ?, ?, ?, ?)"
);

stmt.setInt(1, 100);
stmt.setTimestamp(2, new Timestamp(System.currentTimeMillis()));
stmt.setDouble(3, 98.5);
stmt.setBoolean(4, true);
stmt.setString(5, "Some notes");
stmt.setNull(6, Types.VARCHAR);  // NULL value

stmt.executeUpdate();
```

### Reuse PreparedStatement

```java
String sql = "UPDATE users SET last_login = ? WHERE id = ?";

try (PreparedStatement stmt = conn.prepareStatement(sql)) {
    
    // Update multiple users
    for (int userId : userIds) {
        stmt.setTimestamp(1, new Timestamp(System.currentTimeMillis()));
        stmt.setInt(2, userId);
        stmt.executeUpdate();
        
        stmt.clearParameters();  // Clear for next iteration
    }
}
```

## Connection Pooling

### HikariCP (Recommended)

Add dependency:
```xml
<dependency>
    <groupId>com.zaxxer</groupId>
    <artifactId>HikariCP</artifactId>
    <version>5.0.1</version>
</dependency>
```

Configure:
```java
import com.zaxxer.hikari.HikariConfig;
import com.zaxxer.hikari.HikariDataSource;

HikariConfig config = new HikariConfig();
config.setJdbcUrl("jdbc:mystisql://gateway:8080/instance");
config.setUsername("user");
config.setPassword("token");
config.setMaximumPoolSize(10);
config.setConnectionTimeout(30000);

HikariDataSource dataSource = new HikariDataSource(config);

// Use connection
try (Connection conn = dataSource.getConnection()) {
    // ... database operations
}

// Close pool when done
dataSource.close();
```

### Druid

```java
import com.alibaba.druid.pool.DruidDataSource;

DruidDataSource dataSource = new DruidDataSource();
dataSource.setUrl("jdbc:mystisql://gateway:8080/instance");
dataSource.setUsername("user");
dataSource.setPassword("token");
dataSource.setMaxActive(20);
dataSource.setInitialSize(5);

// Use connection...
```

## IDE Tool Configuration

### DataGrip

1. **Add Driver**
   - Open DataGrip
   - Go to Database → + → Driver and Files
   - Click "Add JAR" and select `mystisql-jdbc.jar`
   - Class: `io.github.mystisql.jdbc.MystiSqlDriver`

2. **Create Data Source**
   - URL: `jdbc:mystisql://gateway:8080/instance`
   - User: (any value, can be empty)
   - Password: Your API token
   - Click "Test Connection"

3. **Configure Options**
   - SSL: Add `?ssl=true` to URL
   - Timeout: Add `?timeout=60` to URL

### DBeaver

1. **Create New Connection**
   - Database → New Database Connection
   - Select "All" → "JDBC" (ODBC)
   
2. **Driver Settings**
   - Driver Class: `io.github.mystisql.jdbc.MystiSqlDriver`
   - URL Template: `jdbc:mystisql://{host}:{port}/{instance}`
   - Add JAR file

3. **Connection Settings**
   - Host: `gateway.example.com`
   - Port: `8080`
   - Database/Instance: `production-mysql`
   - Username: (optional)
   - Password: Your API token

### SQuirreL SQL

1. **Add Driver**
   - Open SQuirreL SQL
   - Go to Drivers → New Driver
   - Name: `MystiSql`
   - Example URL: `jdbc:mystisql://gateway:8080/instance`
   - Class Name: `io.github.mystisql.jdbc.MystiSqlDriver`
   - Add JAR

2. **Create Alias**
   - Aliases → New Alias
   - Select MystiSql driver
   - Fill in connection details

## Error Handling

### Catch and Handle SQLExceptions

```java
try {
    Statement stmt = conn.createStatement();
    ResultSet rs = stmt.executeQuery("SELECT * FROM nonexistent_table");
} catch (SQLException e) {
    // Error information
    String message = e.getMessage();        // Error message
    String sqlState = e.getSQLState();      // SQLState code
    int errorCode = e.getErrorCode();       // Vendor error code
    
    // Common SQLState codes
    switch (sqlState) {
        case "42S02":  // Table not found
            System.err.println("Table does not exist: " + message);
            break;
        case "42000":  // Syntax error
            System.err.println("SQL syntax error: " + message);
            break;
        case "08001":  // Connection failed
            System.err.println("Connection failed: " + message);
            break;
        case "HYT00":  // Timeout
            System.err.println("Query timeout: " + message);
            break;
        default:
            System.err.println("SQL error: " + message);
    }
}
```

### Retry Logic

```java
int maxRetries = 3;
int retryCount = 0;

while (retryCount < maxRetries) {
    try (Connection conn = dataSource.getConnection()) {
        // Execute query
        return executeQuery(conn);
        
    } catch (SQLException e) {
        retryCount++;
        
        if (retryCount >= maxRetries || !isRetryable(e)) {
            throw e;
        }
        
        // Wait before retry
        Thread.sleep(1000 * retryCount);
    }
}
```

## Performance Optimization

### 1. Use Connection Pooling

Always use connection pooling in production. HikariCP is recommended.

### 2. Reuse PreparedStatement

```java
// GOOD: Reuse PreparedStatement
PreparedStatement stmt = conn.prepareStatement("SELECT * FROM users WHERE id = ?");
for (int id : ids) {
    stmt.setInt(1, id);
    ResultSet rs = stmt.executeQuery();
    // process...
    rs.close();
    stmt.clearParameters();
}

// BAD: Create new statement each time
for (int id : ids) {
    PreparedStatement stmt = conn.prepareStatement("SELECT * FROM users WHERE id = " + id);
    // ...
}
```

### 3. Fetch Size

```java
// For large result sets
Statement stmt = conn.createStatement();
stmt.setFetchSize(100);  // Fetch 100 rows at a time
ResultSet rs = stmt.executeQuery("SELECT * FROM large_table");
```

### 4. Limit Result Size

```java
// Use LIMIT in SQL
String sql = "SELECT * FROM users LIMIT 1000";
```

### 5. Configure Timeout

```java
// Connection-level timeout
String url = "jdbc:mystisql://gateway:8080/instance?timeout=60";

// Statement-level timeout
Statement stmt = conn.createStatement();
stmt.setQueryTimeout(30);  // 30 seconds
```

## Troubleshooting

### Connection Failed

```
Error: Connection refused
```

**Solutions:**
1. Check Gateway is running: `curl http://gateway:8080/health`
2. Verify host and port are correct
3. Check firewall rules
4. Try telnet: `telnet gateway 8080`

### Authentication Failed

```
Error: Unauthorized (401)
```

**Solutions:**
1. Verify token is correct
2. Check token hasn't expired
3. Ensure token is passed correctly (as password or URL parameter)

### SSL Certificate Error

```
Error: PKIX path building failed
```

**Solutions:**
1. Use valid SSL certificate
2. Or disable SSL verification (dev only): `?ssl=true&verifySsl=false`

### Query Timeout

```
Error: Query timeout
```

**Solutions:**
1. Increase timeout: `?timeout=120`
2. Optimize query
3. Add indexes to tables

### Slow Performance

**Solutions:**
1. Use connection pooling
2. Enable HTTP keep-alive (default)
3. Increase connection pool size
4. Use PreparedStatement reuse
5. Check network latency to Gateway

### Driver Not Found

```
Error: No suitable driver found
```

**Solutions:**
1. Ensure JAR is in classpath
2. Explicitly load driver:
   ```java
   Class.forName("io.github.mystisql.jdbc.MystiSqlDriver");
   ```

## Getting Help

- **Documentation**: [https://mystisql.io/docs](https://mystisql.io/docs)
- **GitHub Issues**: [https://github.com/mystisql/mystisql/issues](https://github.com/mystisql/mystisql/issues)
- **Community Slack**: [https://mystisql.slack.com](https://mystisql.slack.com)

## License

Apache License 2.0
