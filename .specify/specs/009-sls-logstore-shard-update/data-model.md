# Data Model

## LogStore Shards

### Entity: Shard
From `cws-lib-go` / API response:

- `ShardID`: int (Unique identifier)
- `Status`: string (e.g., "readwrite", "readonly")
- `InclusiveBeginKey`: string (MD5 hex)
- `ExclusiveEndKey`: string (MD5 hex)
- `CreateTime`: int

### State Definitions

- **Active Shard**: `Status == "readwrite"`
- **Historical Shard**: `Status == "readonly"`

### Helper Functions

- `FilterActiveShards(shards []Shard) []Shard`: Returns only readwrite shards.
- `CalculateMidPoint(begin, end string) string`: Key generation for splitting.
