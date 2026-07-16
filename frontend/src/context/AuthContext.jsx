import React, { createContext, useContext, useState, useEffect } from 'react';
import axios from 'axios';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [token, setToken] = useState(localStorage.getItem('token') || '');
  const [isAuthenticated, setIsAuthenticated] = useState(!!localStorage.getItem('token'));
  const [loading, setLoading] = useState(false);

  // Настройка базового URL для Axios
  // В продакшене запросы идут на тот же хост, в деве — на localhost:8080 (прокси или прямой урл)
  const API_URL = window.location.origin.includes('localhost:5173') 
    ? 'http://localhost:8080' 
    : window.location.origin;

  useEffect(() => {
    if (token) {
      localStorage.setItem('token', token);
      setIsAuthenticated(true);
    } else {
      localStorage.removeItem('token');
      setIsAuthenticated(false);
    }
  }, [token]);

  // Axios Interceptor для автоматического добавления JWT токена
  useEffect(() => {
    const requestInterceptor = axios.interceptors.request.use(
      (config) => {
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        config.baseURL = API_URL;
        return config;
      },
      (error) => Promise.reject(error)
    );

    const responseInterceptor = axios.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response && error.response.status === 401) {
          // Если токен протух или невалиден, разлогиниваем
          setToken('');
        }
        return Promise.reject(error);
      }
    );

    return () => {
      axios.interceptors.request.eject(requestInterceptor);
      axios.interceptors.response.eject(responseInterceptor);
    };
  }, [token, API_URL]);

  const login = async (username, password) => {
    setLoading(true);
    try {
      const response = await axios.post('/api/auth/login', { username, password });
      setToken(response.data.token);
      setLoading(false);
      return { success: true };
    } catch (error) {
      setLoading(false);
      const message = error.response?.data?.error || 'Authorization failed';
      return { success: false, error: message };
    }
  };

  const logout = () => {
    setToken('');
  };

  return (
    <AuthContext.Provider value={{ token, isAuthenticated, loading, login, logout, API_URL }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);
