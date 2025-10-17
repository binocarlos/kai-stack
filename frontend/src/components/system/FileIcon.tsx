import React from 'react'
import { Box } from '@mui/material'

// Import the colorful SVG icons
import folderIcon from '../../icons/default_folder.svg'
import folderOpenIcon from '../../icons/default_folder_opened.svg'
import pdfIcon from '../../icons/file_type_pdf2.svg'
import wordIcon from '../../icons/file_type_word2.svg'
import excelIcon from '../../icons/file_type_excel2.svg'
import powerPointIcon from '../../icons/file_type_powerpoint2.svg'
import imageIcon from '../../icons/file_type_image.svg'
import textIcon from '../../icons/file_type_text.svg'
import zipIcon from '../../icons/file_type_zip2.svg'
import defaultIcon from '../../icons/default_file.svg'

interface FileIconProps {
  filename: string
  isFolder: boolean
  isOpen?: boolean
  size?: number
}

// Helper function to get file extension
const getFileExtension = (filename: string): string => {
  return filename.toLowerCase().split('.').pop() || ''
}

export const FileIcon: React.FC<FileIconProps> = ({ 
  filename, 
  isFolder = false, 
  isOpen = false, 
  size = 24 
}) => {
  const getIconPath = (): string => {
    if (isFolder) {
      return isOpen ? folderOpenIcon : folderIcon
    }

    const extension = getFileExtension(filename)

    switch (extension) {
      // Documents
      case 'pdf':
        return pdfIcon
      case 'doc':
      case 'docx':
        return wordIcon
      
      // Spreadsheets
      case 'xls':
      case 'xlsx':
      case 'csv':
        return excelIcon
      
      // Presentations
      case 'ppt':
      case 'pptx':
        return powerPointIcon
      
      // Images
      case 'jpg':
      case 'jpeg':
      case 'png':
      case 'gif':
      case 'bmp':
      case 'webp':
        return imageIcon
      
      // Text files
      case 'txt':
      case 'log':
      case 'md':
        return textIcon
      
      // Archives
      case 'zip':
      case 'rar':
      case '7z':
      case 'tar':
      case 'gz':
        return zipIcon
      
      default:
        return defaultIcon
    }
  }

  return (
    <Box
      component="img"
      src={getIconPath()}
      alt={`${filename} icon`}
      sx={{
        width: size,
        height: size,
        objectFit: 'contain',
        marginRight: 1
      }}
    />
  )
}

export default FileIcon