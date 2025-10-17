import React, { useState } from 'react'
import { Box, TextField, Button, Typography, Alert, Link } from '@mui/material'
import useAccount from '../hooks/useAccount'
import useRouter from '../hooks/useRouter'

export const Home: React.FC = () => {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const account = useAccount()
  const router = useRouter()

  const handleLogin = async () => {
    await account.onLogin(username, password)
  }
 
  const handleForgotPassword = () => {
    router.navigate('forgot-password')
  }

  return (
    <Box sx={{ display: 'flex', height: 'calc(100vh - 64px)' }}>
 
      {/* Right section */}
      <Box sx={{ width: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Box
          sx={{
            maxWidth: 400,
            width: '100%',
            p: 3,
            backgroundColor: 'background.paper',
            borderRadius: 2,
            border: '1px solid',
            borderColor: 'divider',
          }}
        >
          <Typography variant="h4" component="h1" gutterBottom>
            Login
          </Typography>
          <Box>
            <TextField
              fullWidth
              label="Email"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              margin="normal"
              required
              variant="outlined"
              InputLabelProps={{
                shrink: true,
              }}
            />
            <TextField
              fullWidth
              type="password"
              label="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              margin="normal"
              required
              variant="outlined"
              InputLabelProps={{
                shrink: true,
              }}
            />
            <Button
              fullWidth
              variant="contained"
              sx={{ mt: 2 }}
              onClick={handleLogin}
            >
              Login
            </Button>
          </Box>
        </Box>
      </Box>
    </Box>
  )
}

export default Home