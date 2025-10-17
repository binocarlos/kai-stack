import React, { FC } from 'react'
import CircularProgress, {
  CircularProgressProps,
} from '@mui/material/CircularProgress'
import Typography from '@mui/material/Typography'
import Box from '@mui/material/Box'

interface LoadingProps {
  color?: CircularProgressProps["color"],
  message?: string,
}

// Using functional component with sx props instead of makeStyles
const Loading: FC<React.PropsWithChildren<LoadingProps>> = ({
  color = 'primary',
  message = 'loading',
  children,
}) => {
  return (
    // Root container with centered content
    <Box
      sx={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100%',
      }}
    >
      {/* Container with max width constraint */}
      <Box sx={{ maxWidth: '100%' }}>
        {/* Item container for progress and message */}
        <Box sx={{ textAlign: 'center', display: 'inline-block' }}>
          <CircularProgress 
            color={color}
          />
          { 
            message && (
              <Typography
                variant='subtitle1'
                color={color}
              >
                {message}
              </Typography>
            )
          }
          {children}
        </Box>
      </Box>
    </Box>
  )
}

export default Loading
