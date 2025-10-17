import React, { useState, useEffect, FC } from 'react'
import Button from '@mui/material/Button'
import Dialog from '@mui/material/Dialog'
import DialogActions from '@mui/material/DialogActions'
import DialogContent from '@mui/material/DialogContent'
import DialogContentText from '@mui/material/DialogContentText'
import DialogTitle from '@mui/material/DialogTitle'
import TextField from '@mui/material/TextField'

interface DeleteConfirmWindowProps {
  open: boolean
  title?: string
  message?: string
  confirmWord?: string
  onConfirm: () => void
  onCancel: () => void
}

const DeleteConfirmWindow: FC<DeleteConfirmWindowProps> = ({
  open,
  title = "Confirm Deletion",
  message = "Please confirm if you want to delete this item.",
  confirmWord = "confirm",
  onConfirm,
  onCancel,
}) => {
  const [inputValue, setInputValue] = useState('')

  useEffect(() => {
    // Reset input value when dialog is opened or closed
    if (!open) {
      setInputValue('')
    }
  }, [open])

  const handleConfirm = () => {
    if (inputValue === confirmWord) {
      onConfirm()
    }
  }

  let isConfirmDisabled = inputValue !== confirmWord
  if (confirmWord === '') {
    isConfirmDisabled = false
  }

  return (
    <Dialog
      open={open}
      onClose={onCancel}
      aria-labelledby="delete-confirm-dialog-title"
      aria-describedby="delete-confirm-dialog-description"
    >
      <DialogTitle id="delete-confirm-dialog-title">{title}</DialogTitle>
      <DialogContent>
        <DialogContentText id="delete-confirm-dialog-description">
          {message}
        </DialogContentText>
        {
          confirmWord !== '' && (
            <TextField
              autoFocus
              margin="dense"
              id="confirm-input"
              label={`Type "${confirmWord}" to confirm`}
              type="text"
              fullWidth
              variant="standard"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
            />
          )
        }
        
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel}>Cancel</Button>
        <Button 
          onClick={handleConfirm} 
          color="error" 
          disabled={isConfirmDisabled}
        >
          Confirm
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default DeleteConfirmWindow 