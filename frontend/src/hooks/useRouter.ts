import { useContext } from 'react'

import {
  CoreRouterContext,
} from '../contexts/router'

export const useRouter = () => {
  const router = useContext(CoreRouterContext)
  return router
}

export default useRouter