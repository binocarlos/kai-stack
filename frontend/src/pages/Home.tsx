import React from 'react'
import { Box, Typography } from '@mui/material'

export const Home: React.FC = () => {
  return (
    <Box sx={{ display: 'flex', height: 'calc(100vh - 64px)' }}>
      <Typography variant="h1">Home</Typography>
    </Box>
  )
}

export default Home