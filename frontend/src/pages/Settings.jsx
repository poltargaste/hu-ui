import React, { useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Grid,
  Typography,
  TextField,
  Divider,
  Alert,
  CircularProgress
} from '@mui/material';
import {
  Save,
  PlayArrow,
  Stop,
  Autorenew,
  VpnKey,
  SettingsSystemDaydream
} from '@mui/icons-material';
import axios from 'axios';

const Settings = ({ systemStats, refreshStats }) => {
  // Состояния для смены пароля
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [pwdError, setPwdError] = useState('');
  const [pwdSuccess, setPwdSuccess] = useState('');
  const [pwdLoading, setPwdLoading] = useState(false);

  // Состояние для управления ядром
  const [coreLoading, setCoreLoading] = useState(false);

  const isRunning = systemStats?.hysteria_running || false;

  const handleChangePassword = async (e) => {
    e.preventDefault();
    setPwdError('');
    setPwdSuccess('');

    if (newPassword !== confirmPassword) {
      setPwdError("New passwords do not match");
      return;
    }

    setPwdLoading(true);
    try {
      await axios.post('/api/auth/change-password', {
        old_password: oldPassword,
        new_password: newPassword
      });
      setPwdSuccess("Password changed successfully!");
      setOldPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (error) {
      setPwdError(error.response?.data?.error || 'Failed to change password');
    } finally {
      setPwdLoading(false);
    }
  };

  const handleCoreAction = async (action) => {
    setCoreLoading(true);
    try {
      await axios.post(`/api/system/core/${action}`);
      setTimeout(() => {
        refreshStats();
        setCoreLoading(false);
      }, 1000); // Даем ядру время обновить статус
    } catch (error) {
      alert(error.response?.data?.error || `Failed to ${action} core`);
      setCoreLoading(false);
    }
  };

  return (
    <Box sx={{ p: 4, width: '100%' }}>
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" fontWeight="bold" color="text.primary">
          Settings & System Core
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Configure security credentials and monitor Hysteria 2 process states
        </Typography>
      </Box>

      <Grid container spacing={4}>
        {/* Левая колонка: Управление VPN Ядром */}
        <Grid item xs={12} md={6}>
          <Card sx={{ background: 'rgba(255, 255, 255, 0.02)', border: '1px solid rgba(255, 255, 255, 0.05)', borderRadius: 3 }}>
            <CardContent sx={{ p: 3 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2 }}>
                <SettingsSystemDaydream sx={{ color: '#b388ff' }} />
                <Typography variant="h6" fontWeight="bold">Hysteria 2 Process Control</Typography>
              </Box>
              
              <Divider sx={{ borderColor: 'rgba(255, 255, 255, 0.05)', mb: 3 }} />

              <Box sx={{ 
                display: 'flex', 
                flexDirection: 'column', 
                alignItems: 'center', 
                p: 3, 
                mb: 3,
                borderRadius: 2.5,
                background: 'rgba(255, 255, 255, 0.01)',
                border: '1px solid rgba(255, 255, 255, 0.03)'
              }}>
                <Typography variant="body2" color="text.secondary" gutterBottom>Current Process State</Typography>
                <Typography 
                  variant="h5" 
                  fontWeight="bold" 
                  sx={{ 
                    color: isRunning ? '#00e676' : '#ff1744', 
                    letterSpacing: 1, 
                    mb: 1 
                  }}
                >
                  {isRunning ? 'RUNNING' : 'STOPPED'}
                </Typography>
                
                {coreLoading && <CircularProgress size={24} sx={{ color: '#b388ff', mt: 1 }} />}
              </Box>

              {/* Кнопки управления */}
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  {isRunning ? (
                    <Button
                      fullWidth
                      variant="contained"
                      color="error"
                      startIcon={<Stop />}
                      onClick={() => handleCoreAction('stop')}
                      disabled={coreLoading}
                      sx={{ py: 1.2, borderRadius: 2, fontWeight: 'bold' }}
                    >
                      Stop Hysteria Core
                    </Button>
                  ) : (
                    <Button
                      fullWidth
                      variant="contained"
                      color="success"
                      startIcon={<PlayArrow />}
                      onClick={() => handleCoreAction('start')}
                      disabled={coreLoading}
                      sx={{ py: 1.2, borderRadius: 2, fontWeight: 'bold' }}
                    >
                      Start Hysteria Core
                    </Button>
                  )}
                </Grid>
                <Grid item xs={12}>
                  <Button
                    fullWidth
                    variant="outlined"
                    color="secondary"
                    startIcon={<Autorenew />}
                    onClick={() => handleCoreAction('restart')}
                    disabled={coreLoading || !isRunning}
                    sx={{ py: 1.2, borderRadius: 2, fontWeight: 'bold', borderColor: 'rgba(255, 255, 255, 0.1)' }}
                  >
                    Graceful Restart
                  </Button>
                </Grid>
              </Grid>

              {/* Информация о текущей конфигурации */}
              <Box sx={{ mt: 4, p: 2, borderRadius: 2, background: 'rgba(255,255,255,0.01)', border: '1px solid rgba(255,255,255,0.02)' }}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>Core Connection Info</Typography>
                <Typography variant="body2" sx={{ mt: 1 }}>
                  <b>Listen Port:</b> {systemStats?.hysteria_port || 'Loading...'}
                </Typography>
                <Typography variant="body2" sx={{ mt: 0.5, fontFamily: 'monospace' }}>
                  <b>Obfs Key:</b> {systemStats?.hysteria_obfs || 'Disabled (no obfs)'}
                </Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Правая колонка: Смена пароля администратора */}
        <Grid item xs={12} md={6}>
          <Card sx={{ background: 'rgba(255, 255, 255, 0.02)', border: '1px solid rgba(255, 255, 255, 0.05)', borderRadius: 3 }}>
            <CardContent sx={{ p: 3 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 2 }}>
                <VpnKey sx={{ color: '#b388ff' }} />
                <Typography variant="h6" fontWeight="bold">Change Admin Password</Typography>
              </Box>
              
              <Divider sx={{ borderColor: 'rgba(255, 255, 255, 0.05)', mb: 3 }} />

              {pwdError && <Alert severity="error" sx={{ mb: 2, borderRadius: 2 }}>{pwdError}</Alert>}
              {pwdSuccess && <Alert severity="success" sx={{ mb: 2, borderRadius: 2 }}>{pwdSuccess}</Alert>}

              <Box component="form" onSubmit={handleChangePassword} sx={{ display: 'flex', flexDirection: 'column', gap: 2.5 }}>
                <TextField
                  label="Old Password"
                  type="password"
                  required
                  fullWidth
                  value={oldPassword}
                  onChange={(e) => setOldPassword(e.target.value)}
                  InputLabelProps={{ shrink: true }}
                />
                <TextField
                  label="New Password"
                  type="password"
                  required
                  fullWidth
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  InputLabelProps={{ shrink: true }}
                />
                <TextField
                  label="Confirm New Password"
                  type="password"
                  required
                  fullWidth
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  InputLabelProps={{ shrink: true }}
                />
                <Button
                  type="submit"
                  variant="contained"
                  disabled={pwdLoading}
                  startIcon={<Save />}
                  sx={{
                    py: 1.2,
                    borderRadius: 2,
                    fontWeight: 'bold',
                    background: 'linear-gradient(135deg, #651fff 0%, #8e24aa 100%)',
                  }}
                >
                  {pwdLoading ? 'Saving...' : 'Update Password'}
                </Button>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Settings;
