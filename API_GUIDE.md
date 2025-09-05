# Plakar RESTful API Guide

## Overview

The Plakar RESTful API provides programmatic access to all backup and restore operations. Built on top of Plakar's robust backup engine, the API enables integration with external systems, automation of backup workflows, and development of custom backup solutions.

## Key Features

- **Complete Backup Operations**: Create, list, browse, and restore backups
- **Secure Access**: Bearer token authentication with optional OAuth integration
- **Real-time Monitoring**: Repository statistics and snapshot information
- **File System Browsing**: Navigate backup contents with VFS operations
- **Search Capabilities**: Find files across snapshots and time ranges
- **Download Management**: Generate signed URLs and create download packages
- **Integration Support**: Plugin system for custom data sources and storage backends

## Quick Start

### 1. Start the API Server

```bash
# Start Plakar UI with API access
plakar at /path/to/repository ui --addr 0.0.0.0:8080

# Or start without authentication (development only)
plakar at /path/to/repository ui --addr 0.0.0.0:8080 --no-auth
```

### 2. Get API Information

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:8080/api/info
```

Response:
```json
{
  "repository_id": "550e8400-e29b-41d4-a716-446655440000",
  "authenticated": true,
  "version": "1.0.2",
  "browsable": true,
  "demo_mode": false
}
```

### 3. List Snapshots

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/repository/snapshots?limit=10&sort=-Timestamp"
```

## Authentication

### Bearer Token Authentication

Include the token in the Authorization header:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:8080/api/repository/info
```

### OAuth Integration

For GitHub OAuth:

```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -d '{"redirect": "http://localhost:3000/callback"}' \
     http://localhost:8080/api/authentication/login/github
```

Response:
```json
{
  "URL": "https://github.com/login/oauth/authorize?client_id=..."
}
```

## Core API Operations

### Repository Management

#### Create Repository

```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -d '{
       "location": "fs:/var/backups/my-repo",
       "passphrase": "secure-passphrase-here",
       "hashing": "SHA256"
     }' \
     http://localhost:8080/api/repository/create
```

Response:
```json
{
  "repository_id": "550e8400-e29b-41d4-a716-446655440000",
  "location": "fs:/var/backups/my-repo"
}
```

#### Get Repository Information

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:8080/api/repository/info
```

Response includes storage statistics, efficiency metrics, and configuration:

```json
{
  "item": {
    "location": "fs:/var/backups",
    "snapshots": {
      "total": 42,
      "storage_size": 1073741824,
      "logical_size": 5368709120,
      "efficiency": 80.0,
      "snapshots_per_day": [2, 1, 3, 0, 1, ...]
    },
    "configuration": {
      "repository_id": "550e8400-e29b-41d4-a716-446655440000",
      "version": "1.0.2",
      "timestamp": "2025-01-15T10:30:00Z"
    },
    "os": "linux",
    "arch": "amd64"
  }
}
```

#### List Snapshots with Filtering

```bash
# Get recent snapshots
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/repository/snapshots?since=2025-01-01T00:00:00Z&limit=20"

# Filter by importer type
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/repository/snapshots?importer=fs&sort=-Timestamp"
```

### Snapshot Operations

#### Create Backup Snapshot

```bash
curl -X POST \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "source": "/home/user/documents",
       "tags": ["documents", "personal"],
       "excludes": ["*.tmp", "*.log"],
       "concurrency": 4,
       "check": true
     }' \
     http://localhost:8080/api/snapshots/create
```

Response:
```json
{
  "snapshot_id": "abc123def456...",
  "timestamp": "2025-01-15T10:30:00Z",
  "summary": {
    "files": 1250,
    "directories": 45,
    "size": 2147483648
  }
}
```

#### Get Snapshot Details

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:8080/api/snapshot/abc123def456...
```

#### Browse Snapshot Filesystem

```bash
# Browse root directory
curl -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:8080/api/snapshot/vfs/abc123def456

# Browse specific directory
curl -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:8080/api/snapshot/vfs/abc123def456:/home/user/documents
```

#### List Directory Contents

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/snapshot/vfs/children/abc123def456:/home/user?limit=100&sort=Name"
```

#### Search Files

```bash
# Search for files matching pattern
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/snapshot/vfs/search/abc123def456:/home?pattern=*.pdf&limit=50"
```

#### Restore Snapshot

```bash
curl -X POST \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "destination": "/tmp/restore",
       "paths": ["/home/user/documents/important.txt"],
       "skip_permissions": false
     }' \
     http://localhost:8080/api/snapshots/abc123def456/restore
```

