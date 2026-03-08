package io.github.mystisql.jdbc.examples;

import java.sql.*;

/**
 * Simple query example using MystiSql JDBC Driver.
 */
public class SimpleQuery {
    
    public static void main(String[] args) {
        // JDBC URL format: jdbc:mystisql://gateway-host:port/instance-name?params
        String url = "jdbc:mystisql://localhost:8080/production-mysql?timeout=60&ssl=false";
        String user = "your-username";
        String password = "your-token";  // Use token as password
        
        Connection conn = null;
        Statement stmt = null;
        ResultSet rs = null;
        
        try {
            // Load driver (optional with JDBC 4.0+, but good practice)
            Class.forName("io.github.mystisql.jdbc.MystiSqlDriver");
            
            // Connect to MystiSql Gateway
            System.out.println("Connecting to MystiSql Gateway...");
            conn = DriverManager.getConnection(url, user, password);
            System.out.println("Connected successfully!");
            
            // Create statement
            stmt = conn.createStatement();
            
            // Execute query
            String sql = "SELECT id, name, email FROM users LIMIT 10";
            System.out.println("Executing: " + sql);
            rs = stmt.executeQuery(sql);
            
            // Process results
            System.out.println("\nResults:");
            System.out.println("ID\tName\tEmail");
            System.out.println("---\t----\t-----");
            
            while (rs.next()) {
                int id = rs.getInt("id");
                String name = rs.getString("name");
                String email = rs.getString("email");
                
                System.out.printf("%d\t%s\t%s%n", id, name, email);
            }
            
            System.out.println("\nQuery completed!");
            
        } catch (ClassNotFoundException e) {
            System.err.println("Driver not found: " + e.getMessage());
            e.printStackTrace();
        } catch (SQLException e) {
            System.err.println("SQL error: " + e.getMessage());
            System.err.println("SQLState: " + e.getSQLState());
            e.printStackTrace();
        } finally {
            // Close resources
            try {
                if (rs != null) rs.close();
                if (stmt != null) stmt.close();
                if (conn != null) conn.close();
            } catch (SQLException e) {
                e.printStackTrace();
            }
        }
    }
}
