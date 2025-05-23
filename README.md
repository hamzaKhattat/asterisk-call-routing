# S2 - Asterisk Call Routing System

A high-performance call routing system that manages dynamic DID allocation and call flow between multiple servers (S1 → S2 → S3 → S4).

## Features

- **Dynamic DID Management**: Automatic allocation and release of DIDs
- **Real-time Call Routing**: Intelligent routing with ANI/DNIS transformation
- **High Availability**: Built-in error recovery and stuck DID cleanup
- **Comprehensive Monitoring**: Web dashboard and API endpoints
- **Database-backed**: MySQL storage for reliability and scalability
- **Asterisk Integration**: ARI/AMI support for production environments

## Architecture
S1 (Origin) → S2 (Router) → S3 (Intermediate) → S4 (Final)
↑
MySQL DB
### Call Flow

1. S1 sends call with ANI-1 and DNIS-1
2. S2 allocates available DID from pool
3. S2 transforms: ANI-2 = DNIS-1, forwards with DID to S3
4. Call traverses through S3/Sx servers
5. Return call comes back with ANI-2 and DID
6. S2 restores original ANI-1 and DNIS-1
7. S2 forwards to S4 with original parameters
8. DID is released back to pool

## Quick Start

### Prerequisites

- Go 1.19+
- MySQL 5.7+
- Asterisk 18+ (optional, for production)

### Installation

```bash
# Clone and build
cd asterisk-call-routing
make build

# Setup database
make migrate

# Import DIDs
./bin/router -import-dids testdata/sample_dids.csv

# Start router
./bin/router -config configs/config.json
