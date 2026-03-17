import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useEffect } from 'react';
import { Card, Row, Col, Statistic, Typography, Spin, Tag } from 'antd';
import { DatabaseOutlined, CheckCircleOutlined, CloseCircleOutlined, ClockCircleOutlined, } from '@ant-design/icons';
import { useInstanceStore } from '../stores/instanceStore';
import { apiClient } from '../api';
const { Title } = Typography;
export default function Dashboard() {
    const { instances, setInstances, loading, setLoading } = useInstanceStore();
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
            console.error('Failed to fetch instances');
        }
        finally {
            setLoading(false);
        }
    };
    const healthyCount = instances.filter((i) => i.status === 'healthy').length;
    const unhealthyCount = instances.filter((i) => i.status !== 'healthy').length;
    if (loading) {
        return (_jsx("div", { style: { textAlign: 'center', padding: 100 }, children: _jsx(Spin, { size: "large" }) }));
    }
    return (_jsxs("div", { children: [_jsx(Title, { level: 4, style: { marginBottom: 24 }, children: "\u4EEA\u8868\u76D8" }), _jsxs(Row, { gutter: 16, children: [_jsx(Col, { span: 6, children: _jsx(Card, { children: _jsx(Statistic, { title: "\u603B\u5B9E\u4F8B\u6570", value: instances.length, prefix: _jsx(DatabaseOutlined, {}) }) }) }), _jsx(Col, { span: 6, children: _jsx(Card, { children: _jsx(Statistic, { title: "\u5065\u5EB7\u5B9E\u4F8B", value: healthyCount, prefix: _jsx(CheckCircleOutlined, {}), valueStyle: { color: '#3f8600' } }) }) }), _jsx(Col, { span: 6, children: _jsx(Card, { children: _jsx(Statistic, { title: "\u5F02\u5E38\u5B9E\u4F8B", value: unhealthyCount, prefix: _jsx(CloseCircleOutlined, {}), valueStyle: { color: unhealthyCount > 0 ? '#cf1322' : '#52c41a' } }) }) }), _jsx(Col, { span: 6, children: _jsx(Card, { children: _jsx(Statistic, { title: "\u5728\u7EBF\u65F6\u957F", value: 0, suffix: "\u5206\u949F", prefix: _jsx(ClockCircleOutlined, {}) }) }) })] }), _jsx(Title, { level: 5, style: { marginTop: 32, marginBottom: 16 }, children: "\u5B9E\u4F8B\u72B6\u6001\u6982\u89C8" }), _jsx(Row, { gutter: [16, 16], children: instances.map((instance) => (_jsx(Col, { span: 8, children: _jsx(Card, { size: "small", children: _jsxs("div", { style: { display: 'flex', justifyContent: 'space-between', alignItems: 'center' }, children: [_jsxs("div", { children: [_jsx(Typography.Text, { strong: true, children: instance.name }), _jsx("br", {}), _jsxs(Typography.Text, { type: "secondary", children: [instance.type, " \u00B7 ", instance.host, ":", instance.port] })] }), _jsx(Tag, { color: instance.status === 'healthy' ? 'success' : 'error', icon: instance.status === 'healthy' ? (_jsx(CheckCircleOutlined, {})) : (_jsx(CloseCircleOutlined, {})), children: instance.status })] }) }) }, instance.name))) })] }));
}
