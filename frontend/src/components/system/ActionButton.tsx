import React, { useState, useEffect, useMemo } from 'react'
import { Button } from '@mui/material'

interface ActionButtonProps {
  icon: React.ReactElement
  label: string
  onClick: () => void
}

const ActionButton: React.FC<ActionButtonProps> = ({ icon, label, onClick }) => (
  <Button
    variant="outlined"
    startIcon={icon}
    onClick={onClick}
    size="small"
    sx={{ 
      fontSize: '0.75rem',
      padding: '4px 8px',
      color: 'grey.600',
      borderColor: 'grey.400',
      '&:hover': {
        borderColor: 'grey.500',
        backgroundColor: 'grey.50'
      }
    }}
  >
    {label}
  </Button>
)

export default ActionButton
