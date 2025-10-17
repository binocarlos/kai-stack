import React from 'react'
import createRouter5, { Route, Router, State } from 'router5'
import { useRoute } from 'react-router5'
import browserPlugin from 'router5-plugin-browser'
import { User } from './types/app'

export interface IApplicationRouter {
  router: Router,
  getApplicationRoute: () => IApplicationRoute,
  RenderPage: () => JSX.Element,
  setCurrentUser: (user?: User) => void,
}

export type IRouteProcessor = (user: User | undefined, fromState: State, toState: State) => State

export interface IApplicationRoute extends Route {
  render: () => JSX.Element,
  meta: Record<string, any>,
  processRoute?: IRouteProcessor,
}

// Default not found route that can be overridden
export const DEFAULT_NOT_FOUND_ROUTE: IApplicationRoute = {
  name: 'notfound',
  path: '/notfound',
  meta: {
    title: 'Page Not Found',
  },
  render: () => <div>Page Not Found</div>,
}

// Configuration options for router creation
export interface RouterConfig {
  notFoundRoute?: IApplicationRoute
  defaultRouteName?: string
  queryParamsMode?: 'loose' | 'strict'
}

// Main function to create and configure the router
export function createRouter(routes: IApplicationRoute[], config?: RouterConfig): IApplicationRouter {
  const {
    notFoundRoute = DEFAULT_NOT_FOUND_ROUTE,
    defaultRouteName = 'notfound',
    queryParamsMode = 'loose',
  } = config || {}

  let currentUser: User | undefined

  // Combine application routes with not found route
  const allRoutes = [...routes, notFoundRoute]
  
  // Create and configure the router
  const router = createRouter5(allRoutes, {
    defaultRoute: defaultRouteName,
    queryParamsMode,
  })

  const getApplicationRoute = (): IApplicationRoute => {
    const { route } = useRoute()
    return allRoutes.find(r => r.name === route?.name) || notFoundRoute
  }

  const RenderPage = () => {
    const route = getApplicationRoute()
    return route.render()
  }

  const setCurrentUser = (user?: User) => {
    currentUser = user
  }

  router.usePlugin(browserPlugin())
  router.useMiddleware(() => (toState, fromState, done) => {
    const destRoute = allRoutes.find(r => r.name === toState.name)
    if(destRoute?.processRoute) {
      const newState = destRoute.processRoute(currentUser, fromState, toState)
      if(newState) {
        return done(null, newState)
      }
    }
    done()
  })

  return {
    router,
    getApplicationRoute,
    RenderPage,
    setCurrentUser,
  }
}

export default createRouter
