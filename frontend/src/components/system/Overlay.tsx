import React from 'react'
import Box from '@mui/material/Box'

// Reusable overlay component for modals, loading indicators, etc.
const Overlay: React.FC<React.PropsWithChildren> = ({ children }) => {
  return (
    <Box
      component="div"
      // Basic overlay styling
      sx={{
        position: 'fixed',
        left: '0px',
        top: '0px',
        zIndex: 10000, // High z-index to appear on top
        width: '100%',
        height: '100%',
        display: 'flex',
        justifyContent: 'center', // Center content horizontally
        alignItems: 'center',    // Center content vertically
        backgroundColor: 'rgba(255, 255, 255, 0.85)', // Semi-transparent white background
        // Prevent scrolling of the background content when overlay is active
        // Note: This might need adjustments depending on the overall layout
        overflow: 'hidden',
      }}
    >
      {/* Container for the content centered within the overlay */}
      <Box
        component="div"
        sx={{
          padding: 4, // Add some padding around the content
          // Optional: Add a background or border to the content box itself
          // backgroundColor: '#ffffff',
          // borderRadius: 1,
          // boxShadow: 3,
          maxWidth: '90%', // Prevent content from becoming too wide
          maxHeight: '90%', // Prevent content from becoming too tall
          overflow: 'auto', // Allow scrolling within the content box if needed
          textAlign: 'center', // Center text content by default
        }}
      >
        {children} {/* Render the content passed to the overlay */}
      </Box>
    </Box>
  )
}

export default Overlay 