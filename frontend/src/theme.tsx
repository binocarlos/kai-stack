import React from 'react'
import { createTheme, responsiveFontSizes } from '@mui/material/styles'

const theme = responsiveFontSizes(createTheme({
  palette: {
    primary: {
      main: '#000000',
    },
    secondary: {
      main: '#000000',
    },
  },
}))

export default theme
