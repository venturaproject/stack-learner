import type { ReactNode } from 'react'

interface AppLayoutProps {
  children: ReactNode
}

const AppLayout = ({ children }: AppLayoutProps) => {
  return <>{children}</>
}

export { AppLayout }
