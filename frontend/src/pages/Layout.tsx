import React, { useContext } from 'react'
import { AppBar, Toolbar, FormControl, InputLabel, Select, MenuItem, SelectChangeEvent, Box, Button, LinearProgress, Typography } from '@mui/material'
import Backdrop from '@mui/material/Backdrop'
import CircularProgress from '@mui/material/CircularProgress'
import Snackbar from '@mui/material/Snackbar'
import Alert from '@mui/material/Alert'
import axios from 'axios'

import useAccount from '../hooks/useAccount'
import useRouter from '../hooks/useRouter'
import { LoadingContext } from '../contexts/loading'
import { SnackbarContext } from '../contexts/snackbar'

interface LayoutProps {
  children: React.ReactNode
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
  const account = useAccount()
  const router = useRouter()
  const { loading, progressTitle, progressTotal, progressCurrent } = useContext(LoadingContext)
  const { snackbar, setSnackbar } = useContext(SnackbarContext)

  const isRouteActive = (name: string) => router.name == name

  const handleSnackbarClose = (event?: React.SyntheticEvent | Event, reason?: string) => {
    if (reason === 'clickaway') return
    setSnackbar(undefined)
  };

  const isTikTokPage = router.name?.startsWith('tiktok-')

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <AppBar
        position="fixed"
        color="default"
        elevation={0}
      >
        <Toolbar
          sx={{
            justifyContent: 'space-between', 
            backgroundColor: 'white',
            borderBottom: '6px solid #000000',
          }}
        >
          <Box component="div" sx={{ display: 'flex', alignItems: 'center' }}>
            <Typography variant="h3">
              &lt;&gt;&nbsp;
            </Typography>
            <Typography variant="h6">
              stack
            </Typography>
          </Box>
        </Toolbar>
      </AppBar>
      <Box id="main-content" component="main" sx={{ flexGrow: 1, mt: '64px', backgroundColor: 'rgb(245, 245, 245)' }}>
        {children}
      </Box>

      {/* Global Loading Backdrop */}
      <Backdrop
        sx={{ color: '#fff', zIndex: (theme) => theme.zIndex.drawer + 1 }}
        open={loading}
      >
        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
          {/* Show progress title if it exists */}
          {progressTitle && (
            <Typography variant="h6" color="inherit" sx={{ textAlign: 'center' }}>
              {progressTitle}
            </Typography>
          )}

          {/* Show progress bar only if total is greater than 0 */}
          {progressTotal > 0 ? (
            <Box sx={{ width: 300, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 1 }}>
              <LinearProgress
                variant="determinate"
                value={(progressCurrent / progressTotal) * 100}
                sx={{
                  width: '100%',
                  height: 8,
                  borderRadius: 4,
                  backgroundColor: 'rgba(255, 255, 255, 0.3)',
                  '& .MuiLinearProgress-bar': {
                    borderRadius: 4,
                    backgroundColor: '#FFC000'
                  }
                }}
              />
              <Typography variant="body2" color="inherit">
                {progressCurrent} of {progressTotal}
              </Typography>
            </Box>
          ) : (
            // Show circular progress when no specific progress is tracked
            <CircularProgress color="inherit" />
          )}
        </Box>
      </Backdrop>

      {/* Global Snackbar */}
      <Snackbar
        open={!!snackbar}
        autoHideDuration={6000}
        onClose={handleSnackbarClose}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        {/* Render Alert only when snackbar has data */}
        {snackbar && (
          <Alert
            onClose={handleSnackbarClose}
            severity={snackbar.severity}
            sx={{
              width: '100%',
              bgcolor: snackbar.color,
              color: 'white',
              border: `1px solid #333333`, // Add border with 85% opacity of background color
              '& .MuiAlert-icon': {
                color: 'white'
              }
            }}
          >
            {snackbar.message}
          </Alert>
        )}
      </Snackbar>
      
    </Box>
  )
}

export default Layout
