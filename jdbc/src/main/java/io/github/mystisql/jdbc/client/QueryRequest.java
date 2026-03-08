package io.github.mystisql.jdbc.client;

import java.util.ArrayList;
import java.util.List;

public class QueryRequest {
    private String instance;
    private String query;
    private List<QueryParameter> parameters;
    private Integer timeout;

    public QueryRequest() {
        this.parameters = new ArrayList<>();
    }

    public String getInstance() { return instance; }
    public void setInstance(String instance) { this.instance = instance; }

    public String getQuery() { return query; }
    public void setQuery(String query) { this.query = query; }

    public List<QueryParameter> getParameters() { return parameters; }
    public void setParameters(List<QueryParameter> parameters) { this.parameters = parameters; }

    public Integer getTimeout() { return timeout; }
    public void setTimeout(Integer timeout) { this.timeout = timeout; }
}
