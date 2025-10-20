'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Copy, Code, Smartphone, Tv, Monitor } from 'lucide-react'
import { useToast } from '@/components/ui/use-toast'
import CodeEditor from '@/components/code-editor'

const DEVICE_TYPES = {
  web: { name: 'Web', icon: Monitor, sizes: ['728x90', '300x250', '320x50', '160x600', '970x250'] },
  mobile: { name: 'Mobile', icon: Smartphone, sizes: ['320x50', '300x50', '320x100', '300x250'] },
  ctv: { name: 'CTV/OTT', icon: Tv, sizes: ['1920x1080', '1280x720', '640x480'] },
  publica: { name: 'Publica', icon: Tv, sizes: ['1920x1080', '1280x720', '640x480'] },
}

export default function TagGenerator() {
  const { toast } = useToast()
  const [selectedPublisher, setSelectedPublisher] = useState('')
  const [selectedSite, setSelectedSite] = useState('')
  const [selectedDevice, setSelectedDevice] = useState('web')
  const [selectedSize, setSelectedSize] = useState('300x250')
  const [dealId, setDealId] = useState('')
  const [floorPrice, setFloorPrice] = useState('0.50')
  const [generatedTag, setGeneratedTag] = useState('')

  const generateTag = () => {
    const baseUrl = process.env.NEXT_PUBLIC_SSP_URL || 'https://ssp.ad.nexus'
    const tagId = `tag_${Date.now()}`
    
    const params = new URLSearchParams({
      pub: selectedPublisher,
      site: selectedSite,
      size: selectedSize,
      device: selectedDevice,
      floor: floorPrice,
      ...(dealId && { deal: dealId }),
    })

    let tag = ''

    if (selectedDevice === 'publica') {
      // Publica SSAI (Server-Side Ad Insertion) Tag
      tag = `<!-- AdNexus SSP Publica SSAI Integration -->
<!-- Replace existing Publica tag with this configuration -->
{
  "ssai_endpoint": "${baseUrl}/publica/ssai",
  "publisher_id": "${selectedPublisher}",
  "site_id": "${selectedSite}",
  "content_id": "[CONTENT_ID]",
  "device_id": "[DEVICE_ID]",
  "ip": "[IP_ADDRESS]",
  "ua": "[USER_AGENT]",
  "floor_price": ${floorPrice},${dealId ? `\n  "deal_id": "${dealId}",` : ''}
  "params": {
    "size": "${selectedSize}",
    "content_genre": "[CONTENT_GENRE]",
    "content_rating": "[CONTENT_RATING]",
    "content_language": "[CONTENT_LANGUAGE]"
  },
  "vast_url": "${baseUrl}/publica/vast?${params}",
  "tracking": {
    "impression": "${baseUrl}/publica/pixel/impression?${params}",
    "click": "${baseUrl}/publica/click?${params}",
    "complete": "${baseUrl}/publica/pixel/complete?${params}"
  }
}

<!-- Publica Macro Replacement URL -->
${baseUrl}/publica/vast?pub=${selectedPublisher}&site=${selectedSite}&content_id=$$CONTENT_ID$$&device_id=$$DEVICE_ID$$&ip=$$IP_ADDRESS$$&ua=$$USER_AGENT$$&floor=${floorPrice}${dealId ? `&deal=${dealId}` : ''}

<!-- Integration Steps for P1's Publica Instance -->
1. Replace the existing ad decisioning endpoint in Publica with: ${baseUrl}/publica/ssai
2. Configure Publica to pass the required macros: CONTENT_ID, DEVICE_ID, IP_ADDRESS, USER_AGENT
3. Set the floor price to ${floorPrice} CPM
${dealId ? `4. Configure PMP deal ID: ${dealId} for fixed pricing` : ''}
5. Test with your CTV app to ensure proper ad insertion`
    } else if (selectedDevice === 'ctv') {
      // VAST tag for CTV/OTT
      tag = `<!-- AdNexus SSP VAST Tag -->
<VAST version="3.0">
  <Ad id="${tagId}">
    <InLine>
      <AdSystem>AdNexus SSP</AdSystem>
      <AdTitle>Video Ad</AdTitle>
      <Impression><![CDATA[${baseUrl}/pixel/impression?${params}]]></Impression>
      <Creatives>
        <Creative>
          <Linear>
            <MediaFiles>
              <MediaFile delivery="progressive" type="video/mp4" width="1920" height="1080">
                <![CDATA[${baseUrl}/vast/video?${params}]]>
              </MediaFile>
            </MediaFiles>
            <VideoClicks>
              <ClickThrough><![CDATA[${baseUrl}/click?${params}]]></ClickThrough>
            </VideoClicks>
          </Linear>
        </Creative>
      </Creatives>
    </InLine>
  </Ad>
</VAST>`
    } else if (selectedDevice === 'mobile') {
      // Mobile SDK tag
      tag = `// AdNexus SSP Mobile SDK Integration
// iOS Swift
let adView = AdNexusAdView(
  publisherId: "${selectedPublisher}",
  siteId: "${selectedSite}",
  size: CGSize(width: ${selectedSize.split('x')[0]}, height: ${selectedSize.split('x')[1]}),
  floorPrice: ${floorPrice}${dealId ? `,\n  dealId: "${dealId}"` : ''}
)
adView.loadAd()

// Android Kotlin
val adView = AdNexusAdView(
  context = this,
  publisherId = "${selectedPublisher}",
  siteId = "${selectedSite}",
  adSize = AdSize(${selectedSize.split('x')[0]}, ${selectedSize.split('x')[1]}),
  floorPrice = ${floorPrice}f${dealId ? `,\n  dealId = "${dealId}"` : ''}
)
adView.loadAd()`
    } else {
      // Web display tag
      tag = `<!-- AdNexus SSP Display Tag - ${selectedSize} -->
<div id="adnexus-${tagId}"></div>
<script async src="${baseUrl}/tag.js"></script>
<script>
  window.adnexusQueue = window.adnexusQueue || [];
  window.adnexusQueue.push(function() {
    adnexus.display({
      tagId: 'adnexus-${tagId}',
      publisherId: '${selectedPublisher}',
      siteId: '${selectedSite}',
      size: '${selectedSize}',
      floorPrice: ${floorPrice}${dealId ? `,\n      dealId: '${dealId}'` : ''},
      device: '${selectedDevice}'
    });
  });
</script>
<!-- End AdNexus SSP Tag -->`
    }

    setGeneratedTag(tag)
  }

  const copyTag = () => {
    navigator.clipboard.writeText(generatedTag)
    toast({
      title: 'Tag Copied!',
      description: 'The ad tag has been copied to your clipboard.',
    })
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Tag Generator</h1>
        <p className="text-gray-600 mt-2">
          Create ad tags for different devices and placements
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Configuration */}
        <Card>
          <CardHeader>
            <CardTitle>Tag Configuration</CardTitle>
            <CardDescription>
              Select publisher, placement, and targeting options
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Publisher Selection */}
            <div className="space-y-2">
              <Label>Publisher</Label>
              <Select value={selectedPublisher} onValueChange={setSelectedPublisher}>
                <SelectTrigger>
                  <SelectValue placeholder="Select publisher" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="p1-publica">P1 (Publica Instance)</SelectItem>
                  <SelectItem value="pub-001">TechNews Media</SelectItem>
                  <SelectItem value="pub-002">Sports Daily</SelectItem>
                  <SelectItem value="pub-003">Entertainment Hub</SelectItem>
                  <SelectItem value="demo-pub">Demo Publisher</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Site Selection */}
            <div className="space-y-2">
              <Label>Site</Label>
              <Select value={selectedSite} onValueChange={setSelectedSite}>
                <SelectTrigger>
                  <SelectValue placeholder="Select site" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="site-001">Main Website</SelectItem>
                  <SelectItem value="site-002">Mobile App</SelectItem>
                  <SelectItem value="site-003">CTV App</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Device Type */}
            <div className="space-y-2">
              <Label>Device Type</Label>
              <Tabs value={selectedDevice} onValueChange={setSelectedDevice}>
                <TabsList className="grid w-full grid-cols-4">
                  {Object.entries(DEVICE_TYPES).map(([key, device]) => (
                    <TabsTrigger key={key} value={key}>
                      <device.icon className="h-4 w-4 mr-2" />
                      {device.name}
                    </TabsTrigger>
                  ))}
                </TabsList>
              </Tabs>
            </div>

            {/* Ad Size */}
            <div className="space-y-2">
              <Label>Ad Size</Label>
              <Select value={selectedSize} onValueChange={setSelectedSize}>
                <SelectTrigger>
                  <SelectValue placeholder="Select size" />
                </SelectTrigger>
                <SelectContent>
                  {DEVICE_TYPES[selectedDevice as keyof typeof DEVICE_TYPES].sizes.map(size => (
                    <SelectItem key={size} value={size}>
                      {size} {size === '728x90' && '(Leaderboard)'}
                      {size === '300x250' && '(Medium Rectangle)'}
                      {size === '320x50' && '(Mobile Banner)'}
                      {size === '160x600' && '(Wide Skyscraper)'}
                      {size === '970x250' && '(Billboard)'}
                      {size === '1920x1080' && '(Full HD)'}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Floor Price */}
            <div className="space-y-2">
              <Label>Floor Price (CPM)</Label>
              <Input
                type="number"
                step="0.01"
                value={floorPrice}
                onChange={(e) => setFloorPrice(e.target.value)}
                placeholder="0.50"
              />
            </div>

            {/* PMP Deal ID */}
            <div className="space-y-2">
              <Label>PMP Deal ID (Optional)</Label>
              <Input
                value={dealId}
                onChange={(e) => setDealId(e.target.value)}
                placeholder="e.g., PMP-2024-001"
              />
              <p className="text-xs text-gray-500">
                Enter a deal ID for private marketplace with fixed pricing
              </p>
            </div>

            <Button onClick={generateTag} className="w-full">
              <Code className="mr-2 h-4 w-4" />
              Generate Tag
            </Button>
          </CardContent>
        </Card>

        {/* Generated Tag */}
        <Card>
          <CardHeader>
            <CardTitle>Generated Tag</CardTitle>
            <CardDescription>
              Copy this code and add it to your website
            </CardDescription>
          </CardHeader>
          <CardContent>
            {generatedTag ? (
              <div className="space-y-4">
                <CodeEditor
                  value={generatedTag}
                  language={selectedDevice === 'ctv' ? 'xml' : selectedDevice === 'publica' ? 'javascript' : 'html'}
                  readOnly
                />
                <Button onClick={copyTag} className="w-full">
                  <Copy className="mr-2 h-4 w-4" />
                  Copy Tag
                </Button>

                {/* Implementation Notes */}
                <div className="bg-blue-50 border border-blue-200 rounded p-4">
                  <h4 className="font-semibold text-blue-900 mb-2">Implementation Notes</h4>
                  <ul className="text-sm text-blue-800 space-y-1">
                    {selectedDevice === 'web' && (
                      <>
                        <li>• Place this code where you want the ad to appear</li>
                        <li>• The tag loads asynchronously for better performance</li>
                        <li>• Ensure the container div is not hidden by CSS</li>
                      </>
                    )}
                    {selectedDevice === 'mobile' && (
                      <>
                        <li>• Add the AdNexus SDK to your mobile app</li>
                        <li>• Initialize the SDK with your app credentials</li>
                        <li>• Test in both portrait and landscape modes</li>
                      </>
                    )}
                    {selectedDevice === 'ctv' && (
                      <>
                        <li>• Use this VAST tag in your video player</li>
                        <li>• Supports VAST 3.0 and 4.0 compatible players</li>
                        <li>• Test with different video durations</li>
                      </>
                    )}
                    {selectedDevice === 'publica' && (
                      <>
                        <li>• Replace existing Publica ad decisioning endpoint</li>
                        <li>• Configure Publica to pass required macros</li>
                        <li>• Supports server-side ad insertion (SSAI)</li>
                        <li>• Compatible with Publica's manifest manipulation</li>
                        <li>• Test with live and VOD content</li>
                      </>
                    )}
                    {dealId && (
                      <li className="font-semibold">
                        • PMP Deal "{dealId}" will be prioritized with fixed pricing
                      </li>
                    )}
                  </ul>
                </div>
              </div>
            ) : (
              <div className="text-center py-12 text-gray-500">
                Configure your tag settings and click "Generate Tag"
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}