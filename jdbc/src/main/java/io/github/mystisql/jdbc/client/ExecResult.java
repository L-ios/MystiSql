package io.github.mystisql.jdbc.client;

public class ExecResult {
    private Long rowsAffected;
    private Long lastInsertId;
    private Long executionTimeNs;

    public Long getRowsAffected() { return rowsAffected; }
    public void setRowsAffected(Long rowsAffected) { this.rowsAffected = rowsAffected; }

    public Long getLastInsertId() { return lastInsertId; }
    public void setLastInsertId(Long lastInsertId) { this.lastInsertId = lastInsertId; }

    public Long getExecutionTimeNs() { return executionTimeNs; }
    public void setExecutionTimeNs(Long executionTimeNs) { this.executionTimeNs = executionTimeNs; }
}
