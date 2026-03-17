import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useState, useEffect } from 'react';
import { Card, Table, Form, Input, DatePicker, Button, Space, Typography, Tag, message, } from 'antd';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { apiClient } from '../api';
const { Title } = Typography;
const { RangePicker } = DatePicker;
export default function AuditLogs() {
    const [loading, setLoading] = useState(false);
    const [logs, setLogs] = useState([]);
    const [total, setTotal] = useState(0);
    const [form] = Form.useForm();
    const [pagination, setPagination] = useState({ current: 1, pageSize: 20 });
    useEffect(() => {
        fetchLogs();
    }, [pagination]);
    const fetchLogs = async (params) => {
        setLoading(true);
        try {
            const response = await apiClient.getAuditLogs({
                ...params,
                page: pagination.current,
                pageSize: pagination.pageSize,
            });
            if (response.success && response.data) {
                setLogs(response.data.logs);
                setTotal(response.data.total);
            }
            else {
                message.error(response.error?.message || '获取审计日志失败');
            }
        }
        catch {
            message.error('请求失败');
        }
        finally {
            setLoading(false);
        }
    };
    const handleSearch = () => {
        const values = form.getFieldsValue();
        const params = {};
        if (values.userId)
            params.userId = values.userId;
        if (values.instance)
            params.instance = values.instance;
        if (values.timeRange && values.timeRange[0] && values.timeRange[1]) {
            params.startTime = values.timeRange[0].toISOString();
            params.endTime = values.timeRange[1].toISOString();
        }
        setPagination({ ...pagination, current: 1 });
        fetchLogs(params);
    };
    const handleReset = () => {
        form.resetFields();
        setPagination({ ...pagination, current: 1 });
        fetchLogs();
    };
    const columns = [
        {
            title: '时间',
            dataIndex: 'timestamp',
            key: 'timestamp',
            width: 180,
            render: (ts) => dayjs(ts).format('YYYY-MM-DD HH:mm:ss'),
        },
        {
            title: '用户',
            dataIndex: 'userId',
            key: 'userId',
            width: 120,
        },
        {
            title: '实例',
            dataIndex: 'instance',
            key: 'instance',
            width: 120,
        },
        {
            title: 'SQL',
            dataIndex: 'sql',
            key: 'sql',
            ellipsis: true,
            render: (sql) => (_jsx(Typography.Text, { style: { fontFamily: 'monospace', fontSize: 12 }, ellipsis: { tooltip: sql }, children: sql })),
        },
        {
            title: '耗时',
            dataIndex: 'executionTime',
            key: 'executionTime',
            width: 100,
            render: (time) => `${(time / 1000000).toFixed(2)}ms`,
        },
        {
            title: '影响行数',
            dataIndex: 'rowsAffected',
            key: 'rowsAffected',
            width: 100,
        },
        {
            title: '状态',
            dataIndex: 'success',
            key: 'success',
            width: 80,
            render: (success) => (_jsx(Tag, { color: success ? 'success' : 'error', children: success ? '成功' : '失败' })),
        },
        {
            title: '错误信息',
            dataIndex: 'errorMessage',
            key: 'errorMessage',
            ellipsis: true,
            render: (msg) => msg ? (_jsx(Typography.Text, { type: "danger", ellipsis: { tooltip: msg }, children: msg })) : ('-'),
        },
    ];
    return (_jsxs("div", { children: [_jsx(Title, { level: 4, style: { marginBottom: 16 }, children: "\u5BA1\u8BA1\u65E5\u5FD7" }), _jsx(Card, { size: "small", style: { marginBottom: 16 }, children: _jsxs(Form, { form: form, layout: "inline", children: [_jsx(Form.Item, { name: "userId", label: "\u7528\u6237 ID", children: _jsx(Input, { placeholder: "\u8F93\u5165\u7528\u6237 ID", style: { width: 150 } }) }), _jsx(Form.Item, { name: "instance", label: "\u5B9E\u4F8B", children: _jsx(Input, { placeholder: "\u8F93\u5165\u5B9E\u4F8B\u540D", style: { width: 150 } }) }), _jsx(Form.Item, { name: "timeRange", label: "\u65F6\u95F4\u8303\u56F4", children: _jsx(RangePicker, { showTime: true, style: { width: 350 } }) }), _jsx(Form.Item, { children: _jsxs(Space, { children: [_jsx(Button, { type: "primary", icon: _jsx(SearchOutlined, {}), onClick: handleSearch, children: "\u641C\u7D22" }), _jsx(Button, { icon: _jsx(ReloadOutlined, {}), onClick: handleReset, children: "\u91CD\u7F6E" })] }) })] }) }), _jsx(Card, { children: _jsx(Table, { dataSource: logs, columns: columns, rowKey: "id", loading: loading, pagination: {
                        ...pagination,
                        total,
                        showSizeChanger: true,
                        showQuickJumper: true,
                        showTotal: (t) => `共 ${t} 条记录`,
                        onChange: (page, pageSize) => setPagination({ current: page, pageSize }),
                    }, scroll: { x: 1200 } }) })] }));
}
