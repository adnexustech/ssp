# Adnexus Studio - CDN Deployment Guide

## Overview

Adnexus Studio is deployed to `cdn.ad.nexus/studio/` with DNS pointing `studio.ad.nexus` to the CDN.

## Quick Deploy

### Local Deployment

```bash
./deploy.sh
```

The script will prompt for deployment method:
1. **AWS S3/CloudFront** - Automated deployment to AWS infrastructure
2. **Custom CDN Server** - Deploy via rsync/scp to custom server
3. **Manual Tarball** - Create tarball for manual deployment

### GitHub Actions Deployment

Deployment workflow runs automatically on push to `main` branch:
- Builds the project
- Creates deployment artifact
- Ready for CDN upload (configure secrets for automation)

## Deployment Methods

### 1. AWS S3 + CloudFront

**GitHub Secrets Required:**
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `CLOUDFRONT_DISTRIBUTION_ID`

**Manual Deployment:**
```bash
# Build
npm run build-production

# Deploy to S3
aws s3 sync x/ s3://cdn.ad.nexus/studio/ \
  --delete \
  --cache-control "public,max-age=31536000,immutable" \
  --exclude "*.html" \
  --exclude "*.json"

# Deploy HTML/JSON with no-cache
aws s3 sync x/ s3://cdn.ad.nexus/studio/ \
  --exclude "*" \
  --include "*.html" \
  --include "*.json" \
  --cache-control "public,max-age=0,must-revalidate"

# Invalidate CloudFront
aws cloudfront create-invalidation \
  --distribution-id YOUR_DIST_ID \
  --paths "/studio/*"
```

### 2. Custom CDN Server

```bash
# Build
npm run build-production

# Deploy via rsync
rsync -avz --delete x/ user@cdn.ad.nexus:/var/www/studio/
```

### 3. Manual Deployment

```bash
# Build
npm run build-production

# Create tarball
cd x && tar -czf ../studio-dist.tar.gz . && cd ..

# Transfer to server
scp studio-dist.tar.gz user@cdn.ad.nexus:/tmp/

# Extract on server
ssh user@cdn.ad.nexus "cd /var/www/studio && tar -xzf /tmp/studio-dist.tar.gz"
```

## DNS Configuration

Point `studio.ad.nexus` to CDN:

**CloudFront:**
```
studio.ad.nexus. CNAME d123456789.cloudfront.net
```

**Custom Server:**
```
studio.ad.nexus. A 1.2.3.4
```

## Cache Headers

Optimal cache configuration:

**Static Assets** (JS, CSS, images):
```
Cache-Control: public, max-age=31536000, immutable
```

**HTML/JSON:**
```
Cache-Control: public, max-age=0, must-revalidate
```

## CORS Configuration

If Studio needs to access resources from other domains:

```json
{
  "AllowedOrigins": ["https://ad.nexus", "https://dsp.ad.nexus"],
  "AllowedMethods": ["GET", "HEAD"],
  "AllowedHeaders": ["*"],
  "MaxAgeSeconds": 3600
}
```

## Build Output Structure

```
x/
├── index.html              # Main entry point
├── index.css               # Styles
├── main.js                 # Application bundle
├── importmap.json          # Import map
├── coi-serviceworker.js    # Cross-Origin Isolation
├── assets/                 # Static assets
├── components/             # Component bundles
├── context/                # Context bundles
├── icons/                  # Icon libraries
└── tools/                  # Tool bundles
```

## Testing Deployment

1. **Local testing:**
   ```bash
   npm start
   # Visit http://localhost:8000
   ```

2. **Production testing:**
   ```bash
   cd x && npx http-server -p 8080
   # Visit http://localhost:8080
   ```

3. **Live testing:**
   - https://cdn.ad.nexus/studio/
   - https://studio.ad.nexus

## Monitoring

### Health Check Endpoint
- URL: `https://studio.ad.nexus/index.html`
- Expected: 200 OK with HTML content

### Key Metrics
- Page load time: Target <2s
- CDN cache hit rate: Target >95%
- Error rate: Target <0.1%

## Rollback

To rollback to previous version:

```bash
# List S3 versions
aws s3api list-object-versions --bucket cdn.ad.nexus --prefix studio/

# Restore specific version
aws s3api copy-object \
  --copy-source cdn.ad.nexus/studio/index.html?versionId=VERSION_ID \
  --bucket cdn.ad.nexus \
  --key studio/index.html
```

## Troubleshooting

### CORS Errors
- Check CORS configuration on CDN
- Verify `coi-serviceworker.js` is served correctly

### SharedArrayBuffer Issues
- Ensure COOP/COEP headers are set:
  ```
  Cross-Origin-Opener-Policy: same-origin
  Cross-Origin-Embedder-Policy: require-corp
  ```

### Import Map Issues
- Verify `importmap.json` is accessible
- Check browser console for module resolution errors

### Video Codec Issues
- Studio requires WebCodecs API support
- Check browser compatibility: Chrome 94+, Safari 16.4+

## Support

- **Documentation**: https://github.com/adnexusinc/studio
- **Issues**: https://github.com/adnexusinc/studio/issues
- **Support**: support@ad.nexus