import { createClient, SupabaseClient } from '@supabase/supabase-js'
import axios from 'axios'

// The Supabase client is only created when configured. Leaving it null lets the
// stack run on the local fixed-password path alone (dev/CI) without Supabase
// env vars. The anon key is public and safe to embed in the bundle.
const supabaseUrl = import.meta.env.VITE_SUPABASE_URL
const supabaseAnonKey = import.meta.env.VITE_SUPABASE_ANON_KEY

export const supabase: SupabaseClient | null =
  supabaseUrl && supabaseAnonKey ? createClient(supabaseUrl, supabaseAnonKey) : null

export const supabaseEnabled = supabase !== null

// Local fixed-password token (the JWT issued by POST /user/login). Supabase
// sessions are persisted/refreshed by supabase-js itself, so we only store this.
export const LOCAL_TOKEN_KEY = 'stack_local_token'

export const setLocalToken = (token: string) => sessionStorage.setItem(LOCAL_TOKEN_KEY, token)
export const clearLocalToken = () => sessionStorage.removeItem(LOCAL_TOKEN_KEY)

// getAuthToken returns the freshest bearer token to send to the API: a live
// Supabase access token if a session exists (supabase-js auto-refreshes it),
// otherwise the locally stored fixed-password token.
export const getAuthToken = async (): Promise<string | null> => {
  if (supabase) {
    const { data } = await supabase.auth.getSession()
    if (data.session?.access_token) return data.session.access_token
  }
  return sessionStorage.getItem(LOCAL_TOKEN_KEY)
}

// Attach the current token to every API request. Registered once at import.
axios.interceptors.request.use(async (config) => {
  const token = await getAuthToken()
  if (token) {
    config.headers = config.headers || {}
    config.headers['Authorization'] = `Bearer ${token}`
  }
  return config
})
