import React from 'react'
import Box from '@mui/material/Box'
import CircularProgress from '@mui/material/CircularProgress'
import Typography from '@mui/material/Typography'

import Overlay from './Overlay'
import { LoadingContext } from '../../contexts/loading'

const GlobalLoading: React.FC<React.PropsWithChildren<{
  title?: string,
}>> = ({
  title = 'loading...',
  children,
}) => {
  const loadingContext = React.useContext(LoadingContext)

  if(!loadingContext.loading) return null

  return (
    <Overlay>
      <Box>
        <Box
          component="div"
          sx={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            height: '100%',
          }}
        >
          <Box
            component="div"
            sx={{
              maxWidth: '100%'
            }}
          >
            <Box
              component="div"
              sx={{
                textAlign: 'center',
                display: 'inline-block',
              }}
            >
              <CircularProgress />
              <Typography variant='subtitle1'>
                { title }
              </Typography>
            </Box>
          </Box>
        </Box>
        { children }
      </Box>
    </Overlay>
  )
}

export default GlobalLoading
