/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  env: {
    NEXT_PUBLIC_SSP_URL: process.env.SSP_URL || 'https://ssp.ad.nexus',
    NEXT_PUBLIC_API_URL: process.env.API_URL || 'http://localhost:8082/api',
  },
}

module.exports = nextConfig