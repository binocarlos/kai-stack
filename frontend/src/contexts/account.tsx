import { createContext, useCallback, useMemo, useEffect, useState, useContext, useRef } from 'react'
import axios, { AxiosResponse } from 'axios'
import useLoading from '../hooks/useLoading'
import useSnackbar from '../hooks/useSnackbar'
import useRouter from '../hooks/useRouter'

import {
  User,
  UserStatusResponse,
  LoginRequest,
  LoginResponse,
} from '../types/gotypes'

import {
  API_BASE_URL,
} from '../constants/system'

import { extractErrorMessage } from '../utils/apitools'

export const SESSION_STORAGE_KEY = 'stack_session_info'

export interface IAccountContext {
  initialized: boolean,
  loading: boolean,
  user?: User,
  onLogin: (username: string, password: string) => Promise<void>,
  onLogout: () => void,
}

export const AccountContext = createContext<IAccountContext>({
  initialized: false,
  loading: true,
  onLogin: async () => {},
  onLogout: () => {},
})

export const useAccount = () => {
  return useContext(AccountContext);
}

const loadUserFromLocalStorage = (): User | null => {
  const storedUser = sessionStorage.getItem(SESSION_STORAGE_KEY)
  if (!storedUser) return null
  return JSON.parse(storedUser) as User
}

const saveUserToLocalStorage = (user: User) => {
  sessionStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(user))
}

const clearUserFromLocalStorage = () => {
  sessionStorage.removeItem(SESSION_STORAGE_KEY)
}

const addTokenToAxios = (token: string) => {
  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
}

const clearTokenFromAxios = () => {
  delete axios.defaults.headers.common['Authorization']
}

const loadUserStatus = async (token: string): Promise<AxiosResponse<UserStatusResponse, any>> => {
  return axios.get<UserStatusResponse>(`${API_BASE_URL}/user/status`, {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  })
}

export const useAccountContext = (): IAccountContext => {
  const snackbar = useSnackbar()
  const loading = useLoading()
  const router = useRouter()

  const [ initialized, setInitialized ] = useState(false)
  const [ jobLoading, setJobLoading ] = useState(false)
  const [ user, setUser ] = useState<User>()

  const assignUserToState = (user: User) => {
    setUser(user)
    saveUserToLocalStorage(user)
    addTokenToAxios(user.token)
  }

  const clearUserFromState = () => {
    setUser(undefined)
    clearUserFromLocalStorage()
    clearTokenFromAxios()
    router.navigate('login')
  }

  // this is called both by the initialize method and the onLogin method
  const handleUserLoaded = async (user: User) => {
    assignUserToState(user)
    if (router.name === 'login') {
      router.navigate('home')
    }
  }

  // Main initialization sequence - runs once on mount
  const initialize = async (): Promise<void> => {
    const storedUser = loadUserFromLocalStorage()
    if (storedUser) {
      try {
        // if this fails then it means our cached login is invalid
        await loadUserStatus(storedUser.token)
        console.log('found user in local storage', storedUser)
        await handleUserLoaded(storedUser)
      } catch(e) {
        const message = extractErrorMessage(e)
        console.error('Error loading user status:', message)
        clearUserFromState()
      }
    } else {
      clearUserFromState()
    }
    loading.setLoading(false)
    setInitialized(true)
  }

  // call the login method
  const onLogin = async (email: string, password: string) => {
    loading.setLoading(true)
    try {
      const loginResponse = await axios.post<LoginResponse>(`${API_BASE_URL}/user/login`, {
        email,
        password,
      } as LoginRequest)
      const statusResponse = await loadUserStatus(loginResponse.data.token)
      
      const loggedInUser: User = {
        email,
        token: loginResponse.data.token,
        user_id: statusResponse.data.user_id, 
      }

      console.log(`user logged in`, loggedInUser)
      await handleUserLoaded(loggedInUser)
      
      snackbar.success(`Logged in as ${email}`)
    } catch (e) {
      const message = extractErrorMessage(e)
      snackbar.error(message)
      console.error('Error logging in:', e)
      clearUserFromState()
    } finally {
      loading.setLoading(false)
    }
  }

  const onLogout = async () => {
    try {
      await axios.post(`${API_BASE_URL}/user/logout`)  
    } catch(e) {
      
    }
    clearUserFromState()
    snackbar.success('Logged out')
    router.navigate('login')
  }

  useEffect(() => {
    initialize()
  }, [])

  return {
    initialized,
    loading: jobLoading,
    user,
    onLogin,
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