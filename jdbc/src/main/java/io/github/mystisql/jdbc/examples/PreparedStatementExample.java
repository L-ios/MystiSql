package io.github.mystisql.jdbc.examples;

import java.sql.*;

/**
 * PreparedStatement example demonstrating parameterized queries.
 */
public class PreparedStatementExample {
    
    public static void main(String[] args) {
        String url = "jdbc:mystisql://localhost:8080/production-mysql?ssl=true";
        String user = "app-user";
        String password = "your-api-token";
        
        Connection conn = null;
        
        try {
            conn = DriverManager.getConnection(url, user, password);
            
            // Example 1: SELECT with parameters
            System.out.println("Example 1: SELECT with parameters");
            String selectSql = "SELECT id, name, age FROM users WHERE age > ? AND status = ?";
            
            try (PreparedStatement stmt = conn.prepareStatement(selectSql)) {
                stmt.setInt(1, 18);
                stmt.setString(2, "active");
                
                ResultSet rs = stmt.executeQuery();
                
                while (rs.next()) {
                    System.out.printf("ID: %d, Name: %s, Age: %d%n",
                        rs.getInt("id"),
                        rs.getString("name"),
                        rs.getInt("age"));
                }
                rs.close();
            }
            
            // Example 2: INSERT with parameters
            System.out.println("\nExample 2: INSERT with parameters");
            String insertSql = "INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)";
            
            try (PreparedStatement stmt = conn.prepareStatement(insertSql, Statement.RETURN_GENERATED_KEYS)) {
                stmt.setString(1, "John Doe");
                stmt.setString(2, "john@example.com");
                stmt.setInt(3, 25);
                stmt.setString(4, "active");
                
                int rowsAffected = stmt.executeUpdate();
                System.out.println("Rows inserted: " + rowsAffected);
                
                // Get generated keys
                ResultSet keys = stmt.getGeneratedKeys();
                if (keys.next()) {
                    System.out.println("Generated ID: " + keys.getLong(1));
                }
                keys.close();
            }
            
            // Example 3: UPDATE with parameters
            System.out.println("\nExample 3: UPDATE with parameters");
            String updateSql = "UPDATE users SET age = ?, status = ? WHERE id = ?";
            
            try (PreparedStatement stmt = conn.prepareStatement(updateSql)) {
                stmt.setInt(1, 26);
                stmt.setString(2, "inactive");
                stmt.setInt(3, 123);
                
                int rowsAffected = stmt.executeUpdate();
                System.out.println("Rows updated: " + rowsAffected);
            }
            
            // Example 4: DELETE with parameters
            System.out.println("\nExample 4: DELETE with parameters");
            String deleteSql = "DELETE FROM users WHERE status = ? AND age < ?";
            
            try (PreparedStatement stmt = conn.prepareStatement(deleteSql)) {
                stmt.setString(1, "inactive");
                stmt.setInt(2, 18);
                
                int rowsAffected = stmt.executeUpdate();
                System.out.println("Rows deleted: " + rowsAffected);
            }
            
            // Example 5: Multiple parameter types
            System.out.println("\nExample 5: Various parameter types");
            String complexSql = "INSERT INTO events (user_id, event_time, score, active, data) VALUES (?, ?, ?, ?, ?)";
            
            try (PreparedStatement stmt = conn.prepareStatement(complexSql)) {
                stmt.setInt(1, 100);
                stmt.setTimestamp(2, new Timestamp(System.currentTimeMillis()));
                stmt.setDouble(3, 98.5);
                stmt.setBoolean(4, true);
                stmt.setString(5, "{\"key\": \"value\"}");
                
                int rowsAffected = stmt.executeUpdate();
                System.out.println("Rows inserted: " + rowsAffected);
            }
            
            System.out.println("\nAll examples completed successfully!");
            
        } catch (SQLException e) {
            System.err.println("SQL Error:");
            System.err.println("  Message: " + e.getMessage());
            System.err.println("  SQLState: " + e.getSQLState());
            System.err.println("  ErrorCode: " + e.getErrorCode());
            e.printStackTrace();
        } finally {
            try {
                if (conn != null) conn.close();
            } catch (SQLException e) {
                e.printStackTrace();
            }
        }
    }
}