#### Check Snapshot Integrity

```bash
curl -X POST \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "fast": false,
       "no_verify": false,
       "concurrency": 4
     }' \
     http://localhost:8080/api/snapshots/abc123def456/check
```

#### Compare Snapshots

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/snapshots/abc123def456/diff/def456abc123?recursive=true&highlight=true"
```

#### Mount Snapshot as Filesystem

```bash
curl -X POST \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "mountpoint": "/mnt/backup"
     }' \
     http://localhost:8080/api/snapshots/abc123def456/mount
```

### File Operations

#### Read File Content

```bash
# Read file as text
curl -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:8080/api/snapshot/reader/abc123def456:/home/user/document.txt

# Download file
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/snapshot/reader/abc123def456:/home/user/file.zip?download=true" \
     -o downloaded_file.zip

# Get syntax-highlighted code
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/snapshot/reader/abc123def456:/home/user/script.py?render=code"
```

#### Generate Signed URLs

```bash
curl -X POST \
     -H "Authorization: Bearer YOUR_TOKEN" \
     http://localhost:8080/api/snapshot/reader-sign-url/abc123def456:/home/user/file.txt
```

Response:
```json
{
  "url": "http://localhost:8080/api/snapshot/reader/abc123def456:/home/user/file.txt?signature=...",
  "expires_at": "2025-01-15T11:30:00Z"
}
```

#### Get File with Processing

```bash
# Get file with syntax highlighting
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/files/cat/abc123def456:/home/user/script.py?highlight=true"

# Get compressed file with decompression
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/files/cat/abc123def456:/var/log/app.log.gz?decompress=true"
```

#### Get File Digest

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/files/digest/abc123def456:/home/user/document.pdf?algorithm=SHA256"
```

Response:
```json
{
  "path": "/home/user/document.pdf",
  "algorithm": "SHA256",
  "digest": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "size": 1048576
}
```

### Download Packages

#### Create Download Package

```bash
curl -X POST \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "files": ["/home/user/documents", "/home/user/photos"],
       "format": "zip",
       "rebase": true
     }' \
     http://localhost:8080/api/snapshot/vfs/downloader/abc123def456
```

Response:
```json
{
  "download_id": "dl_xyz789",
  "download_url": "http://localhost:8080/api/snapshot/vfs/downloader-sign-url/dl_xyz789"
}
```

### Search and Locate

#### Find File Across Snapshots

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/repository/locate-pathname?resource=/home/user/important.txt&limit=10&sort=-Timestamp"
```

This returns all versions of the file across different snapshots:

```json
{
  "total": 5,
  "items": [
    {
      "snapshot": {
        "id": "abc123def456...",
        "timestamp": "2025-01-15T10:00:00Z",
        "summary": {"files": 1000, "directories": 50, "size": 1048576}
      },
      "vfs_entry": {
        "name": "important.txt",
        "path": "/home/user/important.txt",
        "type": "file",
        "size": 2048,
        "mtime": "2025-01-15T09:30:00Z"
      }
    }
  ]
}
```

#### Locate Files by Pattern

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     "http://localhost:8080/api/search/locate?patterns=*.pdf&patterns=*.doc&limit=50"
```

### Repository Maintenance

#### Run Maintenance

```bash
curl -X POST \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "operations": ["cleanup", "rebuild_cache", "verify_integrity"]
     }' \
     http://localhost:8080/api/repository/maintenance
```

#### Prune Repository

```bash
curl -X POST \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "apply": true,
       "filters": {
         "before": "2024-01-01T00:00:00Z"
       }
     }' \
     http://localhost:8080/api/repository/prune
```

#### Synchronize Repositories

```bash
curl -X POST \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "target_repository": "s3://backup-bucket/remote-repo",
       "direction": "to",
       "filters": {
         "tags": ["important"]
       }
     }' \
     http://localhost:8080/api/repository/sync
```

#### Remove Snapshots

```bash
curl -X DELETE \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "snapshot_ids": ["abc123def456", "def456abc123"],
       "apply": true
     }' \
     http://localhost:8080/api/snapshots/remove
```

### Scheduler Operations

#### Start Scheduler

```bash
curl -X POST http://localhost:8080/api/scheduler/start
```

#### Get Scheduler Status

```bash
curl http://localhost:8080/api/scheduler/status
```

Response:
```json
{
  "running": true,
  "next_run": "2025-01-16T02:00:00Z",
  "active_jobs": 2
}
```

