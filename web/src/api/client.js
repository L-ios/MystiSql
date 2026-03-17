import axios from 'axios';
import { useAuthStore } from '../stores/authStore';
const API_BASE_URL = '/api/v1';
class ApiClient {
    constructor() {
        this.client = axios.create({
            baseURL: API_BASE_URL,
            timeout: 60000,
            headers: {
                'Content-Type': 'application/json',
            },
        });
        this.client.interceptors.request.use((config) => {
            const token = useAuthStore.getState().token;
            if (token) {
                config.headers.Authorization = `Bearer ${token}`;
            }
            return config;
        }, (error) => Promise.reject(error));
        this.client.interceptors.response.use((response) => response, (error) => {
            if (error.response?.status === 401) {
                useAuthStore.getState().clearAuth();
                window.location.href = '/login';
            }
            return Promise.reject(error);
        });
    }
    async login(userId, role) {
        const response = await this.client.post('/auth/token', { user_id: userId, role });
        return response.data;
    }
    async getTokenInfo(token) {
        const response = await this.client.get('/auth/token/info', { params: { token } });
        return response.data;
    }
    async getInstances() {
        const response = await this.client.get('/instances');
        return response.data;
    }
    async getInstanceHealth(name) {
        const response = await this.client.get(`/instances/${encodeURIComponent(name)}/health`);
        return response.data;
    }
    async getPoolStats(name) {
        const response = await this.client.get(`/instances/${encodeURIComponent(name)}/pool`);
        return response.data;
    }
    async query(request) {
        const response = await this.client.post('/query', request);
        return response.data;
    }
    async exec(request) {
        const response = await this.client.post('/exec', request);
        return response.data;
    }
    async getAuditLogs(params) {
        const response = await this.client.get('/audit/logs', {
            params,
        });
        return response.data;
    }
}
export const apiClient = new ApiClient();
