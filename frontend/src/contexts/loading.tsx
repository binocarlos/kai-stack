import React, { FC, createContext, useMemo, useState, PropsWithChildren } from 'react'

export interface ILoadingContext {
  loading: boolean,
  progressTitle: string,
  progressTotal: number,
  progressCurrent: number,
  setLoading: {
    (val: boolean): void,
  },
  setProgress: {
    (title: string, total: number, current: number): void,
  },
}

export const LoadingContext = createContext<ILoadingContext>({
  loading: false,
  progressTitle: '',
  progressTotal: 0,
  progressCurrent: 0,
  setLoading: () => {},
  setProgress: () => {},
})

export const useLoadingContext = (): ILoadingContext => {
  // we start off with loading=true because we are assuming that the userContext
  // will initialize and this will prevent the flicker of off -> on -> off
  const [ loading, setLoading ] = useState(true)
  const [ progressTitle, setProgressTitle ] = useState('')
  const [ progressTotal, setProgressTotal ] = useState(0)
  const [ progressCurrent, setProgressCurrent ] = useState(0)

  const setProgress = (title: string, total: number, current: number) => {
    setProgressTitle(title)
    setProgressTotal(total)
    setProgressCurrent(current)
  }

  const contextValue = useMemo<ILoadingContext>(() => ({
    loading,
    progressTitle,
    progressTotal,
    progressCurrent,
    setLoading,
    setProgress,
  }), [
    loading,
    progressTitle,
    progressTotal,
    progressCurrent,
    setLoading,
    setProgress,
  ])
  return contextValue
}

export const LoadingContextProvider: FC<PropsWithChildren> = ({ children }) => {
  // Get the loading context value
  const value = useLoadingContext()

  return (
    <LoadingContext.Provider value={value}>
      {children}
    </LoadingContext.Provider>
  );
};
