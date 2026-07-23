# lsm-storage-engine

> **Work in progress.** Partially implemented — see status below.

A from-scratch LSM (Log-Structured Merge-Tree) storage engine written in Go, built for learning purposes by studying [goleveldb](https://github.com/syndtr/goleveldb).

## API

```go
db.Open(session)
db.Get(key []byte) ([]byte, error)
db.Has(key []byte) (bool, error)
db.Put(key, value []byte) error
db.Delete(key []byte) error
db.NewScanner(start, limit []byte) Scanner
db.Close() error
```

## Components

| Component | Role |
|---|---|
| **WAL** | Append-only write-ahead log for crash recovery |
| **Memtable** | In-memory sorted buffer; holds multiple versions per key via sequence number |
| **Storage** | Manages SSTables on disk |
| **Version** | In-memory representation of the LSM tree (which SSTable files exist at each level); rebuilt at startup, updated after each flush or compaction |
| **Compactor** | Merges SSTables across levels, drops stale versions and tombstones |
| **Scanner** | Merges results from memtable and SSTables; must be released after use |

## Write path

```
Put/Delete → WAL → Memtable → (flush) → L0 SSTable → (compaction) → L1+ SSTables
```

## Read path

```
Get → Memtable → Version (L0 → L1 → L2 ...)
```

## Key invariants

- Sequence number increases monotonically on every mutation
- WAL always contains at least what is in the memtable
- Memtable records are always sorted by key, then sequence number
- SSTables are flushed in sorted order and never exceed a max size
- L0 SSTables may overlap in key range; L1+ never overlap
- L1+ levels are kept sorted by key range in the version manifest
- Compaction retains only the latest version of each key; tombstones are dropped only when no lower level holds an older version

## Failure recovery

| Scenario | Recovery |
|---|---|
| Crash before WAL write completes | Truncate corrupted entry, recover to last valid WAL entry |
| Crash after WAL but before memtable insert | Replay WAL into memtable |
| Crash during flush | Recover WAL → memtable → re-flush |

## Status

| Component | Status |
|---|---|
| Skiplist | In progress |
| Memtable | In progress |
| WAL | Stub |
| Storage | Stub |
| Version | Stub |
| Compactor | Stub |
| DB | Not started |

## Design notes

See [adr/planning.md](adr/planning.md) for the full architectural design record.

Components are built incrementally to satisfy invariants rather than built to completion in isolation.
