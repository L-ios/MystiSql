import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useState } from 'react';
import { Form, Input, Button, Card, message, Typography } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useAuthStore } from '../stores/authStore';
import { apiClient } from '../api';
const { Title } = Typography;
export default function Login() {
    const [loading, setLoading] = useState(false);
    const setAuth = useAuthStore((state) => state.setAuth);
    const [form] = Form.useForm();
    const handleSubmit = async (values) => {
        setLoading(true);
        try {
            const response = await apiClient.login(values.userId, values.role);
            if (response.success && response.data) {
                const expiresAt = new Date(response.data.expiresAt).getTime();
                setAuth(response.data.token, response.data.userId, response.data.role, expiresAt);
                message.success('登录成功');
                window.location.href = '/dashboard';
            }
            else {
                message.error(response.error?.message || '登录失败');
            }
        }
        catch {
            message.error('登录请求失败，请检查网络连接');
        }
        finally {
            setLoading(false);
        }
    };
    return (_jsx("div", { style: {
            height: '100vh',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        }, children: _jsxs(Card, { style: { width: 400, boxShadow: '0 4px 12px rgba(0,0,0,0.15)' }, children: [_jsxs("div", { style: { textAlign: 'center', marginBottom: 24 }, children: [_jsx(Title, { level: 2, style: { marginBottom: 8 }, children: "MystiSql" }), _jsx(Typography.Text, { type: "secondary", children: "\u6570\u636E\u5E93\u8BBF\u95EE\u7F51\u5173" })] }), _jsxs(Form, { form: form, onFinish: handleSubmit, layout: "vertical", initialValues: { role: 'readonly' }, children: [_jsx(Form.Item, { name: "userId", label: "\u7528\u6237 ID", rules: [{ required: true, message: '请输入用户 ID' }], children: _jsx(Input, { prefix: _jsx(UserOutlined, {}), placeholder: "\u8BF7\u8F93\u5165\u7528\u6237 ID", size: "large" }) }), _jsx(Form.Item, { name: "role", label: "\u89D2\u8272", rules: [{ required: true, message: '请输入角色' }], children: _jsx(Input, { prefix: _jsx(LockOutlined, {}), placeholder: "\u8BF7\u8F93\u5165\u89D2\u8272\uFF08\u5982\uFF1Aadmin\u3001readonly\uFF09", size: "large" }) }), _jsx(Form.Item, { children: _jsx(Button, { type: "primary", htmlType: "submit", loading: loading, block: true, size: "large", children: "\u767B\u5F55" }) })] }), _jsx(Typography.Text, { type: "secondary", style: { display: 'block', textAlign: 'center' }, children: "\u4F7F\u7528 Token \u8BA4\u8BC1\u767B\u5F55" })] }) }));
}
