'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'
import { Home, Users, Tag, Shield, BarChart, Settings } from 'lucide-react'

const navigation = [
  { name: 'Dashboard', href: '/', icon: Home },
  { name: 'Publishers', href: '/publishers', icon: Users },
  { name: 'Tag Generator', href: '/tags', icon: Tag },
  { name: 'PMP Deals', href: '/deals', icon: Shield },
  { name: 'Analytics', href: '/analytics', icon: BarChart },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export function Navigation() {
  const pathname = usePathname()

  return (
    <nav className="bg-white shadow">
      <div className="container mx-auto px-4">
        <div className="flex justify-between h-16">
          <div className="flex">
            <div className="flex-shrink-0 flex items-center">
              <Link href="/" className="text-2xl font-bold text-blue-600">
                AdNexus SSP
              </Link>
            </div>
            <div className="hidden sm:ml-8 sm:flex sm:space-x-8">
              {navigation.map((item) => {
                const Icon = item.icon
                return (
                  <Link
                    key={item.name}
                    href={item.href}
                    className={cn(
                      'inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium',
                      pathname === item.href
                        ? 'border-blue-500 text-gray-900'
                        : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'
                    )}
                  >
                    <Icon className="mr-2 h-4 w-4" />
                    {item.name}
                  </Link>
                )
              })}
            </div>
          </div>
          <div className="flex items-center">
            <span className="text-sm text-gray-500">
              Connected to: <span className="font-medium">BidsCube DSP</span>
            </span>
          </div>
        </div>
      </div>
    </nav>
  )
}