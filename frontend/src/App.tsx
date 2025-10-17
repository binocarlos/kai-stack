import React, { useEffect } from 'react'
import { ThemeProvider } from '@mui/material/styles'
import CssBaseline from '@mui/material/CssBaseline'
import { RouterProvider } from 'react-router5'
import {
  QueryClient,
  QueryClientProvider,
} from '@tanstack/react-query'

import createRouter from './router'
import { RouterContextProvider } from './contexts/router'
import { SnackbarContextProvider } from './contexts/snackbar'
import { LoadingContextProvider } from './contexts/loading'
import { AccountContextProvider } from './contexts/account'

import Layout from './pages/Layout'
import routes from './routes'
import theme from './theme'
import useAccount from './hooks/useAccount'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      gcTime: 10 * 60 * 1000,
    },
  },
})

const AppRouter = createRouter(routes)
const {
  router,
  setCurrentUser,
  RenderPage,
} = AppRouter

const AuthSync: React.FC = () => {
  const account = useAccount()
  
  useEffect(() => {
    setCurrentUser(account.user)
  }, [account.user])
  
  return null
}

router.start()

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <RouterProvider router={router}>
          <RouterContextProvider router={AppRouter}>
            <SnackbarContextProvider>
              <LoadingContextProvider>
                <AccountContextProvider>
                  <AuthSync />
                  <Layout>
                    <RenderPage />
                  </Layout>
                </AccountContextProvider>
              </LoadingContextProvider>
            </SnackbarContextProvider>
          </RouterContextProvider>
        </RouterProvider>
      </ThemeProvider>
    </QueryClientProvider>
  )
}