## Integration Examples

### Python Integration

```python
import requests
import json
from datetime import datetime, timedelta

class PlakarAPI:
    def __init__(self, base_url, token):
        self.base_url = base_url.rstrip('/')
        self.headers = {'Authorization': f'Bearer {token}'}
    
    def create_repository(self, location, passphrase, hashing='SHA256', plaintext=False):
        data = {
            'location': location,
            'passphrase': passphrase,
            'hashing': hashing,
            'plaintext': plaintext
        }
        response = requests.post(f'{self.base_url}/api/repository/create',
                               headers=self.headers, json=data)
        return response.json()
    
    def get_repository_info(self):
        response = requests.get(f'{self.base_url}/api/repository/info', 
                              headers=self.headers)
        return response.json()
    
    def create_backup(self, source, tags=None, excludes=None, concurrency=4, check=False):
        data = {
            'source': source,
            'tags': tags or [],
            'excludes': excludes or [],
            'concurrency': concurrency,
            'check': check
        }
        response = requests.post(f'{self.base_url}/api/snapshots/create',
                               headers=self.headers, json=data)
        return response.json()
    
    def list_snapshots(self, limit=50, since=None, importer=None):
        params = {'limit': limit}
        if since:
            params['since'] = since.isoformat()
        if importer:
            params['importer'] = importer
            
        response = requests.get(f'{self.base_url}/api/repository/snapshots',
                              headers=self.headers, params=params)
        return response.json()
    
    def restore_snapshot(self, snapshot_id, destination, paths=None, skip_permissions=False):
        data = {
            'destination': destination,
            'paths': paths or [],
            'skip_permissions': skip_permissions
        }
        response = requests.post(f'{self.base_url}/api/snapshots/{snapshot_id}/restore',
                               headers=self.headers, json=data)
        return response.json()
    
    def check_snapshot(self, snapshot_id, paths=None, fast=False, concurrency=4):
        data = {
            'paths': paths or [],
            'fast': fast,
            'concurrency': concurrency
        }
        response = requests.post(f'{self.base_url}/api/snapshots/{snapshot_id}/check',
                               headers=self.headers, json=data)
        return response.json()
    
    def browse_snapshot(self, snapshot_id, path=''):
        url = f'{self.base_url}/api/snapshot/vfs/{snapshot_id}'
        if path:
            url += f':{path}'
        response = requests.get(url, headers=self.headers)
        return response.json()
    
    def search_files(self, snapshot_id, pattern, path='', limit=100):
        url = f'{self.base_url}/api/snapshot/vfs/search/{snapshot_id}'
        if path:
            url += f':{path}'
        params = {'pattern': pattern, 'limit': limit}
        response = requests.get(url, headers=self.headers, params=params)
        return response.json()
    
    def locate_files(self, patterns, snapshot=None, limit=100):
        params = {'patterns': patterns, 'limit': limit}
        if snapshot:
            params['snapshot'] = snapshot
        response = requests.get(f'{self.base_url}/api/search/locate',
                              headers=self.headers, params=params)
        return response.json()
    
    def get_file_digest(self, snapshot_id, file_path, algorithm='SHA256'):
        url = f'{self.base_url}/api/files/digest/{snapshot_id}:{file_path}'
        params = {'algorithm': algorithm}
        response = requests.get(url, headers=self.headers, params=params)
        return response.json()
    
    def download_file(self, snapshot_id, file_path, local_path):
        url = f'{self.base_url}/api/snapshot/reader/{snapshot_id}:{file_path}'
        response = requests.get(url, headers=self.headers, stream=True)
        
        with open(local_path, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)
    
    def mount_snapshot(self, snapshot_id, mountpoint):
        data = {'mountpoint': mountpoint}
        response = requests.post(f'{self.base_url}/api/snapshots/{snapshot_id}/mount',
                               headers=self.headers, json=data)
        return response.json()
    
    def run_maintenance(self, operations=None):
        data = {'operations': operations or ['cleanup']}
        response = requests.post(f'{self.base_url}/api/repository/maintenance',
                               headers=self.headers, json=data)
        return response.json()
    
    def prune_repository(self, apply=False, filters=None):
        data = {'apply': apply, 'filters': filters or {}}
        response = requests.post(f'{self.base_url}/api/repository/prune',
                               headers=self.headers, json=data)
        return response.json()
    
    def sync_repository(self, target_repository, direction, snapshot_ids=None, filters=None):
        data = {
            'target_repository': target_repository,
            'direction': direction,
            'snapshot_ids': snapshot_ids or [],
            'filters': filters or {}
        }
        response = requests.post(f'{self.base_url}/api/repository/sync',
                               headers=self.headers, json=data)
        return response.json()

# Usage example
api = PlakarAPI('http://localhost:8080', 'your-token-here')

# Create a new repository
repo_info = api.create_repository(
    location='fs:/var/backups/my-repo',
    passphrase='secure-passphrase'
)
print(f"Created repository: {repo_info['repository_id']}")

# Create a backup
backup_result = api.create_backup(
    source='/home/user/documents',
    tags=['documents', 'personal'],
    excludes=['*.tmp', '*.log'],
    check=True
)
print(f"Backup created: {backup_result['snapshot_id']}")

# Get repository statistics
info = api.get_repository_info()
print(f"Repository has {info['item']['snapshots']['total']} snapshots")

# List recent snapshots
since = datetime.now() - timedelta(days=7)
snapshots = api.list_snapshots(limit=10, since=since)

for snapshot in snapshots['items']:
    print(f"Snapshot {snapshot['id'][:8]} from {snapshot['timestamp']}")
    
    # Check snapshot integrity
    check_result = api.check_snapshot(snapshot['id'])
    if check_result['status'] == 'success':
        print(f"  ✓ Integrity check passed")
    
    # Browse snapshot root
    contents = api.browse_snapshot(snapshot['id'])
    print(f"  Contains {len(contents.get('children', []))} items")

# Locate files across snapshots
matches = api.locate_files(['*.pdf', '*.doc'])
print(f"Found {len(matches['matches'])} document files across all snapshots")

# Run maintenance
maintenance_result = api.run_maintenance(['cleanup', 'verify_integrity'])
print(f"Maintenance completed, cleaned {maintenance_result['cleaned_objects']} objects")
```

