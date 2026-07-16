import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Grid,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Switch,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControlLabel,
  Tooltip,
  LinearProgress,
  InputAdornment,
  MenuItem
} from '@mui/material';
import {
  Add,
  Delete,
  Edit,
  QrCode,
  Refresh,
  Speed,
  DataUsage,
  People,
  Dns,
  FileCopy
} from '@mui/icons-material';
import { QRCodeSVG } from 'qrcode.react';
import axios from 'axios';
import { formatBytes, formatSpeed, generateHysteriaUrl } from '../utils/vpn';

const Users = ({ systemStats, refreshStats }) => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  
  // Состояния для модальных окон
  const [userModalOpen, setUserModalOpen] = useState(false);
  const [qrModalOpen, setQrModalOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  
  const [selectedUser, setSelectedUser] = useState(null);
  
  // Поля формы пользователя
  const [username, setUsername] = useState('');
  const [authValue, setAuthValue] = useState('');
  const [isEnabled, setIsEnabled] = useState(true);
  const [limitSpeedTx, setLimitSpeedTx] = useState(0); // Mbps
  const [limitSpeedRx, setLimitSpeedRx] = useState(0); // Mbps
  const [limitTraffic, setLimitTraffic] = useState(0); // GB
  const [expireDate, setExpireDate] = useState('');

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    setLoading(true);
    try {
      const response = await axios.get('/api/users');
      setUsers(response.data);
    } catch (error) {
      console.error('Failed to fetch users', error);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenCreateModal = () => {
    setSelectedUser(null);
    setUsername('');
    setAuthValue('');
    setIsEnabled(true);
    setLimitSpeedTx(0);
    setLimitSpeedRx(0);
    setLimitTraffic(0);
    setExpireDate('');
    setUserModalOpen(true);
  };

  const handleOpenEditModal = (user) => {
    setSelectedUser(user);
    setUsername(user.username);
    setAuthValue(user.auth_value);
    setIsEnabled(user.is_enabled);
    setLimitSpeedTx(user.limit_speed_tx / 1000000); // bps -> Mbps
    setLimitSpeedRx(user.limit_speed_rx / 1000000); // bps -> Mbps
    setLimitTraffic(user.limit_traffic / 1073741824); // bytes -> GB
    
    if (user.expire_date) {
      // Преобразуем дату в формат YYYY-MM-DD
      const date = new Date(user.expire_date);
      setExpireDate(date.toISOString().split('T')[0]);
    } else {
      setExpireDate('');
    }
    
    setUserModalOpen(true);
  };

  const handleSaveUser = async (e) => {
    e.preventDefault();
    
    const payload = {
      username,
      auth_value: authValue || undefined,
      is_enabled: isEnabled,
      limit_speed_tx: limitSpeedTx * 1000000, // Mbps -> bps
      limit_speed_rx: limitSpeedRx * 1000000, // Mbps -> bps
      limit_traffic: limitTraffic * 1073741824, // GB -> bytes
      expire_date: expireDate ? new Date(expireDate).toISOString() : null
    };

    try {
      if (selectedUser) {
        await axios.put(`/api/users/${selectedUser.id}`, payload);
      } else {
        await axios.post('/api/users', payload);
      }
      setUserModalOpen(false);
      fetchUsers();
      refreshStats();
    } catch (error) {
      alert(error.response?.data?.error || 'Failed to save user');
    }
  };

  const handleToggleStatus = async (user) => {
    try {
      const updatedUser = { ...user, is_enabled: !user.is_enabled };
      await axios.put(`/api/users/${user.id}`, {
        username: updatedUser.username,
        auth_value: updatedUser.auth_value,
        is_enabled: updatedUser.is_enabled,
        limit_speed_tx: updatedUser.limit_speed_tx,
        limit_speed_rx: updatedUser.limit_speed_rx,
        limit_traffic: updatedUser.limit_traffic,
        expire_date: updatedUser.expire_date
      });
      fetchUsers();
      refreshStats();
    } catch (error) {
      console.error('Failed to toggle user status', error);
    }
  };

  const handleResetStats = async (user) => {
    if (window.confirm(`Are you sure you want to reset traffic statistics for ${user.username}?`)) {
      try {
        await axios.post(`/api/users/${user.id}/reset`);
        fetchUsers();
        refreshStats();
      } catch (error) {
        console.error('Failed to reset stats', error);
      }
    }
  };

  const handleDeleteClick = (user) => {
    setSelectedUser(user);
    setDeleteDialogOpen(true);
  };

  const handleConfirmDelete = async () => {
    try {
      await axios.delete(`/api/users/${selectedUser.id}`);
      setDeleteDialogOpen(false);
      fetchUsers();
      refreshStats();
    } catch (error) {
      console.error('Failed to delete user', error);
    }
  };

  const handleOpenQrModal = (user) => {
    setSelectedUser(user);
    setQrModalOpen(true);
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    alert('Link copied to clipboard!');
  };

  return (
    <Box sx={{ p: 4, width: '100%' }}>
      {/* Шапка страницы */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
        <Box>
          <Typography variant="h4" fontWeight="bold" color="text.primary">
            VPN Users Management
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Manage Hysteria 2 client accounts, usage limits, and credentials
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<Add />}
          onClick={handleOpenCreateModal}
          sx={{
            borderRadius: 2,
            background: 'linear-gradient(135deg, #651fff 0%, #8e24aa 100%)',
            boxShadow: '0 4px 15px rgba(101, 31, 255, 0.3)',
            '&:hover': {
              background: 'linear-gradient(135deg, #7c4dff 0%, #ab47bc 100%)',
            }
          }}
        >
          Add User
        </Button>
      </Box>

      {/* Виджеты общей статистики */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} md={4}>
          <Card sx={{ background: 'rgba(255, 255, 255, 0.02)', border: '1px solid rgba(255, 255, 255, 0.05)', borderRadius: 3 }}>
            <CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <Box sx={{ p: 2, borderRadius: 2, background: 'rgba(101, 31, 255, 0.1)', color: '#b388ff' }}>
                <People />
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">Total Clients / Active</Typography>
                <Typography variant="h5" fontWeight="bold">
                  {systemStats?.total_users || 0} <span style={{ fontSize: '0.8em', fontWeight: 'normal', color: '#00e676' }}>/ {systemStats?.active_users || 0}</span>
                </Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={4}>
          <Card sx={{ background: 'rgba(255, 255, 255, 0.02)', border: '1px solid rgba(255, 255, 255, 0.05)', borderRadius: 3 }}>
            <CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <Box sx={{ p: 2, borderRadius: 2, background: 'rgba(0, 230, 118, 0.1)', color: '#00e676' }}>
                <Dns />
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">Active Connections</Typography>
                <Typography variant="h5" fontWeight="bold">
                  {systemStats?.active_connections || 0}
                </Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={4}>
          <Card sx={{ background: 'rgba(255, 255, 255, 0.02)', border: '1px solid rgba(255, 255, 255, 0.05)', borderRadius: 3 }}>
            <CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <Box sx={{ p: 2, borderRadius: 2, background: 'rgba(255, 23, 68, 0.1)', color: '#ff1744' }}>
                <DataUsage />
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">Total Consumed Traffic</Typography>
                <Typography variant="h5" fontWeight="bold">
                  {formatBytes((systemStats?.total_tx || 0) + (systemStats?.total_rx || 0))}
                </Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Таблица пользователей */}
      <TableContainer component={Paper} sx={{ background: 'rgba(255, 255, 255, 0.01)', border: '1px solid rgba(255, 255, 255, 0.05)', borderRadius: 3 }}>
        <Table>
          <TableHead sx={{ background: 'rgba(255, 255, 255, 0.02)' }}>
            <TableRow>
              <TableCell>Username</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Traffic Usage</TableCell>
              <TableCell>Speed Limits</TableCell>
              <TableCell>Expiration</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {users.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} align="center" sx={{ py: 5, color: 'text.secondary' }}>
                  No users found. Click 'Add User' to create one.
                </TableCell>
              </TableRow>
            ) : (
              users.map((user) => {
                const consumed = (user.stats?.traffic_tx || 0) + (user.stats?.traffic_rx || 0);
                const limit = user.limit_traffic;
                const percentage = limit > 0 ? Math.min((consumed / limit) * 100, 100) : 0;
                
                // Расчет критичности трафика (красный, если > 80%)
                const progressColor = percentage > 80 ? 'error' : 'primary';

                return (
                  <TableRow key={user.id} sx={{ '&:hover': { background: 'rgba(255, 255, 255, 0.01)' } }}>
                    <TableCell>
                      <Typography variant="subtitle2" fontWeight="bold">{user.username}</Typography>
                      <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                        auth: {user.auth_value}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <FormControlLabel
                        control={
                          <Switch
                            checked={user.is_enabled}
                            onChange={() => handleToggleStatus(user)}
                            color="success"
                            size="small"
                          />
                        }
                        label={
                          <Typography variant="body2" sx={{ color: user.is_enabled ? '#00e676' : 'text.secondary', fontWeight: 'bold' }}>
                            {user.is_enabled ? 'Active' : 'Disabled'}
                          </Typography>
                        }
                      />
                    </TableCell>
                    <TableCell sx={{ minWidth: 200 }}>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                        <Typography variant="caption">{formatBytes(consumed)}</Typography>
                        <Typography variant="caption" color="text.secondary">
                          limit: {limit > 0 ? formatBytes(limit) : '∞'}
                        </Typography>
                      </Box>
                      {limit > 0 ? (
                        <LinearProgress
                          variant="determinate"
                          value={percentage}
                          color={progressColor}
                          sx={{ height: 6, borderRadius: 3, background: 'rgba(255,255,255,0.05)' }}
                        />
                      ) : (
                        <Typography variant="caption" color="text.secondary">No limits</Typography>
                      )}
                    </TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.2 }}>
                        <Typography variant="caption" sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                          ↑ {formatSpeed(user.limit_speed_tx)}
                        </Typography>
                        <Typography variant="caption" sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                          ↓ {formatSpeed(user.limit_speed_rx)}
                        </Typography>
                      </Box>
                    </TableCell>
                    <TableCell>
                      {user.expire_date ? (
                        <Typography variant="body2" sx={{ 
                          color: new Date(user.expire_date) < new Date() ? '#ff1744' : 'text.primary',
                          fontWeight: 'medium'
                        }}>
                          {new Date(user.expire_date).toLocaleDateString()}
                        </Typography>
                      ) : (
                        <Typography variant="caption" color="text.secondary">Never expires</Typography>
                      )}
                    </TableCell>
                    <TableCell align="right">
                      <Tooltip title="Get QR Code & Link">
                        <IconButton onClick={() => handleOpenQrModal(user)} color="secondary" size="small">
                          <QrCode />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Edit User">
                        <IconButton onClick={() => handleOpenEditModal(user)} color="primary" size="small">
                          <Edit />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Reset Traffic Stats">
                        <IconButton onClick={() => handleResetStats(user)} color="info" size="small">
                          <Refresh />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Delete User">
                        <IconButton onClick={() => handleDeleteClick(user)} color="error" size="small">
                          <Delete />
                        </IconButton>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </TableContainer>

      {/* 1. Модальное окно создания / редактирования */}
      <Dialog 
        open={userModalOpen} 
        onClose={() => setUserModalOpen(false)} 
        maxWidth="xs" 
        fullWidth
        PaperProps={{
          style: {
            background: '#121020',
            border: '1px solid rgba(255, 255, 255, 0.08)',
            borderRadius: 16
          }
        }}
      >
        <DialogTitle fontWeight="bold">
          {selectedUser ? 'Edit VPN User' : 'Create VPN User'}
        </DialogTitle>
        <form onSubmit={handleSaveUser}>
          <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2.5, pt: 1 }}>
            <TextField
              label="Username / Client Name"
              required
              fullWidth
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="e.g. user_macbook"
              InputLabelProps={{ shrink: true }}
            />
            <TextField
              label="Hysteria Password (Auth Value)"
              fullWidth
              value={authValue}
              onChange={(e) => setAuthValue(e.target.value)}
              placeholder="Leave empty for auto-generation"
              InputLabelProps={{ shrink: true }}
            />
            <Grid container spacing={2}>
              <Grid item xs={6}>
                <TextField
                  label="Speed Limit Up"
                  type="number"
                  fullWidth
                  value={limitSpeedTx}
                  onChange={(e) => setLimitSpeedTx(parseFloat(e.target.value) || 0)}
                  InputProps={{
                    endAdornment: <InputAdornment position="end">Mbps</InputAdornment>,
                  }}
                  helperText="0 = no limit"
                  InputLabelProps={{ shrink: true }}
                />
              </Grid>
              <Grid item xs={6}>
                <TextField
                  label="Speed Limit Down"
                  type="number"
                  fullWidth
                  value={limitSpeedRx}
                  onChange={(e) => setLimitSpeedRx(parseFloat(e.target.value) || 0)}
                  InputProps={{
                    endAdornment: <InputAdornment position="end">Mbps</InputAdornment>,
                  }}
                  helperText="0 = no limit"
                  InputLabelProps={{ shrink: true }}
                />
              </Grid>
            </Grid>
            <TextField
              label="Traffic Limit"
              type="number"
              fullWidth
              value={limitTraffic}
              onChange={(e) => setLimitTraffic(parseFloat(e.target.value) || 0)}
              InputProps={{
                endAdornment: <InputAdornment position="end">GB</InputAdornment>,
              }}
              helperText="0 = no limit"
              InputLabelProps={{ shrink: true }}
            />
            <TextField
              label="Expiration Date"
              type="date"
              fullWidth
              value={expireDate}
              onChange={(e) => setExpireDate(e.target.value)}
              InputLabelProps={{ shrink: true }}
              helperText="Leave empty for unlimited time"
            />
            <FormControlLabel
              control={
                <Switch
                  checked={isEnabled}
                  onChange={(e) => setIsEnabled(e.target.checked)}
                  color="success"
                />
              }
              label="Enable Account"
            />
          </DialogContent>
          <DialogActions sx={{ p: 3 }}>
            <Button onClick={() => setUserModalOpen(false)} color="inherit">Cancel</Button>
            <Button
              type="submit"
              variant="contained"
              sx={{
                borderRadius: 2,
                background: 'linear-gradient(135deg, #651fff 0%, #8e24aa 100%)',
              }}
            >
              Save
            </Button>
          </DialogActions>
        </form>
      </Dialog>

      {/* 2. Модальное окно QR кода и ссылки */}
      <Dialog 
        open={qrModalOpen} 
        onClose={() => setQrModalOpen(false)}
        PaperProps={{
          style: {
            background: '#121020',
            border: '1px solid rgba(255, 255, 255, 0.08)',
            borderRadius: 16
          }
        }}
      >
        <DialogTitle fontWeight="bold">Share Connection Settings</DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3, pt: 1 }}>
          {selectedUser && (
            <>
              <Typography variant="subtitle1" fontWeight="bold">
                Client: {selectedUser.username}
              </Typography>
              
              {/* QR Code Container */}
              <Box sx={{ 
                p: 2, 
                borderRadius: 3, 
                background: '#fff', 
                boxShadow: '0 0 20px rgba(255,255,255,0.1)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center'
              }}>
                <QRCodeSVG 
                  value={generateHysteriaUrl(selectedUser, systemStats)} 
                  size={220} 
                  level="M"
                />
              </Box>

              <TextField
                fullWidth
                variant="outlined"
                value={generateHysteriaUrl(selectedUser, systemStats)}
                label="Hysteria 2 Link"
                InputProps={{
                  readOnly: true,
                  endAdornment: (
                    <InputAdornment position="end">
                      <IconButton onClick={() => copyToClipboard(generateHysteriaUrl(selectedUser, systemStats))}>
                        <FileCopy />
                      </IconButton>
                    </InputAdornment>
                  ),
                }}
                InputLabelProps={{ shrink: true }}
              />
            </>
          )}
        </DialogContent>
        <DialogActions sx={{ p: 3 }}>
          <Button onClick={() => setQrModalOpen(false)} variant="contained" color="secondary">
            Close
          </Button>
        </DialogActions>
      </Dialog>

      {/* 3. Диалог подтверждения удаления */}
      <Dialog 
        open={deleteDialogOpen} 
        onClose={() => setDeleteDialogOpen(false)}
        PaperProps={{
          style: {
            background: '#121020',
            border: '1px solid rgba(255, 255, 255, 0.08)',
            borderRadius: 16
          }
        }}
      >
        <DialogTitle fontWeight="bold">Confirm User Deletion</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to delete user <b>{selectedUser?.username}</b>? This action cannot be undone and their statistics will be cleared.
          </Typography>
        </DialogContent>
        <DialogActions sx={{ p: 3 }}>
          <Button onClick={() => setDeleteDialogOpen(false)} color="inherit">Cancel</Button>
          <Button onClick={handleConfirmDelete} variant="contained" color="error">
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Users;
