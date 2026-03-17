import { jsx as _jsx, jsxs as _jsxs, Fragment as _Fragment } from "react/jsx-runtime";
import { useEffect, useState } from 'react';
import { Card, Table, Tag, Button, Space, Typography, Modal, Descriptions, Spin, message, Badge, } from 'antd';
import { ReloadOutlined, EyeOutlined, CheckCircleOutlined, CloseCircleOutlined, } from '@ant-design/icons';
import { useInstanceStore } from '../stores/instanceStore';
import { apiClient } from '../api';
const { Title } = Typography;
export default function Instances() {
    const { instances, setInstances, loading, setLoading } = useInstanceStore();
    const [detailVisible, setDetailVisible] = useState(false);
    const [selectedInstance, setSelectedInstance] = useState(null);
    const [detailLoading, setDetailLoading] = useState(false);
    useEffect(() => {
        fetchInstances();
    }, []);
    const fetchInstances = async () => {
        setLoading(true);
        try {
            const response = await apiClient.getInstances();
            setInstances(response.instances);
        }
        catch {
            message.error('获取实例列表失败');
        }
        finally {
            setLoading(false);
        }
    };
    const showInstanceDetail = async (name) => {
        setDetailLoading(true);
        setDetailVisible(true);
        setSelectedInstance({ name, health: 'unknown' });
        try {
            const [healthRes, poolRes] = await Promise.all([
                apiClient.getInstanceHealth(name),
                apiClient.getPoolStats(name),
            ]);
            setSelectedInstance({
                name,
                health: healthRes.status,
                poolStats: poolRes.stats,
            });
        }
        catch {
            message.error('获取实例详情失败');
        }
        finally {
            setDetailLoading(false);
        }
    };
    const columns = [
        {
            title: '名称',
            dataIndex: 'name',
            key: 'name',
            render: (name) => _jsx(Typography.Text, { strong: true, children: name }),
        },
        {
            title: '类型',
            dataIndex: 'type',
            key: 'type',
            render: (type) => _jsx(Tag, { children: type.toUpperCase() }),
        },
        {
            title: '地址',
            key: 'address',
            render: (_, record) => `${record.host}:${record.port}`,
        },
        {
            title: '数据库',
            dataIndex: 'database',
            key: 'database',
            render: (db) => db || '-',
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            render: (status) => (_jsx(Badge, { status: status === 'healthy' ? 'success' : 'error', text: status === 'healthy' ? '健康' : '异常' })),
        },
        {
            title: '标签',
            dataIndex: 'labels',
            key: 'labels',
            render: (labels) => labels
                ? Object.entries(labels).map(([k, v]) => (_jsxs(Tag, { color: "blue", children: [k, "=", v] }, k)))
                : '-',
        },
        {
            title: '操作',
            key: 'action',
            render: (_, record) => (_jsx(Button, { type: "link", icon: _jsx(EyeOutlined, {}), onClick: () => showInstanceDetail(record.name), children: "\u8BE6\u60C5" })),
        },
    ];
    return (_jsxs("div", { children: [_jsxs("div", { style: { marginBottom: 16, display: 'flex', justifyContent: 'space-between' }, children: [_jsx(Title, { level: 4, style: { margin: 0 }, children: "\u5B9E\u4F8B\u7BA1\u7406" }), _jsx(Button, { icon: _jsx(ReloadOutlined, {}), onClick: fetchInstances, loading: loading, children: "\u5237\u65B0" })] }), _jsx(Card, { children: _jsx(Table, { dataSource: instances, columns: columns, rowKey: "name", loading: loading, pagination: {
                        showSizeChanger: true,
                        showTotal: (total) => `共 ${total} 个实例`,
                    } }) }), _jsx(Modal, { title: "\u5B9E\u4F8B\u8BE6\u60C5", open: detailVisible, onCancel: () => setDetailVisible(false), footer: null, width: 600, children: detailLoading ? (_jsx("div", { style: { textAlign: 'center', padding: 50 }, children: _jsx(Spin, {}) })) : selectedInstance ? (_jsxs(Space, { direction: "vertical", style: { width: '100%' }, size: "large", children: [_jsxs(Descriptions, { bordered: true, column: 2, size: "small", children: [_jsx(Descriptions.Item, { label: "\u5B9E\u4F8B\u540D\u79F0", span: 2, children: selectedInstance.name }), _jsx(Descriptions.Item, { label: "\u5065\u5EB7\u72B6\u6001", children: selectedInstance.health === 'healthy' ? (_jsx(Tag, { icon: _jsx(CheckCircleOutlined, {}), color: "success", children: "\u5065\u5EB7" })) : (_jsx(Tag, { icon: _jsx(CloseCircleOutlined, {}), color: "error", children: "\u5F02\u5E38" })) })] }), selectedInstance.poolStats && (_jsxs(_Fragment, { children: [_jsx(Title, { level: 5, children: "\u8FDE\u63A5\u6C60\u72B6\u6001" }), _jsxs(Descriptions, { bordered: true, column: 2, size: "small", children: [_jsx(Descriptions.Item, { label: "\u6700\u5927\u8FDE\u63A5\u6570", children: selectedInstance.poolStats.maxOpenConnections }), _jsx(Descriptions.Item, { label: "\u5F53\u524D\u8FDE\u63A5\u6570", children: selectedInstance.poolStats.openConnections }), _jsx(Descriptions.Item, { label: "\u4F7F\u7528\u4E2D", children: selectedInstance.poolStats.inUse }), _jsx(Descriptions.Item, { label: "\u7A7A\u95F2", children: selectedInstance.poolStats.idle }), _jsx(Descriptions.Item, { label: "\u7B49\u5F85\u6B21\u6570", children: selectedInstance.poolStats.waitCount }), _jsxs(Descriptions.Item, { label: "\u7B49\u5F85\u65F6\u95F4", children: [(selectedInstance.poolStats.waitDuration / 1000000).toFixed(2), "ms"] })] })] }))] })) : null })] }));
}
