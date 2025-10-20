'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ArrowRight, DollarSign, Globe, Shield, Users } from 'lucide-react'
import Link from 'next/link'
import { useEffect, useState } from 'react'
import { fetchMetrics } from '@/lib/api'

export default function Dashboard() {
  const [metrics, setMetrics] = useState({
    publishers: 0,
    sites: 0,
    impressions: 0,
    revenue: 0,
    deals: 0,
    fillRate: 0,
  })

  useEffect(() => {
    fetchMetrics().then(setMetrics)
  }, [])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">SSP Dashboard</h1>
        <p className="text-gray-600 mt-2">
          Manage publishers, create tags, and configure private marketplace deals
        </p>
      </div>

      {/* Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Publishers</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics.publishers}</div>
            <p className="text-xs text-muted-foreground">Active publishers</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Sites</CardTitle>
            <Globe className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics.sites}</div>
            <p className="text-xs text-muted-foreground">Connected sites</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Revenue</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${metrics.revenue.toFixed(2)}</div>
            <p className="text-xs text-muted-foreground">Today's revenue</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">PMP Deals</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics.deals}</div>
            <p className="text-xs text-muted-foreground">Active deals</p>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card>
          <CardHeader>
            <CardTitle>Publishers</CardTitle>
            <CardDescription>
              Manage publisher accounts and settings
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/publishers">
              <Button className="w-full">
                Manage Publishers
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </Link>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Tag Generator</CardTitle>
            <CardDescription>
              Create ad tags for web, mobile, and CTV
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/tags">
              <Button className="w-full">
                Generate Tags
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </Link>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Private Marketplace</CardTitle>
            <CardDescription>
              Configure PMP deals with fixed pricing
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Link href="/deals">
              <Button className="w-full">
                Manage Deals
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Activity</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">New Publisher Added</p>
                <p className="text-xs text-gray-500">TechNews Media - 2 minutes ago</p>
              </div>
              <Button size="sm" variant="outline">View</Button>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">PMP Deal Created</p>
                <p className="text-xs text-gray-500">Deal ID: PMP-2024-001 - $5.00 CPM - 1 hour ago</p>
              </div>
              <Button size="sm" variant="outline">View</Button>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">Tag Generated</p>
                <p className="text-xs text-gray-500">Mobile Banner 320x50 - 3 hours ago</p>
              </div>
              <Button size="sm" variant="outline">View</Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}