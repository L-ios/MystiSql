package io.github.mystisql.jdbc.client;

import java.util.List;

public class QueryResult {
    private List<ColumnInfo> columns;
    private List<List<Object>> rows;
    private Integer rowCount;
    private Long executionTimeNs;

    public List<ColumnInfo> getColumns() { return columns; }
    public void setColumns(List<ColumnInfo> columns) { this.columns = columns; }

    public List<List<Object>> getRows() { return rows; }
    public void setRows(List<List<Object>> rows) { this.rows = rows; }

    public Integer getRowCount() { return rowCount; }
    public void setRowCount(Integer rowCount) { this.rowCount = rowCount; }

    public Long getExecutionTimeNs() { return executionTimeNs; }
    public void setExecutionTimeNs(Long executionTimeNs) { this.executionTimeNs = executionTimeNs; }

    public static class ColumnInfo {
        private String name;
        private String type;

        public String getName() { return name; }
        public void setName(String name) { this.name = name; }

        public String getType() { return type; }
        public void setType(String type) { this.type = type; }
    }
}
