import React from 'react'
import { IApplicationRoute, IRouteProcessor } from './router'

import Login from './pages/Login'
import Home from './pages/Home'

const requireUser: IRouteProcessor = (user, fromState, toState) => {
  return user ? toState : {
    ...toState,
    name: 'home',
  }
}

// Bypass authentication for POC/demo pages with fixture data
const bypassAuth: IRouteProcessor = (user, fromState, toState) => {
  // Always allow access to these pages
  return toState
}

const routes: IApplicationRoute[] = [{
  name: 'login',
  path: '/',
  meta: {
    title: 'Login',
  },
  render: () => <Login />,
}, {
  name: 'home',
  path: '/home',
  meta: {
    title: 'Home',
  },
  render: () => <Home />,
}]

export default routes
