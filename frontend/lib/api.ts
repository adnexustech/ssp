const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8082/api'

export interface Publisher {
  id: string
  name: string
  email: string
  domain: string
  revShare: number
  active: boolean
  createdAt: string
}

export interface Deal {
  id: string
  name: string
  advertiserId: string
  publisherId: string
  fixedCpm: number
  currency: string
  status: string
  targeting: any
}

// Publishers
export async function fetchPublishers(): Promise<Publisher[]> {
  const res = await fetch(`${API_BASE}/publishers`)
  if (!res.ok) throw new Error('Failed to fetch publishers')
  return res.json()
}

export async function createPublisher(publisher: Partial<Publisher>): Promise<Publisher> {
  const res = await fetch(`${API_BASE}/publishers`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(publisher),
  })
  if (!res.ok) throw new Error('Failed to create publisher')
  return res.json()
}

// Deals
export async function fetchDeals(): Promise<Deal[]> {
  const res = await fetch(`${API_BASE}/deals`)
  if (!res.ok) throw new Error('Failed to fetch deals')
  return res.json()
}

export async function createDeal(deal: any): Promise<Deal> {
  const res = await fetch(`${API_BASE}/deals`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(deal),
  })
  if (!res.ok) throw new Error('Failed to create deal')
  return res.json()
}

export async function updateDeal(id: string, deal: any): Promise<Deal> {
  const res = await fetch(`${API_BASE}/deals/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(deal),
  })
  if (!res.ok) throw new Error('Failed to update deal')
  return res.json()
}

export async function deleteDeal(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/deals/${id}`, {
    method: 'DELETE',
  })
  if (!res.ok) throw new Error('Failed to delete deal')
}

// Metrics
export async function fetchMetrics() {
  // Mock data for now
  return {
    publishers: 12,
    sites: 45,
    impressions: 1234567,
    revenue: 12345.67,
    deals: 8,
    fillRate: 87.5,
  }
}