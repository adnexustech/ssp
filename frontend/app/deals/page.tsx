'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { Plus, Edit, Trash, DollarSign, Users, TrendingUp, Shield } from 'lucide-react'
import { useToast } from '@/components/ui/use-toast'
import { fetchDeals, createDeal, updateDeal, deleteDeal } from '@/lib/api'

interface Deal {
  id: string
  name: string
  advertiserId: string
  advertiserName?: string
  publisherId: string
  publisherName?: string
  fixedCpm: number
  currency: string
  impressionCap?: number
  startDate?: string
  endDate?: string
  targeting?: {
    sizes?: string[]
    devices?: string[]
    geos?: string[]
  }
  status?: 'active' | 'paused' | 'ended' | string
  impressions?: number
  revenue?: number
  adnexusId?: string // BidsCube DSP deal ID
}

export default function DealsPage() {
  const { toast } = useToast()
  const [deals, setDeals] = useState<Deal[]>([])
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [selectedDeal, setSelectedDeal] = useState<Deal | null>(null)
  
  // Form state
  const [formData, setFormData] = useState({
    name: '',
    advertiserId: '',
    publisherId: '',
    fixedCpm: 5.00,
    currency: 'USD',
    impressionCap: '',
    startDate: new Date().toISOString().split('T')[0],
    endDate: '',
    sizes: [] as string[],
    devices: [] as string[],
    geos: [] as string[],
    adnexusId: '', // For BidsCube integration
  })

  useEffect(() => {
    loadDeals()
  }, [])

  const loadDeals = async () => {
    const data = await fetchDeals()
    setDeals(data)
  }

  const handleCreate = async () => {
    try {
      const deal = await createDeal({
        ...formData,
        targeting: {
          sizes: formData.sizes,
          devices: formData.devices,
          geos: formData.geos,
        },
      })
      
      toast({
        title: 'Deal Created',
        description: `Deal ID: ${deal.id} has been created successfully.`,
      })
      
      setIsCreateOpen(false)
      loadDeals()
      resetForm()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to create deal.',
        variant: 'destructive',
      })
    }
  }

  const handleUpdate = async () => {
    if (!selectedDeal) return
    
    try {
      await updateDeal(selectedDeal.id, formData)
      toast({
        title: 'Deal Updated',
        description: 'The deal has been updated successfully.',
      })
      setSelectedDeal(null)
      loadDeals()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to update deal.',
        variant: 'destructive',
      })
    }
  }

  const handleDelete = async (dealId: string) => {
    if (!confirm('Are you sure you want to delete this deal?')) return
    
    try {
      await deleteDeal(dealId)
      toast({
        title: 'Deal Deleted',
        description: 'The deal has been deleted successfully.',
      })
      loadDeals()
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete deal.',
        variant: 'destructive',
      })
    }
  }

  const resetForm = () => {
    setFormData({
      name: '',
      advertiserId: '',
      publisherId: '',
      fixedCpm: 5.00,
      currency: 'USD',
      impressionCap: '',
      startDate: new Date().toISOString().split('T')[0],
      endDate: '',
      sizes: [],
      devices: [],
      geos: [],
      adnexusId: '',
    })
  }

  // Mock data for demonstration
  const mockDeals: Deal[] = [
    {
      id: 'PMP-2024-001',
      name: 'Premium Sports Content Deal',
      advertiserId: 'adv-001',
      advertiserName: 'Nike Sports',
      publisherId: 'pub-002',
      publisherName: 'Sports Daily',
      fixedCpm: 12.50,
      currency: 'USD',
      impressionCap: 1000000,
      startDate: '2024-01-01',
      endDate: '2024-12-31',
      targeting: {
        sizes: ['728x90', '300x250'],
        devices: ['web', 'mobile'],
        geos: ['US', 'CA'],
      },
      status: 'active',
      impressions: 450000,
      revenue: 5625.00,
      adnexusId: 'BC-DEAL-123',
    },
    {
      id: 'PMP-2024-002',
      name: 'Tech Audience Exclusive',
      advertiserId: 'adv-002',
      advertiserName: 'Apple Inc.',
      publisherId: 'pub-001',
      publisherName: 'TechNews Media',
      fixedCpm: 15.00,
      currency: 'USD',
      impressionCap: 500000,
      startDate: '2024-02-01',
      endDate: '2024-06-30',
      targeting: {
        sizes: ['300x250', '320x50'],
        devices: ['mobile'],
        geos: ['US'],
      },
      status: 'active',
      impressions: 225000,
      revenue: 3375.00,
      adnexusId: 'BC-DEAL-456',
    },
    {
      id: 'PMP-2024-003',
      name: 'CTV Holiday Campaign',
      advertiserId: 'adv-003',
      advertiserName: 'Amazon Prime',
      publisherId: 'pub-003',
      publisherName: 'Entertainment Hub',
      fixedCpm: 25.00,
      currency: 'USD',
      startDate: '2024-11-01',
      endDate: '2024-12-31',
      targeting: {
        sizes: ['1920x1080'],
        devices: ['ctv'],
        geos: ['US', 'UK', 'DE'],
      },
      status: 'paused',
      impressions: 0,
      revenue: 0,
      adnexusId: 'BC-DEAL-789',
    },
  ]

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Private Marketplace Deals</h1>
          <p className="text-gray-600 mt-2">
            Manage PMP deals with fixed pricing for premium inventory
          </p>
        </div>
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Create Deal
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>Create New PMP Deal</DialogTitle>
              <DialogDescription>
                Set up a private marketplace deal with fixed CPM pricing
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Deal Name</Label>
                  <Input
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="e.g., Premium Inventory Q1"
                  />
                </div>
                <div className="space-y-2">
                  <Label>BidsCube Deal ID</Label>
                  <Input
                    value={formData.adnexusId}
                    onChange={(e) => setFormData({ ...formData, adnexusId: e.target.value })}
                    placeholder="e.g., BC-DEAL-XXX"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Publisher</Label>
                  <Select value={formData.publisherId} onValueChange={(v) => setFormData({ ...formData, publisherId: v })}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select publisher" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="pub-001">TechNews Media</SelectItem>
                      <SelectItem value="pub-002">Sports Daily</SelectItem>
                      <SelectItem value="pub-003">Entertainment Hub</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>Advertiser</Label>
                  <Select value={formData.advertiserId} onValueChange={(v) => setFormData({ ...formData, advertiserId: v })}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select advertiser" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="adv-001">Nike Sports</SelectItem>
                      <SelectItem value="adv-002">Apple Inc.</SelectItem>
                      <SelectItem value="adv-003">Amazon Prime</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div className="space-y-2">
                  <Label>Fixed CPM</Label>
                  <Input
                    type="number"
                    step="0.01"
                    value={formData.fixedCpm}
                    onChange={(e) => setFormData({ ...formData, fixedCpm: parseFloat(e.target.value) })}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Currency</Label>
                  <Select value={formData.currency} onValueChange={(v) => setFormData({ ...formData, currency: v })}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="USD">USD</SelectItem>
                      <SelectItem value="EUR">EUR</SelectItem>
                      <SelectItem value="GBP">GBP</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>Impression Cap</Label>
                  <Input
                    type="number"
                    value={formData.impressionCap}
                    onChange={(e) => setFormData({ ...formData, impressionCap: e.target.value })}
                    placeholder="Optional"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>Start Date</Label>
                  <Input
                    type="date"
                    value={formData.startDate}
                    onChange={(e) => setFormData({ ...formData, startDate: e.target.value })}
                  />
                </div>
                <div className="space-y-2">
                  <Label>End Date</Label>
                  <Input
                    type="date"
                    value={formData.endDate}
                    onChange={(e) => setFormData({ ...formData, endDate: e.target.value })}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label>Ad Sizes</Label>
                <div className="flex flex-wrap gap-2">
                  {['728x90', '300x250', '320x50', '160x600', '1920x1080'].map(size => (
                    <Badge
                      key={size}
                      variant={formData.sizes.includes(size) ? 'default' : 'outline'}
                      className="cursor-pointer"
                      onClick={() => {
                        setFormData({
                          ...formData,
                          sizes: formData.sizes.includes(size)
                            ? formData.sizes.filter(s => s !== size)
                            : [...formData.sizes, size]
                        })
                      }}
                    >
                      {size}
                    </Badge>
                  ))}
                </div>
              </div>

              <div className="space-y-2">
                <Label>Devices</Label>
                <div className="flex gap-2">
                  {['web', 'mobile', 'ctv'].map(device => (
                    <Badge
                      key={device}
                      variant={formData.devices.includes(device) ? 'default' : 'outline'}
                      className="cursor-pointer"
                      onClick={() => {
                        setFormData({
                          ...formData,
                          devices: formData.devices.includes(device)
                            ? formData.devices.filter(d => d !== device)
                            : [...formData.devices, device]
                        })
                      }}
                    >
                      {device.toUpperCase()}
                    </Badge>
                  ))}
                </div>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsCreateOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleCreate}>
                Create Deal
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Deals</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {mockDeals.filter(d => d.status === 'active').length}
            </div>
            <p className="text-xs text-muted-foreground">Currently running</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Revenue</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${mockDeals.reduce((sum, d) => sum + (d.revenue || 0), 0).toFixed(2)}
            </div>
            <p className="text-xs text-muted-foreground">All time</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Avg CPM</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${(mockDeals.reduce((sum, d) => sum + d.fixedCpm, 0) / mockDeals.length).toFixed(2)}
            </div>
            <p className="text-xs text-muted-foreground">Across all deals</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">BidsCube Integrated</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {mockDeals.filter(d => d.adnexusId).length}
            </div>
            <p className="text-xs text-muted-foreground">Connected deals</p>
          </CardContent>
        </Card>
      </div>

      {/* Deals Table */}
      <Card>
        <CardHeader>
          <CardTitle>All Deals</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b">
                  <th className="text-left py-2 px-4">Deal ID</th>
                  <th className="text-left py-2 px-4">Name</th>
                  <th className="text-left py-2 px-4">Publisher</th>
                  <th className="text-left py-2 px-4">Advertiser</th>
                  <th className="text-left py-2 px-4">Fixed CPM</th>
                  <th className="text-left py-2 px-4">Status</th>
                  <th className="text-left py-2 px-4">Impressions</th>
                  <th className="text-left py-2 px-4">Revenue</th>
                  <th className="text-left py-2 px-4">Actions</th>
                </tr>
              </thead>
              <tbody>
                {mockDeals.map(deal => (
                  <tr key={deal.id} className="border-b hover:bg-gray-50">
                    <td className="py-2 px-4">
                      <code className="text-sm bg-gray-100 px-2 py-1 rounded">
                        {deal.id}
                      </code>
                      {deal.adnexusId && (
                        <Badge variant="outline" className="ml-2 text-xs">
                          BC
                        </Badge>
                      )}
                    </td>
                    <td className="py-2 px-4">{deal.name}</td>
                    <td className="py-2 px-4">{deal.publisherName || 'N/A'}</td>
                    <td className="py-2 px-4">{deal.advertiserName || 'N/A'}</td>
                    <td className="py-2 px-4">${deal.fixedCpm.toFixed(2)}</td>
                    <td className="py-2 px-4">
                      <Badge variant={deal.status === 'active' ? 'default' : 'secondary'}>
                        {deal.status || 'unknown'}
                      </Badge>
                    </td>
                    <td className="py-2 px-4">{(deal.impressions || 0).toLocaleString()}</td>
                    <td className="py-2 px-4">${(deal.revenue || 0).toFixed(2)}</td>
                    <td className="py-2 px-4">
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => setSelectedDeal(deal)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleDelete(deal.id)}
                        >
                          <Trash className="h-4 w-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}