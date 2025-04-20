import { createContext, useState, useContext, useCallback } from 'react'

const AppContext = createContext()

export function AppProvider({ children }) {
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)

    const startLoading = useCallback(() => setLoading(true), [])
    const stopLoading = useCallback(() => setLoading(false), [])

    return (
        <AppContext.Provider value={{
            loading,
            startLoading,
            stopLoading,
            error,
            setError
        }}>
            {children}
            {loading && (
                <div className="fixed top-4 right-4 z-50">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
                </div>
            )}
        </AppContext.Provider>
    )
}

export const useAppContext = () => useContext(AppContext)