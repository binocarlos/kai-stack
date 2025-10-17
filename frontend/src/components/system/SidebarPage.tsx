import React, { useState, useCallback } from 'react'
import { Box, Paper, Button, Typography, Tooltip, IconButton, Stack } from '@mui/material'
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft'
import ChevronRightIcon from '@mui/icons-material/ChevronRight'

export const SideBarPage: React.FC<{
  title: string,
  sidebar: React.ReactNode,
  children: React.ReactNode,
}> = ({
  title,
  sidebar,
  children,
}) => {
  const [sidebarVisible, setSidebarVisible] = useState(true)
  return (
    <Box sx={{ 
      display: 'flex', 
      height: 'calc(100vh - 64px)', 
      overflow: 'hidden',
      pl: 2, 
      pr: 2,
      py: 2,
    }}>
      {
        sidebarVisible ? (
          <Box sx={{ 
            display: 'flex', 
            flexDirection: 'column', 
            mr: 2, 
            p: 2,
            width: 420,
            height: '100%',
            overflow: 'hidden',
            backgroundColor: 'background.paper',
          }}>
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
              <Typography variant="h6">{ title }</Typography>
              <Tooltip title="Collapse sidebar">
                <IconButton 
                  size="small" 
                  onClick={() => {
                    setSidebarVisible(false)
                  }}
                  sx={{ ml: 1 }}
                >
                  <ChevronLeftIcon />
                </IconButton>
              </Tooltip>
            </Box>
            <Paper 
              elevation={0} 
              sx={{ 
                flex: 1,
                display: 'flex', 
                flexDirection: 'column', 
                overflow: 'hidden',
                minWidth: 0
              }}
            >
              {sidebar}
            </Paper>
          </Box>
        ) : (
          <Paper
            elevation={0}
            sx={{
              width: 48,
              mr: 2,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              pt: 2
            }}
          >
            <Tooltip title="Show Table of Contents" placement="right">
              <IconButton
                onClick={() => setSidebarVisible(true)}
                size="small"
                sx={{
                  bgcolor: 'action.hover',
                  '&:hover': { bgcolor: 'action.selected' }
                }}
              >
                <ChevronRightIcon />
              </IconButton>
            </Tooltip>
          </Paper>
        )
      }

      <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden', position: 'relative' }}>
        { children }
      </Box>

    </Box>
  )
}

export default SideBarPage