### JavaScript/Node.js Integration

```javascript
const axios = require('axios');

class PlakarAPI {
    constructor(baseUrl, token) {
        this.baseUrl = baseUrl.replace(/\/$/, '');
        this.headers = { 'Authorization': `Bearer ${token}` };
    }

    async getRepositoryInfo() {
        const response = await axios.get(`${this.baseUrl}/api/repository/info`, {
            headers: this.headers
        });
        return response.data;
    }

    async listSnapshots(options = {}) {
        const params = new URLSearchParams();
        if (options.limit) params.append('limit', options.limit);
        if (options.since) params.append('since', options.since);
        if (options.importer) params.append('importer', options.importer);

        const response = await axios.get(`${this.baseUrl}/api/repository/snapshots?${params}`, {
            headers: this.headers
        });
        return response.data;
    }

    async searchFiles(snapshotId, pattern, path = '', limit = 100) {
        const url = `${this.baseUrl}/api/snapshot/vfs/search/${snapshotId}${path ? ':' + path : ''}`;
        const response = await axios.get(url, {
            headers: this.headers,
            params: { pattern, limit }
        });
        return response.data;
    }

    async createDownloadPackage(snapshotId, files, format = 'zip') {
        const url = `${this.baseUrl}/api/snapshot/vfs/downloader/${snapshotId}`;
        const response = await axios.post(url, {
            files,
            format,
            rebase: true
        }, {
            headers: { ...this.headers, 'Content-Type': 'application/json' }
        });
        return response.data;
    }
}

// Usage example
const api = new PlakarAPI('http://localhost:8080', 'your-token-here');

async function backupReport() {
    try {
        const info = await api.getRepositoryInfo();
        console.log(`Repository Efficiency: ${info.item.snapshots.efficiency}%`);
        
        const snapshots = await api.listSnapshots({ limit: 5 });
        console.log(`Latest ${snapshots.items.length} snapshots:`);
        
        for (const snapshot of snapshots.items) {
            console.log(`- ${snapshot.id.substring(0, 8)} (${snapshot.timestamp})`);
            
            // Search for log files in this snapshot
            const logs = await api.searchFiles(snapshot.id, '*.log', '/var/log');
            console.log(`  Found ${logs.items.length} log files`);
        }
    } catch (error) {
        console.error('API Error:', error.response?.data || error.message);
    }
}

backupReport();
```

### Bash/Shell Integration

