import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useState, useCallback } from 'react';
import { Card, Select, Button, Spin, Alert, Table, Typography, Empty, message, Space, Dropdown, Tabs, } from 'antd';
import { PlayCircleOutlined, DownloadOutlined, ClearOutlined, HistoryOutlined, CopyOutlined, } from '@ant-design/icons';
import Editor from '@monaco-editor/react';
import { useInstanceStore } from '../stores/instanceStore';
import { useQueryStore } from '../stores/queryStore';
import { apiClient } from '../api';
const { Text, Paragraph } = Typography;
export default function Query() {
    const { instances, currentInstance, setCurrentInstance } = useInstanceStore();
    const { currentSql, setCurrentSql, history, addHistory, clearHistory } = useQueryStore();
    const [loading, setLoading] = useState(false);
    const [result, setResult] = useState(null);
    const [error, setError] = useState(null);
    const [execTime, setExecTime] = useState(0);
    const handleExecute = useCallback(async () => {
        if (!currentInstance) {
            message.warning('请先选择实例');
            return;
        }
        if (!currentSql.trim()) {
            message.warning('请输入 SQL 语句');
            return;
        }
        setLoading(true);
        setError(null);
        setResult(null);
        try {
            const response = await apiClient.query({
                instance: currentInstance,
                sql: currentSql,
            });
            setExecTime(response.executionTime);
            if (response.success && response.data) {
                setResult(response.data);
                addHistory({
                    sql: currentSql,
                    instance: currentInstance,
                    success: true,
                });
                message.success(`查询成功，返回 ${response.data.rowCount} 行`);
            }
            else {
                setError(response.error?.message || '查询失败');
                addHistory({
                    sql: currentSql,
                    instance: currentInstance,
                    success: false,
                });
            }
        }
        catch (err) {
            const errorMsg = err instanceof Error ? err.message : '请求失败';
            setError(errorMsg);
            addHistory({
                sql: currentSql,
                instance: currentInstance,
                success: false,
            });
        }
        finally {
            setLoading(false);
        }
    }, [currentInstance, currentSql, addHistory]);
    const handleExport = useCallback((format) => {
        if (!result)
            return;
        let content;
        let filename;
        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
        if (format === 'json') {
            content = JSON.stringify(result.rows, null, 2);
            filename = `query-result-${timestamp}.json`;
        }
        else {
            const headers = result.columns.map((c) => c.name).join(',');
            const rows = result.rows
                .map((row) => result.columns
                .map((c) => {
                const val = row[c.name];
                if (val === null || val === undefined)
                    return '';
                if (typeof val === 'string' && val.includes(',')) {
                    return `"${val.replace(/"/g, '""')}"`;
                }
                return String(val);
            })
                .join(','))
                .join('\n');
            content = `${headers}\n${rows}`;
            filename = `query-result-${timestamp}.csv`;
        }
        const blob = new Blob([content], {
            type: format === 'json' ? 'application/json' : 'text/csv',
        });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        a.click();
        URL.revokeObjectURL(url);
        message.success('导出成功');
    }, [result]);
    const handleCopy = useCallback(() => {
        if (!result)
            return;
        navigator.clipboard.writeText(JSON.stringify(result.rows, null, 2));
        message.success('已复制到剪贴板');
    }, [result]);
    const handleHistoryClick = useCallback((sql) => {
        setCurrentSql(sql);
    }, [setCurrentSql]);
    const columns = result?.columns.map((col) => ({
        title: col.name,
        dataIndex: col.name,
        key: col.name,
        ellipsis: true,
        sorter: (a, b) => {
            const valA = a[col.name];
            const valB = b[col.name];
            if (valA === null || valA === undefined)
                return -1;
            if (valB === null || valB === undefined)
                return 1;
            if (typeof valA === 'number' && typeof valB === 'number') {
                return valA - valB;
            }
            return String(valA).localeCompare(String(valB));
        },
        render: (value) => {
            if (value === null)
                return _jsx(Text, { type: "secondary", children: "NULL" });
            if (value === undefined)
                return '';
            return String(value);
        },
    })) || [];
    const historyItems = history.slice(0, 20).map((item, index) => ({
        key: index,
        label: (_jsxs("div", { style: { maxWidth: 300 }, children: [_jsx(Text, { ellipsis: true, style: { display: 'block' }, children: item.sql }), _jsxs(Text, { type: "secondary", style: { fontSize: 12 }, children: [new Date(item.timestamp).toLocaleString(), " \u00B7 ", item.instance] })] })),
        onClick: () => handleHistoryClick(item.sql),
    }));
    return (_jsxs("div", { style: { height: '100%', display: 'flex', flexDirection: 'column' }, children: [_jsx(Card, { size: "small", style: { marginBottom: 16 }, children: _jsxs(Space, { children: [_jsx(Select, { style: { width: 200 }, placeholder: "\u9009\u62E9\u5B9E\u4F8B", value: currentInstance, onChange: setCurrentInstance, options: instances.map((i) => ({
                                label: i.name,
                                value: i.name,
                            })) }), _jsx(Button, { type: "primary", icon: _jsx(PlayCircleOutlined, {}), onClick: handleExecute, loading: loading, children: "\u6267\u884C (Ctrl+Enter)" }), _jsx(Dropdown, { menu: { items: historyItems }, placement: "bottomLeft", children: _jsxs(Button, { icon: _jsx(HistoryOutlined, {}), children: ["\u5386\u53F2\u8BB0\u5F55 (", history.length, ")"] }) }), _jsx(Button, { icon: _jsx(ClearOutlined, {}), onClick: () => clearHistory(), children: "\u6E05\u7A7A\u5386\u53F2" })] }) }), _jsxs("div", { style: { flex: 1, display: 'flex', gap: 16, minHeight: 0 }, children: [_jsx(Card, { size: "small", title: "SQL \u7F16\u8F91\u5668", style: { flex: 1, display: 'flex', flexDirection: 'column' }, styles: { body: { flex: 1, padding: 0 } }, children: _jsx(Editor, { height: "100%", defaultLanguage: "sql", value: currentSql, onChange: (value) => setCurrentSql(value || ''), theme: "vs", options: {
                                minimap: { enabled: false },
                                fontSize: 14,
                                lineNumbers: 'on',
                                wordWrap: 'on',
                                automaticLayout: true,
                                scrollBeyondLastLine: false,
                            }, onMount: (editor, monaco) => {
                                editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
                                    handleExecute();
                                });
                            } }) }), _jsxs(Card, { size: "small", title: _jsxs("span", { children: ["\u67E5\u8BE2\u7ED3\u679C", execTime > 0 && (_jsxs(Text, { type: "secondary", style: { marginLeft: 8, fontSize: 12 }, children: ["\u8017\u65F6: ", (execTime / 1000000).toFixed(2), "ms"] }))] }), extra: result && (_jsxs(Space, { children: [_jsx(Button, { size: "small", icon: _jsx(CopyOutlined, {}), onClick: handleCopy, children: "\u590D\u5236" }), _jsx(Dropdown, { menu: {
                                        items: [
                                            { key: 'csv', label: '导出 CSV', onClick: () => handleExport('csv') },
                                            { key: 'json', label: '导出 JSON', onClick: () => handleExport('json') },
                                        ],
                                    }, children: _jsx(Button, { size: "small", icon: _jsx(DownloadOutlined, {}), children: "\u5BFC\u51FA" }) })] })), style: { flex: 1, display: 'flex', flexDirection: 'column' }, styles: { body: { flex: 1, overflow: 'auto', padding: 12 } }, children: [loading && (_jsx("div", { style: { textAlign: 'center', padding: 50 }, children: _jsx(Spin, { size: "large" }) })), error && _jsx(Alert, { type: "error", message: error }), !loading && !error && !result && (_jsx(Empty, { description: "\u6267\u884C\u67E5\u8BE2\u4EE5\u67E5\u770B\u7ED3\u679C" })), !loading && !error && result && (_jsx(Tabs, { items: [
                                    {
                                        key: 'table',
                                        label: `表格 (${result.rowCount} 行)`,
                                        children: (_jsx(Table, { dataSource: result.rows, columns: columns, rowKey: (_, index) => `row-${index}`, size: "small", scroll: { x: 'max-content', y: 400 }, pagination: {
                                                showSizeChanger: true,
                                                showQuickJumper: true,
                                                showTotal: (total) => `共 ${total} 行`,
                                                defaultPageSize: 50,
                                                pageSizeOptions: ['20', '50', '100', '200'],
                                            } })),
                                    },
                                    {
                                        key: 'json',
                                        label: 'JSON',
                                        children: (_jsx(Paragraph, { children: _jsx("pre", { style: { margin: 0, maxHeight: 400, overflow: 'auto' }, children: JSON.stringify(result.rows, null, 2) }) })),
                                    },
                                ] }))] })] })] }));
}
