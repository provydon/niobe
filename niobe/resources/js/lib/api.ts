import { router } from '@inertiajs/vue3';
import axios from 'axios';

function getCsrfToken(): string {
    const meta = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
    if (meta) return meta;
    const match = document.cookie.match(/XSRF-TOKEN=([^;]+)/);
    return match ? decodeURIComponent(match[1]) : '';
}

const api = axios.create({
    baseURL: '/api',
    timeout: 60000,
    headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
    },
    withCredentials: true,
});

api.interceptors.request.use(
    (config) => {
        const token = getCsrfToken();
        if (token) {
            config.headers.set('X-XSRF-TOKEN', token);
        }
        config.headers.set('Accept', 'application/json');
        config.headers.set('X-Requested-With', 'XMLHttpRequest');
        if (config.data instanceof FormData) {
            config.headers.delete('Content-Type');
        }
        return config;
    },
    (error) => Promise.reject(error),
);

api.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.code === 'ECONNABORTED' || error.message === 'Network Error') {
            console.error('Request timeout or network error:', error);
            return Promise.reject(error);
        }

        if (error.response?.status === 401) {
            const path = window.location.pathname;
            const authPaths = ['/login', '/register', '/forgot-password', '/reset-password', '/verify-email'];
            const isAuthPage = authPaths.some((p) => path.startsWith(p));
            if (!isAuthPage) {
                router.visit('/login');
            }
        }

        if (error.response?.status === 419) {
            router.reload({ only: [] });
        }

        return Promise.reject(error);
    },
);

export default api;
