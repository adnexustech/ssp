#!/bin/bash
set -e

# Adnexus Studio CDN Deployment Script
# This script deploys the built Studio files to cdn.ad.nexus

echo "🚀 Adnexus Studio CDN Deployment"
echo "================================"

# Build the project
echo "📦 Building Studio..."
npm run build-production

# Check if x/ directory exists
if [ ! -d "x" ]; then
  echo "❌ Error: x/ directory not found. Build may have failed."
  exit 1
fi

echo "✅ Build complete!"
echo ""
echo "📊 Build output:"
ls -lh x/ | head -20

echo ""
echo "📋 Deployment Options:"
echo "1. AWS S3/CloudFront"
echo "2. Custom CDN server (rsync/scp)"
echo "3. Create tarball for manual deployment"
echo ""

read -p "Select deployment method (1/2/3): " choice

case $choice in
  1)
    echo "🌩️  AWS S3/CloudFront Deployment"
    read -p "S3 Bucket (e.g., cdn.ad.nexus): " bucket
    read -p "Path prefix (e.g., /studio/): " prefix

    # Deploy to S3 with cache headers
    echo "📤 Uploading to S3..."
    aws s3 sync x/ "s3://$bucket$prefix" \
      --delete \
      --cache-control "public,max-age=31536000,immutable" \
      --exclude "*.html" \
      --exclude "*.json"

    aws s3 sync x/ "s3://$bucket$prefix" \
      --exclude "*" \
      --include "*.html" \
      --include "*.json" \
      --cache-control "public,max-age=0,must-revalidate"

    echo "✅ Uploaded to S3!"

    read -p "CloudFront distribution ID (or skip): " cf_id
    if [ ! -z "$cf_id" ]; then
      echo "🔄 Invalidating CloudFront cache..."
      aws cloudfront create-invalidation \
        --distribution-id "$cf_id" \
        --paths "$prefix*"
      echo "✅ CloudFront invalidated!"
    fi
    ;;

  2)
    echo "🖥️  Custom CDN Server Deployment"
    read -p "Server address (user@host): " server
    read -p "Remote path: " remote_path

    echo "📤 Deploying via rsync..."
    rsync -avz --delete x/ "$server:$remote_path"
    echo "✅ Deployed to $server:$remote_path!"
    ;;

  3)
    echo "📦 Creating deployment tarball..."
    cd x
    tar -czf ../studio-dist-$(date +%Y%m%d-%H%M%S).tar.gz .
    cd ..
    echo "✅ Tarball created!"
    ls -lh studio-dist-*.tar.gz | tail -1
    echo ""
    echo "📋 Manual deployment steps:"
    echo "1. Extract tarball on CDN server"
    echo "2. Point cdn.ad.nexus/studio/ to extracted files"
    echo "3. Configure proper cache headers"
    echo "4. Ensure CORS is enabled if needed"
    ;;

  *)
    echo "❌ Invalid choice"
    exit 1
    ;;
esac

echo ""
echo "✅ Deployment complete!"
echo "🌐 Studio should be accessible at: https://cdn.ad.nexus/studio/"
echo "   or via: https://studio.ad.nexus (if DNS configured)"