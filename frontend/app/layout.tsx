import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { ToastProvider } from '@/components/ui/use-toast'
import { Navigation } from '@/components/navigation'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'AdNexus SSP - Supply Side Platform',
  description: 'Manage publishers, placements, and private marketplace deals',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <ToastProvider>
          <div className="min-h-screen bg-gray-50">
            <Navigation />
            <main className="container mx-auto px-4 py-8">
              {children}
            </main>
          </div>
        </ToastProvider>
      </body>
    </html>
  )
}