import React from 'react'
import Box from '@mui/material/Box'
import LinearProgress from '@mui/material/LinearProgress'
import Typography from '@mui/material/Typography'
import prettyBytes from 'pretty-bytes'

import Overlay from './Overlay'

interface UploadProgressProps {
  isUploading: boolean
  progress: number // Percentage 0-100
  fileSize: number // Total file size in bytes
}

const UploadProgress: React.FC<UploadProgressProps> = ({
  isUploading,
  progress,
  fileSize,
}) => {
  if (!isUploading) {
    return null // Don't render anything if not uploading
  }

  // Calculate the number of bytes uploaded so far
  const uploadedBytes = Math.round((progress / 100) * fileSize)

  // Format numbers using pretty-bytes
  const formattedUploaded = prettyBytes(uploadedBytes)
  const formattedTotal = prettyBytes(fileSize)

  return (
    <Overlay>
      <Box sx={{ width: '300px', textAlign: 'center' }}>
        <Typography variant="h6" gutterBottom>
          Uploading File...
        </Typography>
        <LinearProgress variant="determinate" value={progress} sx={{ mb: 2 }} />
        <Typography variant="body1" sx={{ mb: 1 }}>
          {progress}% Complete
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {formattedUploaded} / {formattedTotal}
        </Typography>
      </Box>
    </Overlay>
  )
}

export default UploadProgress 