import React from 'react';
import { 
  Box, 
  Drawer, 
  List, 
  ListItem, 
  ListItemButton, 
  ListItemIcon, 
  ListItemText, 
  Typography, 
  Divider, 
  IconButton,
  Tooltip
} from '@mui/material';
import { 
  People, 
  Settings, 
  Logout, 
  Dns,
  FiberManualRecord
} from '@mui/icons-material';
import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const drawerWidth = 260;

const Sidebar = ({ systemStats }) => {
  const { logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  const isRunning = systemStats?.hysteria_running || false;

  const menuItems = [
    { text: 'VPN Users', icon: <People />, path: '/' },
    { text: 'Settings & Core', icon: <Settings />, path: '/settings' },
  ];

  return (
    <Drawer
      variant="permanent"
      sx={{
        width: drawerWidth,
        flexShrink: 0,
        [`& .MuiDrawer-paper`]: { 
          width: drawerWidth, 
          boxSizing: 'border-box',
          background: '#0d0b18',
          borderRight: '1px solid rgba(255, 255, 255, 0.05)',
          color: '#fff',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'space-between'
        },
      }}
    >
      <Box>
        {/* Заголовок */}
        <Box sx={{ p: 3, display: 'flex', alignItems: 'center', gap: 1.5 }}>
          <Box
            sx={{
              p: 1,
              borderRadius: '50%',
              background: 'linear-gradient(135deg, #651fff, #8e24aa)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              boxShadow: '0 0 10px rgba(101, 31, 255, 0.4)'
            }}
          >
            <Dns sx={{ fontSize: 20, color: '#fff' }} />
          </Box>
          <Typography variant="h6" fontWeight="bold" letterSpacing={0.5}>
            HYSTERIA 2
          </Typography>
        </Box>

        <Divider sx={{ borderColor: 'rgba(255, 255, 255, 0.05)' }} />

        {/* Меню навигации */}
        <List sx={{ px: 1.5, py: 2 }}>
          {menuItems.map((item) => {
            const active = location.pathname === item.path;
            return (
              <ListItem key={item.text} disablePadding sx={{ mb: 0.5 }}>
                <ListItemButton
                  onClick={() => navigate(item.path)}
                  sx={{
                    borderRadius: 2,
                    py: 1.2,
                    background: active ? 'rgba(101, 31, 255, 0.15)' : 'transparent',
                    color: active ? '#b388ff' : '#9fa8da',
                    '&:hover': {
                      background: 'rgba(255, 255, 255, 0.03)',
                      color: '#fff',
                      '& .MuiListItemIcon-root': { color: '#fff' }
                    },
                    '& .MuiListItemIcon-root': {
                      color: active ? '#b388ff' : '#7986cb',
                      minWidth: 40
                    }
                  }}
                >
                  <ListItemIcon>{item.icon}</ListItemIcon>
                  <ListItemText 
                    primary={item.text} 
                    primaryTypographyProps={{ fontSize: 14, fontWeight: active ? 'bold' : 'medium' }} 
                  />
                </ListItemButton>
              </ListItem>
            );
          })}
        </List>
      </Box>

      {/* Футер сайдбара с индикатором и кнопкой выхода */}
      <Box sx={{ p: 2 }}>
        <Divider sx={{ borderColor: 'rgba(255, 255, 255, 0.05)', mb: 2 }} />
        
        <Box sx={{ 
          display: 'flex', 
          alignItems: 'center', 
          justifyContent: 'space-between',
          background: 'rgba(255, 255, 255, 0.02)',
          borderRadius: 2,
          p: 1.5,
          border: '1px solid rgba(255, 255, 255, 0.03)'
        }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Tooltip title={isRunning ? "Core is running" : "Core is stopped"}>
              <FiberManualRecord 
                sx={{ 
                  color: isRunning ? '#00e676' : '#ff1744', 
                  fontSize: 14,
                  animation: isRunning ? 'pulse 2s infinite' : 'none',
                  '@keyframes pulse': {
                    '0%': { transform: 'scale(0.95)', opacity: 0.5 },
                    '50%': { transform: 'scale(1.1)', opacity: 1 },
                    '100%': { transform: 'scale(0.95)', opacity: 0.5 }
                  }
                }} 
              />
            </Tooltip>
            <Box>
              <Typography variant="caption" color="text.secondary" display="block">
                VPN Core Status
              </Typography>
              <Typography variant="body2" fontWeight="bold" sx={{ color: isRunning ? '#00e676' : '#ff1744' }}>
                {isRunning ? 'RUNNING' : 'STOPPED'}
              </Typography>
            </Box>
          </Box>

          <IconButton 
            onClick={logout} 
            sx={{ 
              color: '#ff1744', 
              '&:hover': { background: 'rgba(255, 23, 68, 0.1)' } 
            }}
          >
            <Logout fontSize="small" />
          </IconButton>
        </Box>
      </Box>
    </Drawer>
  );
};

export default Sidebar;
