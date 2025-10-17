import React, { FC, useEffect, useState } from 'react'
import Dialog from '@mui/material/Dialog'
import DialogTitle from '@mui/material/DialogTitle'
import DialogContent from '@mui/material/DialogContent'
import DialogContentText from '@mui/material/DialogContentText'
import DialogActions from '@mui/material/DialogActions'
import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'

type StringEditorWindowMode = 'add' | 'edit'

interface StringEditorWindowProps {
  open: boolean
  mode?: StringEditorWindowMode
  title?: string
  description?: string
  label?: string
  initialValue?: string
  confirmLabel?: string
  onCancel: () => void
  onSave: (value: string) => void
}

const getDefaultTitle = (mode: StringEditorWindowMode) => {
  if (mode === 'edit') {
    return 'Edit Value'
  }
  return 'Add Value'
}

const getDefaultConfirmLabel = (mode: StringEditorWindowMode) => {
  if (mode === 'edit') {
    return 'Save'
  }
  return 'Add'
}

const StringEditorWindow: FC<StringEditorWindowProps> = ({
  open,
  mode = 'add',
  title,
  description,
  label = 'Value',
  initialValue = '',
  confirmLabel,
  onCancel,
  onSave,
}) => {
  const [value, setValue] = useState(initialValue)

  useEffect(() => {
    if (open) {
      setValue(initialValue)
    }
  }, [open, initialValue])

  const handleSave = () => {
    const trimmed = value.trim()
    if (!trimmed) {
      return
    }
    onSave(trimmed)
  }

  const effectiveTitle = title ?? getDefaultTitle(mode)
  const effectiveConfirmLabel = confirmLabel ?? getDefaultConfirmLabel(mode)

  return (
    <Dialog open={open} onClose={onCancel} fullWidth maxWidth='xs'>
      <DialogTitle>{effectiveTitle}</DialogTitle>
      <DialogContent>
        {description ? (
          <DialogContentText sx={{ mb: 2 }}>
            {description}
          </DialogContentText>
        ) : null}
        <TextField
          fullWidth
          label={label}
          value={value}
          onChange={event => setValue(event.target.value)}
          sx={{mt: 1}}
          autoFocus
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel}>Cancel</Button>
        <Button onClick={handleSave} disabled={value.trim() === ''} variant='contained'>
          {effectiveConfirmLabel}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default StringEditorWindow
