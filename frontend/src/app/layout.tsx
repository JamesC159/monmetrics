import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'MonMetrics - Professional Trading Card Analysis',
  description: 'Track, analyze, and predict trading card prices with advanced technical indicators. Make informed investment decisions with 5 years of market data.',
  keywords: 'trading cards, price analysis, Pokemon, Magic the Gathering, Yu-Gi-Oh, investment, technical analysis',
  authors: [{ name: 'MonMetrics Team' }],
  viewport: 'width=device-width, initial-scale=1',
  robots: 'index, follow',
  openGraph: {
    title: 'MonMetrics - Professional Trading Card Analysis',
    description: 'Professional trading analysis for the modern card collector',
    type: 'website',
    siteName: 'MonMetrics',
    images: [
      {
        url: '/og-image.jpg',
        width: 1200,
        height: 630,
        alt: 'MonMetrics - Trading Card Analysis Platform',
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'MonMetrics - Professional Trading Card Analysis',
    description: 'Professional trading analysis for the modern card collector',
    images: ['/og-image.jpg'],
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <head>
        <link rel="icon" href="/favicon.ico" />
        <link rel="apple-touch-icon" href="/apple-touch-icon.png" />
        <link rel="manifest" href="/manifest.json" />
        <meta name="theme-color" content="#1e40af" />
      </head>
      <body className={inter.className}>
        <div id="root">
          {children}
        </div>
      </body>
    </html>
  )
}