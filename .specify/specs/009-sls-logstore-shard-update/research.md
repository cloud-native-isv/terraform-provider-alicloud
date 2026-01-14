# Research Findings

## Unknown: Shard IDs and Active State

**Decision**:
Use `GetHistograms` or `ListShards` equivalent to identify shards. Active shards typically have `Status="readwrite"`.

**Rationale**:
Only `readwrite` shards can be split or merged. `readonly` shards are historical.

## Unknown: Split/Merge API Signatures

**Decision**:
Checked `cws-lib-go`:
- `SplitShard(project, logstore, shardId, splitKey)`
- `MergeShards(project, logstore, shardId)`

**Rationale**:
Standard API. The "key" for splitting is usually the midpoint of the shard's inclusiveBeginKey and exclusiveEndKey.

## Unknown: Concurrency

**Decision**:
Sequential execution.

**Rationale**:
Splitting Shard A creates Shard B and C. If we need to split again, we might need to target B or C. We cannot target them until they exist and are in `readwrite` state. Therefore, strict sequential operations with state polling in between steps is required.

