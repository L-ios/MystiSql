import { Fragment as _Fragment, jsx as _jsx, jsxs as _jsxs } from "react/jsx-runtime";
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from './stores/authStore';
import Layout from './components/Layout';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Query from './pages/Query';
import Instances from './pages/Instances';
import AuditLogs from './pages/AuditLogs';
function PrivateRoute({ children }) {
    const { token } = useAuthStore();
    return token ? _jsx(_Fragment, { children: children }) : _jsx(Navigate, { to: "/login", replace: true });
}
function App() {
    return (_jsx(BrowserRouter, { children: _jsxs(Routes, { children: [_jsx(Route, { path: "/login", element: _jsx(Login, {}) }), _jsxs(Route, { path: "/", element: _jsx(PrivateRoute, { children: _jsx(Layout, {}) }), children: [_jsx(Route, { index: true, element: _jsx(Navigate, { to: "/dashboard", replace: true }) }), _jsx(Route, { path: "dashboard", element: _jsx(Dashboard, {}) }), _jsx(Route, { path: "query", element: _jsx(Query, {}) }), _jsx(Route, { path: "instances", element: _jsx(Instances, {}) }), _jsx(Route, { path: "audit-logs", element: _jsx(AuditLogs, {}) })] }), _jsx(Route, { path: "*", element: _jsx(Navigate, { to: "/dashboard", replace: true }) })] }) }));
}
export default App;
