import React, { FC } from 'react'
import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'

interface ErrorProps {
  message: string;
}

// Functional component to display an error message using MUI Alert
const ErrorComponent: FC<ErrorProps> = ({ message }) => {
  return (
    // Centered container for the alert
    <Box
      sx={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100%', // Take full height to center vertically
        padding: 2, // Add some padding around the alert
      }}
    >
      {/* MUI Alert component to display the error */}
      <Alert severity="error" sx={{ width: '100%', maxWidth: 600 }}> 
        {/* Display the error message passed via props */}
        {message}
      </Alert>
    </Box>
  )
}

export default ErrorComponent