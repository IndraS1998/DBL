# Distributed Balance Ledger (Raft Consensus)

![Raft Consensus](https://img.shields.io/badge/consensus-Raft-brightgreen) 
![Distributed System](https://img.shields.io/badge/architecture-distributed-blue)

A fault-tolerant distributed ledger system that maintains account balances using the Raft consensus algorithm.

## Key Features

- **Strong consistency** via Raft consensus protocol
- **Fault-tolerant** - continues operating with node failures (N/2+1 nodes available)
- **Transactional operations**:
  - `create_account(account_id, initial_balance)`
  - `get_balance(account_id)`
  - `transfer(from, to, amount)`
- **Durability** - persisted log and snapshots

## Architecture

```mermaid
graph TD
    Client -->|HTTP| Leader
    Leader -->|AppendEntryRPC| Follower1
    Leader -->|AppendEntryRPC| Follower2
    Follower1 -->|ResponseRPC| Leader
    Follower2 -->|ResponseRPC| Leader
