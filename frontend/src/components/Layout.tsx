import { ReactNode } from 'react'
import { Navigation } from './Navigation'
import { Footer } from './Footer'

interface LayoutProps {
  children: ReactNode
  showNavigation?: boolean
  showFooter?: boolean
  className?: string
}

export const Layout = ({
  children,
  showNavigation = true,
  showFooter = true,
  className = '',
}: LayoutProps) => {
  return (
    <div className='min-h-screen flex flex-col bg-dark-950'>
      {showNavigation && <Navigation />}

      <main className={`flex-1 ${showNavigation ? 'pt-16' : ''} ${className}`}>{children}</main>

      {showFooter && <Footer />}
    </div>
  )
}