```bash
#!/bin/bash

PLAKAR_API_URL="http://localhost:8080"
PLAKAR_TOKEN="your-token-here"

# Function to make API calls
api_call() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    
    if [ "$method" = "GET" ]; then
        curl -s -H "Authorization: Bearer $PLAKAR_TOKEN" \
             "$PLAKAR_API_URL/api$endpoint"
    elif [ "$method" = "POST" ]; then
        curl -s -X POST \
             -H "Authorization: Bearer $PLAKAR_TOKEN" \
             -H "Content-Type: application/json" \
             -d "$data" \
             "$PLAKAR_API_URL/api$endpoint"
    fi
}

# Get repository info
echo "Repository Information:"
api_call GET "/repository/info" | jq '.item.snapshots'

# List recent snapshots
echo -e "\nRecent Snapshots:"
api_call GET "/repository/snapshots?limit=5&sort=-Timestamp" | \
    jq -r '.items[] | "\(.id[0:8]) - \(.timestamp) - \(.summary.files) files"'

# Search for configuration files across all snapshots
echo -e "\nSearching for config files:"
LATEST_SNAPSHOT=$(api_call GET "/repository/snapshots?limit=1&sort=-Timestamp" | jq -r '.items[0].id')
api_call GET "/snapshot/vfs/search/$LATEST_SNAPSHOT?pattern=*.conf&limit=10" | \
    jq -r '.items[] | "  \(.path) (\(.size) bytes)"'

# Create a download package
echo -e "\nCreating download package:"
DOWNLOAD_DATA='{"files":["/etc/hosts","/etc/passwd"],"format":"zip","rebase":true}'
api_call POST "/snapshot/vfs/downloader/$LATEST_SNAPSHOT" "$DOWNLOAD_DATA" | \
    jq -r '"Download URL: \(.download_url)"'
```

## Error Handling

The API uses standard HTTP status codes and returns structured error responses:

```json
{
  "error": {
    "code": "not-found",
    "message": "Snapshot not found",
    "params": {
      "snapshot": {
        "code": "invalid_argument",
        "message": "Invalid snapshot ID format"
      }
    }
  }
}
```

