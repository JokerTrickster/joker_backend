#!/bin/bash

# Docker cleanup script
# Usage: ./scripts/cleanup.sh

set -e

echo "🧹 Starting Docker cleanup..."

# Stop all containers
echo "⏹️  Stopping all containers..."
docker stop $(docker ps -aq) 2>/dev/null || echo "No containers to stop"

# Remove all containers
echo "🗑️  Removing all containers..."
docker rm $(docker ps -aq) 2>/dev/null || echo "No containers to remove"

# Remove all images
echo "🖼️  Removing all images..."
docker rmi $(docker images -q) -f 2>/dev/null || echo "No images to remove"

# Remove all volumes
echo "📦 Removing all volumes..."
docker volume rm $(docker volume ls -q) 2>/dev/null || echo "No volumes to remove"

# Remove all networks (except default ones)
echo "🌐 Removing custom networks..."
docker network prune -f

# Clean up system
echo "🧽 Running system prune..."
docker system prune -af --volumes

# Show disk space
echo ""
echo "💾 Disk space after cleanup:"
df -h | grep -E '(Filesystem|/$|/home)'

echo ""
echo "✅ Docker cleanup completed!"
