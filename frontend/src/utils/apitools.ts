export const extractErrorMessage = (error: any): string => {
  if(error.response && error.response.data) {
    if (error.response.data.message || error.response.data.error) {
      return (error.response.data.message || error.response.data.error) as string
    }
    if (error.response.data) return error.response.data as string
    return error.toString()
  }
  else if(error.error) {
    return error.error
  }
  else if(error.message) {
    return error.message
  }
  else {
    return JSON.stringify(error)
  }
}


