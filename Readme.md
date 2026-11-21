# The Morteza Chain

> A custom build blockchain in golang.

## **1. What we build **

| Component                 | Purpose                                                                       |
| ------------------------- | ----------------------------------------------------------------------------- |
| Decentralized P2P network | Nodes communicate without relying on a central server.                        |
| Proof of Work             | Miners compete to produce blocks that satisfy the network's difficulty rules. |
| Synchronization           | Peers discover one another and synchronize blockchain history.                |
| Cryptographic hashing     | Hashes help detect changes to blocks and link blockchain history together.    |
| Transactions              | Transactions describe state changes, such as token transfers.                 |
| Digital signatures        | Authenticate transactions using private keys, with public-key verification.   |
| Keystore                  | Protect private keys at rest using password-based encryption.                 |
| CLI and HTTP API          | Provide interfaces for interacting with nodes.                                |
| Persistent storage        | Store blockchain data and recover state across restarts.                      |

## Transactions:

- A transaction is a record that can change the blockchain state. meaning: we can replay block.transactions and get the current state of our chain, eg: Balances and so on.

```json
//example
{
   From: Account,
   To: Account,
   Nonce: 1,
   Value: x,
   Data: "",
   Time: time,
}
```

## Blocks:

## Consensus and Proof of work:

```text
proof of work:
- make a candidate block
- keep blockHeader the same
- change nonce
- hash the block
- check whether hash passes the mining difficulty rule
- repeat until valid Hash is found
```
