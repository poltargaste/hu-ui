import React, { useEffect, useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, createTheme, CssBaseline, Box } from '@mui/material';
import axios from 'axios';

import { AuthProvider, useAuth } from './context/AuthContext';
import Sidebar from './components/Sidebar';
import Login from './pages/Login';
import Users from './pages/Users';
import Settings from './pages/Settings';

// Создание премиум темной темы MUI
const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: '#b388ff', // Светло-фиолетовый акцент
    },
    secondary: {
      main: '#ff4081', // Розовый
    },
    background: {
      default: '#0a0813', // Глубокий темный фон
      paper: '#121020',   // Цвет карточек
    },
    text: {
      primary: '#e0e0e0',
      secondary: '#9e9e9e',
    },
  },
  shape: {
    borderRadius: 12,
  },
  typography: {
    fontFamily: '"Outfit", "Inter", "Roboto", "Helvetica", "Arial", sans-serif',
    button: {
      textTransform: 'none',
      fontWeight: 'bold',
    },
  },
  components: {
    MuiCard: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
          boxShadow: '0 4px 20px 0 rgba(0,0,0,0.15)',
        },
      },
    },
    MuiTableCell: {
      styleOverrides: {
        root: {
          borderBottom: '1px solid rgba(255, 255, 255, 0.05)',
        },
      },
    },
  },
});

// Защищенный роут с макетом
const ProtectedLayout = () => {
  const { isAuthenticated, logout } = useAuth();
  const [systemStats, setSystemStats] = useState(null);

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Опрашиваем статистику системы для отображения в сайдбаре и на страницах
  const refreshStats = async () => {
    try {
      const response = await axios.get('/api/system/stats');
      setSystemStats(response.data);
    } catch (error) {
      console.error('Failed to get system stats', error);
      if (error.response && error.response.status === 401) {
        logout();
      }
    }
  };

  useEffect(() => {
    refreshStats();
    // Опрашиваем статус раз в 10 секунд
    const interval = setInterval(refreshStats, 10000);
    return () => clearInterval(interval);
  }, []);

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <Sidebar systemStats={systemStats} />
      <Box component="main" sx={{ flexGrow: 1, overflowX: 'auto', background: '#0a0813' }}>
        <Routes>
          <Route path="/" element={<Users systemStats={systemStats} refreshStats={refreshStats} />} />
          <Route path="/settings" element={<Settings systemStats={systemStats} refreshStats={refreshStats} />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Box>
    </Box>
  );
};

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/*" element={<ProtectedLayout />} />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;
