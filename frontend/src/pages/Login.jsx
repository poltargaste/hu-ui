import React, { useState } from 'react';
import { 
  Box, 
  Card, 
  CardContent, 
  TextField, 
  Button, 
  Typography, 
  Alert,
  InputAdornment,
  IconButton
} from '@mui/material';
import { Visibility, VisibilityOff, VpnLock } from '@mui/icons-material';
import { useAuth } from '../context/AuthContext';

const Login = () => {
  const { login } = useAuth();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    const result = await login(username, password);
    if (!result.success) {
      setError(result.error);
      setLoading(false);
    }
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'radial-gradient(circle at 10% 20%, rgba(26, 20, 48, 1) 0%, rgba(12, 10, 22, 1) 90.2%)',
        padding: 3,
      }}
    >
      {/* Декоративные размытые круги на фоне */}
      <Box
        sx={{
          position: 'absolute',
          width: '300px',
          height: '300px',
          borderRadius: '50%',
          background: 'linear-gradient(45deg, #7c4dff, #18ffff)',
          filter: 'blur(80px)',
          opacity: 0.15,
          top: '20%',
          left: '30%',
          zIndex: 0,
        }}
      />
      <Box
        sx={{
          position: 'absolute',
          width: '250px',
          height: '250px',
          borderRadius: '50%',
          background: 'linear-gradient(45deg, #ff4081, #651fff)',
          filter: 'blur(70px)',
          opacity: 0.15,
          bottom: '20%',
          right: '30%',
          zIndex: 0,
        }}
      />

      <Card
        sx={{
          maxWidth: 400,
          width: '100%',
          background: 'rgba(255, 255, 255, 0.03)',
          backdropFilter: 'blur(20px)',
          border: '1px solid rgba(255, 255, 255, 0.08)',
          borderRadius: 4,
          boxShadow: '0 8px 32px 0 rgba(0, 0, 0, 0.37)',
          zIndex: 1,
          animation: 'fadeInUp 0.6s ease-out',
          '@keyframes fadeInUp': {
            from: { opacity: 0, transform: 'translateY(20px)' },
            to: { opacity: 1, transform: 'translateY(0)' },
          }
        }}
      >
        <CardContent sx={{ p: 4 }}>
          <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', mb: 4 }}>
            <Box
              sx={{
                p: 2,
                borderRadius: '50%',
                background: 'linear-gradient(135deg, #651fff, #8e24aa)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                boxShadow: '0 0 20px rgba(101, 31, 255, 0.4)',
                mb: 2
              }}
            >
              <VpnLock sx={{ fontSize: 32, color: '#fff' }} />
            </Box>
            <Typography variant="h5" component="h1" fontWeight="bold" color="text.primary">
              HYSTERIA 2
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
              VPN Admin Panel Management
            </Typography>
          </Box>

          {error && (
            <Alert severity="error" sx={{ mb: 3, borderRadius: 2 }}>
              {error}
            </Alert>
          )}

          <Box component="form" onSubmit={handleSubmit}>
            <TextField
              margin="normal"
              required
              fullWidth
              id="username"
              label="Admin Login"
              name="username"
              autoComplete="username"
              autoFocus
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              variant="outlined"
              InputLabelProps={{ shrink: true }}
              placeholder="Enter username"
              sx={{
                '& .MuiOutlinedInput-root': {
                  borderRadius: 2.5,
                  '& fieldset': { borderColor: 'rgba(255,255,255,0.1)' },
                  '&:hover fieldset': { borderColor: 'rgba(255,255,255,0.2)' },
                }
              }}
            />
            <TextField
              margin="normal"
              required
              fullWidth
              name="password"
              label="Password"
              type={showPassword ? 'text' : 'password'}
              id="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              variant="outlined"
              InputLabelProps={{ shrink: true }}
              placeholder="Enter password"
              InputProps={{
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton
                      aria-label="toggle password visibility"
                      onClick={() => setShowPassword(!showPassword)}
                      edge="end"
                    >
                      {showPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                ),
              }}
              sx={{
                mt: 2,
                mb: 3,
                '& .MuiOutlinedInput-root': {
                  borderRadius: 2.5,
                  '& fieldset': { borderColor: 'rgba(255,255,255,0.1)' },
                  '&:hover fieldset': { borderColor: 'rgba(255,255,255,0.2)' },
                }
              }}
            />
            <Button
              type="submit"
              fullWidth
              variant="contained"
              disabled={loading}
              sx={{
                py: 1.5,
                borderRadius: 2.5,
                fontWeight: 'bold',
                textTransform: 'none',
                background: 'linear-gradient(135deg, #651fff 0%, #8e24aa 100%)',
                boxShadow: '0 4px 15px rgba(101, 31, 255, 0.3)',
                '&:hover': {
                  background: 'linear-gradient(135deg, #7c4dff 0%, #ab47bc 100%)',
                  boxShadow: '0 6px 20px rgba(101, 31, 255, 0.4)',
                }
              }}
            >
              {loading ? 'Entering...' : 'Log In'}
            </Button>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
};

export default Login;