Common error codes:
- `400` - Bad Request (invalid parameters)
- `401` - Unauthorized (missing or invalid token)
- `404` - Not Found (resource doesn't exist)
- `500` - Internal Server Error

## Rate Limiting and Best Practices

### Rate Limiting
- API requests may be rate-limited to ensure system stability
- Implement exponential backoff for retries
- Use pagination for large result sets

### Best Practices

1. **Use Pagination**: Always use `limit` and `offset` parameters for large datasets
2. **Cache Results**: Cache repository info and snapshot lists when possible
3. **Signed URLs**: Use signed URLs for file downloads to avoid token exposure
4. **Error Handling**: Implement proper error handling and retry logic
5. **Filtering**: Use query parameters to filter results and reduce bandwidth

### Performance Tips

```python
# Good: Use pagination and filtering
snapshots = api.list_snapshots(limit=50, since=last_week, importer='fs')

# Good: Search within specific paths
results = api.search_files(snapshot_id, '*.log', '/var/log', limit=100)

# Good: Use signed URLs for downloads
signed_url = api.create_signed_url(snapshot_id, file_path)
# Download using the signed URL without authentication headers

# Avoid: Loading all snapshots at once
# snapshots = api.list_snapshots(limit=10000)  # Don't do this
```

## Advanced Use Cases

### Automated Backup Monitoring

```python
import time
from datetime import datetime, timedelta

def monitor_backups(api):
    """Monitor backup health and send alerts"""
    info = api.get_repository_info()
    
    # Check if efficiency is dropping
    efficiency = info['item']['snapshots']['efficiency']
    if efficiency < 70:
        send_alert(f"Backup efficiency dropped to {efficiency}%")
    
    # Check for recent backups
    yesterday = datetime.now() - timedelta(days=1)
    recent = api.list_snapshots(since=yesterday, limit=1)
    
    if not recent['items']:
        send_alert("No backups in the last 24 hours")
    
    # Check storage growth
    storage_size = info['item']['snapshots']['storage_size']
    if storage_size > STORAGE_THRESHOLD:
        send_alert(f"Storage usage: {storage_size / 1e9:.1f} GB")
    
    # Run integrity checks on recent snapshots
    for snapshot in recent['items'][:5]:  # Check last 5 snapshots
        check_result = api.check_snapshot(snapshot['id'], fast=True)
        if check_result['status'] != 'success':
            send_alert(f"Integrity check failed for snapshot {snapshot['id'][:8]}")

def send_alert(message):
    print(f"ALERT: {message}")
    # Integrate with your alerting system

def automated_backup_and_maintenance(api):
    """Automated backup with maintenance"""
    # Create daily backup
    backup_result = api.create_backup(
        source='/important/data',
        tags=['daily', 'automated'],
        check=True
    )
    
    # Run maintenance weekly
    if datetime.now().weekday() == 6:  # Sunday
        maintenance_result = api.run_maintenance(['cleanup', 'verify_integrity'])
        print(f"Weekly maintenance: cleaned {maintenance_result['cleaned_objects']} objects")
    
    # Prune old backups monthly
    if datetime.now().day == 1:  # First day of month
        prune_result = api.prune_repository(
            apply=True,
            filters={'before': (datetime.now() - timedelta(days=90)).isoformat()}
        )
        print(f"Monthly prune: reclaimed {prune_result['reclaimed_space']} bytes")
```

### File Recovery Automation

```python
def recover_file_versions(api, file_path, days_back=30):
    """Recover all versions of a file from the last N days"""
    since = datetime.now() - timedelta(days=days_back)
    
    # Find all versions of the file
    locations = api.locate_pathname(
        resource=file_path,
        since=since.isoformat(),
        limit=100
    )
    
    recovery_dir = f"recovered_{file_path.replace('/', '_')}"
    os.makedirs(recovery_dir, exist_ok=True)
    
    for i, location in enumerate(locations['items']):
        snapshot_id = location['snapshot']['id']
        timestamp = location['snapshot']['timestamp']
        
        # Download each version
        filename = f"{i:03d}_{timestamp}_{os.path.basename(file_path)}"
        local_path = os.path.join(recovery_dir, filename)
        
        api.download_file(snapshot_id, file_path, local_path)
        print(f"Recovered: {filename}")

def disaster_recovery_workflow(api, critical_paths):
    """Complete disaster recovery workflow"""
    # Find latest snapshot
    snapshots = api.list_snapshots(limit=1, sort='-Timestamp')
    if not snapshots['items']:
        raise Exception("No snapshots found for recovery")
    
    latest_snapshot = snapshots['items'][0]
    print(f"Using snapshot {latest_snapshot['id'][:8]} from {latest_snapshot['timestamp']}")
    
    # Verify snapshot integrity
    check_result = api.check_snapshot(latest_snapshot['id'])
    if check_result['status'] != 'success':
        raise Exception("Snapshot integrity check failed")
    
    # Restore critical paths
    for path in critical_paths:
        print(f"Restoring {path}...")
        restore_result = api.restore_snapshot(
            latest_snapshot['id'],
            destination=f"/recovery{path}",
            paths=[path]
        )
        print(f"Restored {restore_result['restored_files']} files")
    
    # Generate recovery report
    return {
        'snapshot_id': latest_snapshot['id'],
        'timestamp': latest_snapshot['timestamp'],
        'restored_paths': critical_paths,
        'total_files': sum(restore_result['restored_files'] for path in critical_paths)
    }
```

## Security Considerations

1. **Token Management**: Store API tokens securely and rotate them regularly
2. **HTTPS**: Always use HTTPS in production environments
3. **Access Control**: Implement proper access controls and audit logging
4. **Input Validation**: Validate all input parameters to prevent injection attacks
5. **Rate Limiting**: Implement rate limiting to prevent abuse

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   ```bash
   # Check if token is valid
   curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/info
   ```

2. **Snapshot Not Found**
   ```bash
   # List available snapshots
   curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/repository/snapshots
   ```

3. **Path Resolution Issues**
   ```bash
   # Browse snapshot root to understand structure
   curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/snapshot/vfs/SNAPSHOT_ID
   ```

### Debug Mode

Enable debug logging in your API client:

```python
import logging
logging.basicConfig(level=logging.DEBUG)

# This will show all HTTP requests and responses
```

## API Versioning

The current API version is `1.0.2`. Future versions will maintain backward compatibility where possible. Version information is available in the `/api/info` endpoint.

## Support and Resources

- **Documentation**: [https://docs.plakar.io](https://docs.plakar.io)
- **GitHub**: [https://github.com/PlakarKorp/plakar](https://github.com/PlakarKorp/plakar)
- **Discord**: [https://discord.gg/uqdP9Wfzx3](https://discord.gg/uqdP9Wfzx3)
- **Website**: [https://www.plakar.io](https://www.plakar.io)

For API-specific questions, please use the GitHub issues or Discord community channels.