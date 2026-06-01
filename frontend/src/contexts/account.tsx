import { createContext, useEffect, useState, useContext } from 'react'
import axios from 'axios'
import useLoading from '../hooks/useLoading'
import useSnackbar from '../hooks/useSnackbar'
import useRouter from '../hooks/useRouter'

import {
  User,
  UserStatusResponse,
  LoginResponse,
  LoginRequest,
} from '../types/gotypes'

import {
  API_BASE_URL,
} from '../constants/system'

import {
  supabase,
  supabaseEnabled,
  setLocalToken,
  clearLocalToken,
  getAuthToken,
} from '../supabase'

import { extractErrorMessage } from '../utils/apitools'

export interface IAccountContext {
  initialized: boolean,
  loading: boolean,
  supabaseEnabled: boolean,
  user?: User,
  onLogin: (email: string, password: string) => Promise<void>,
  onLoginWithGoogle: () => Promise<void>,
  onLogout: () => void,
}

export const AccountContext = createContext<IAccountContext>({
  initialized: false,
  loading: true,
  supabaseEnabled,
  onLogin: async () => {},
  onLoginWithGoogle: async () => {},
  onLogout: () => {},
})

export const useAccount = () => {
  return useContext(AccountContext);
}

// loadUser fetches the canonical profile for whatever bearer token is currently
// active (the axios interceptor attaches it). Returns null if not authenticated.
const loadUser = async (): Promise<User | null> => {
  const token = await getAuthToken()
  if (!token) return null
  const { data } = await axios.get<UserStatusResponse>(`${API_BASE_URL}/user/status`)
  return {
    user_id: data.user_id,
    email: data.email,
    token,
  }
}

export const useAccountContext = (): IAccountContext => {
  const snackbar = useSnackbar()
  const loading = useLoading()
  const router = useRouter()

  const [ initialized, setInitialized ] = useState(false)
  const [ user, setUser ] = useState<User>()

  const handleUserLoaded = (loadedUser: User) => {
    setUser(loadedUser)
    if (router.name === 'login') {
      router.navigate('home')
    }
  }

  const clearUser = (navigateToLogin: boolean = true) => {
    setUser(undefined)
    clearLocalToken()
    if (navigateToLogin) router.navigate('login')
  }

  // refresh re-reads the active session/token and loads the profile.
  const refresh = async (navigateOnEmpty: boolean = true) => {
    try {
      const loadedUser = await loadUser()
      if (loadedUser) {
        handleUserLoaded(loadedUser)
      } else {
        clearUser(navigateOnEmpty)
      }
    } catch (e) {
      console.error('Error loading user status:', extractErrorMessage(e))
      clearUser(navigateOnEmpty)
    }
  }

  // Initialization: pick up an existing Supabase session or local token, and
  // subscribe to Supabase auth changes (covers the OAuth redirect on return).
  const initialize = async (): Promise<void> => {
    await refresh(false)

    if (supabase) {
      supabase.auth.onAuthStateChange((_event, session) => {
        if (session) {
          refresh(false)
        } else {
          clearUser()
        }
      })
    }

    loading.setLoading(false)
    setInitialized(true)
  }

  // Local fixed-password login (dev/CI).
  const onLogin = async (email: string, password: string) => {
    loading.setLoading(true)
    try {
      const loginResponse = await axios.post<LoginResponse>(`${API_BASE_URL}/user/login`, {
        email,
        password,
      } as LoginRequest)
      setLocalToken(loginResponse.data.token)
      await refresh()
      snackbar.success(`Logged in as ${email}`)
    } catch (e) {
      snackbar.error(extractErrorMessage(e))
      console.error('Error logging in:', e)
      clearUser()
    } finally {
      loading.setLoading(false)
    }
  }

  // Google sign-in via Supabase. supabase-js runs the OAuth redirect and, on
  // return to the app, fires onAuthStateChange which drives refresh().
  const onLoginWithGoogle = async () => {
    if (!supabase) {
      snackbar.error('Google sign-in is not configured')
      return
    }
    const { error } = await supabase.auth.signInWithOAuth({
      provider: 'google',
      options: { redirectTo: window.location.origin },
    })
    if (error) {
      snackbar.error(extractErrorMessage(error))
    }
  }

  const onLogout = async () => {
    if (supabase) {
      await supabase.auth.signOut()
    } else {
      try {
        await axios.post(`${API_BASE_URL}/user/logout`)
      } catch (e) {}
    }
    clearUser()
    snackbar.success('Logged out')
  }

  useEffect(() => {
    initialize()
  }, [])

  return {
    initialized,
    loading: loading.loading,
    supabaseEnabled,
    user,
    onLogin,
    onLoginWithGoogle,
    onLogout,
  }
}

export const AccountContextProvider: React.FC<React.PropsWithChildren> = ({ children }) => {
  const value = useAccountContext()
  return (
    <AccountContext.Provider value={value}>
      {children}
    </AccountContext.Provider>
  )
}
