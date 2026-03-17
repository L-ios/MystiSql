import { jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout as AntLayout, Menu, Dropdown, Avatar, Button, theme } from 'antd';
import { DatabaseOutlined, CodeOutlined, AuditOutlined, DashboardOutlined, UserOutlined, LogoutOutlined, MenuFoldOutlined, MenuUnfoldOutlined, } from '@ant-design/icons';
import { useAuthStore } from '../stores/authStore';
const { Header, Sider, Content } = AntLayout;
const menuItems = [
    { key: '/dashboard', icon: _jsx(DashboardOutlined, {}), label: '仪表盘' },
    { key: '/query', icon: _jsx(CodeOutlined, {}), label: 'SQL 查询' },
    { key: '/instances', icon: _jsx(DatabaseOutlined, {}), label: '实例管理' },
    { key: '/audit-logs', icon: _jsx(AuditOutlined, {}), label: '审计日志' },
];
export default function Layout() {
    const [collapsed, setCollapsed] = useState(false);
    const navigate = useNavigate();
    const location = useLocation();
    const { userId, role, clearAuth } = useAuthStore();
    const { token: { colorBgContainer, borderRadiusLG } } = theme.useToken();
    const handleMenuClick = ({ key }) => {
        navigate(key);
    };
    const handleLogout = () => {
        clearAuth();
        navigate('/login');
    };
    const userMenuItems = [
        {
            key: 'logout',
            icon: _jsx(LogoutOutlined, {}),
            label: '退出登录',
            onClick: handleLogout,
        },
    ];
    return (_jsxs(AntLayout, { style: { minHeight: '100vh' }, children: [_jsxs(Sider, { trigger: null, collapsible: true, collapsed: collapsed, children: [_jsx("div", { style: {
                            height: 64,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            color: '#fff',
                            fontSize: collapsed ? 16 : 20,
                            fontWeight: 'bold',
                        }, children: collapsed ? 'MS' : 'MystiSql' }), _jsx(Menu, { theme: "dark", mode: "inline", selectedKeys: [location.pathname], items: menuItems, onClick: handleMenuClick })] }), _jsxs(AntLayout, { children: [_jsxs(Header, { style: {
                            padding: '0 24px',
                            background: colorBgContainer,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'space-between',
                        }, children: [_jsx(Button, { type: "text", icon: collapsed ? _jsx(MenuUnfoldOutlined, {}) : _jsx(MenuFoldOutlined, {}), onClick: () => setCollapsed(!collapsed), style: { fontSize: 16, width: 64, height: 64 } }), _jsx(Dropdown, { menu: { items: userMenuItems }, placement: "bottomRight", children: _jsxs("div", { style: { cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }, children: [_jsx(Avatar, { icon: _jsx(UserOutlined, {}) }), _jsxs("span", { children: [userId, " (", role, ")"] })] }) })] }), _jsx(Content, { style: {
                            margin: 24,
                            padding: 24,
                            background: colorBgContainer,
                            borderRadius: borderRadiusLG,
                            minHeight: 280,
                            overflow: 'auto',
                        }, children: _jsx(Outlet, {}) })] })] }));
}
