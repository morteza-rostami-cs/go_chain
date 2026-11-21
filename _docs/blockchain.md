<!--

# building a block chain in golang:

# setup a genesis file:

   the idea is:
      - blockchain is a database. and it starts with a genesis file.
      - this is the first page of the ledger.

   # declares things like:
      - when the chain was born: genesis_time
      - a unique id for the chain: chain_id
      - the initial token balances: balances

# chain_id
   - prevents a transaction signed for our chain, to being replayed on a different chain.

# balances: {
   morteza: 1,000,000 TMC token
}

1 TMC = 1$ or euoro

#======================

# blockchain is a database.
# tokens are units recorded in that database.
# users own tokens.
# user's balances are tracked.
# genesis is the initial state of the chain. initial start, config and token distribution.

#======================

# transactions and state changes:

# store latest chain state

we creat state.json -- aggregated database state containing the latest users.balances

# a customer called: BabaYaga buys E 1000 tokens.

1 TMC = E 1

E 1000 = 1000 TMC
100% bonus = another 1000 TMC

babaYaga has 2000 TMC tokens now.

# morteza -- gets +100 TMC -- daily reward.

#====================

# the take away is:

on a chain:

morteza = 1000
then babayaga buys -2
morteza has 998
then morteza gets reward: +100
morteza = 1098

## so: money can be created on some chain, based on initial rules.

## but: normally token is transfered from A to B and so on.
## so: we just keep a ledger of transactions.
eg: A send +1 to B

# so: state.json => is basically current snapshot of our chain state -- so we do not have to run all transactions, every time we want to see what is the balance for morteza.

#======================

andrej-3;     // 3 shots of vodka
andrej+3;     // technically purchasing from his own bar (money comes back)
andrej+700;   // reward for a week of work (7 × 100 tokens/day)

# so every transaction, we update state.json

#======================

Transactions (TX)

#======================

# we need these things in our chain:

signatures:
   - babayaga, only her can authorize the exact amount to send. so no one else can take money from her account.

blokcs + hashing:
   - so history of our tx.db chain , can not be rewritten. cause: if you change one block, feature blocks get invalid, and nodes find out.

decentralized p2p: (nodes)
   - so no single morteza can own the tx.db file.

#======================

# tokens are fungible:

   interchangeable, 1 btc vs 1 btc -- has the same value.
   vs NFT or consert tickets - price for each sits might be different.

#======================

 centralized systems put a few people in a position of power, and users don't have a choice.

#======================

# 8 problems:

# every transaction has to go through morteza, and that takes time.
   solution:: decentralized nodes, any node can process transactions at the same time.

# customer can not borrow/lend tokens to each other without oging throu morteza.
   solution: peer-to-peer transfers with no middleman.

# one disk can fail: lose client data
   solution: distributed replication -- every node keep a copy of the chain state or transactions

# database can be silently edited by morteza
   solution:: immutability via hashing + linked lists -- once agreed and transaction is recorded -- pase is frozen

# tax audits are hard, cuase tx history not trustworthy
   solution:: append only chain

# shareholders can sneak in new fees:
   solution:: transparent governance -- everyone sees rule changes

# users have no say in how fee change
   solution:: decentralized governance / voting

# users can not leave with their data id they disagree
   solution:: they hold a full copy

#======================

Hashing — take any data, produce a fixed-size "fingerprint." Change one character, the fingerprint changes completely. This is how you detect tampering.

Linked lists — each block points to the previous one. If you change block N, block N+1's pointer no longer matches → tampering cascades visibly. (We'll get to blocks soon.)

Immutable data structure — once data is written, it can never be modified. Only appended.

Distributed replication — the same data on many machines. One person cannot rewrite "the" database; only their copy.

Asymmetric cryptography — public/private key pairs. You sign with a private key; anyone can verify with the public key. This is what prevents Andrej from forging a transfer in BabaYaga's name.

#======================

fingerprint, snapshot, hash

block => block => block

each block keeps a hash to it's previous block.
- the hash if created by all the bytes in a single block/transaction => if you change one block => hash does not match and whole chain breaks.

hash ==> input => fixed sized bytes

SHA-256, which always outputs 32 bytes (256 bits = 64 hex characters)

# hash must be:
   - deterministic
      same input produces same output
   - hash should be fast to compute
   - preimage- resistant (one-way)
   Given a hash value, it's infeasible to figure out the input — except by trying every possible input (a brute-force search).

   so, it is not possible to go back from hash => to the original bytes. we can only verify a guess against the hash.

   - avalanche effect:
      change one byte and the whole hash changes into something else.
   - collision resistant:
      very hard to find 2 inputs , that produce the same hash.

#======================

# after each transaction, and anytime we call NewState... , we also hash the entire tx.db ??!

   not scalable!

# =========
we need a new data model:
   hash, block, blockHeader, blockFS

   batch processing
   hashing
   and a linked list

# write now, we snapshot/hash per transaction, o(n)
# if we implement a blockchain => 0(1)

# new ideas:
   - parent hash
   - blockFS wrapper

# things that pushes us towards a blockchain:
   # right now we rehash the whole tx.db file on each transaction
   1. network cost
      if tx.db is 10gb

      and nodes need the hash to agree on chain state.
      then each node needs a 10gb file to calc the hash locally.
      and when we make a tx, everyone needs to get the new 10gb file and calc the hash.

      so we end up sending this 10gb file all over the place.

   2. cpu cost:
      o(n) means: for running hash() on 10gb file, it is 10 times more than 1GB file operations.

#======================

# we solve the problem of hashing the entire file on each transaction, by building a blockchain::
   we do that using these ideas: Batch + Hash + Linked List ⇒ Blocks

# Batch::
   instead of persisting one TX at a time -- group many TXs into one unit/batch. (batch processing)

   so: for eg: we write 100 txs to our tx.db -- and then we do ont snapshot.

# Hash::
   hash the batch/block -- not the entire DB/chain. each batch/block gets it's own (fingerprint/hash)

# Linked list::
   new batch holds a reference/hash of the previous batch. so we create a chain/linked list.

## so:: we get a linear file of blocks. a chain! and transactions are batched into individual blocks.

#======================

# with hashing the entire tx.db -- we have to re hash the whole thing to -- confirm that chain is unchanged.

# but with blocks -- we only need to hash the last block, to say chain is intact. ??????

## because:: last block.hash -- is effected by the hash and basically content of all the previous blocks , so if you changes something in the block[0] and recompute the whole chain, the hash for last block also changes. parent hash is part of the content of each block.

## and the tempered version of the chain, gets rejected by other nodes , through some kinda of algorithms.

## 1 edit at block[0] cascade to every block after it.

## also:: we only need to re hash the new tx and one tx behind it, not the entire chain, on each tx.
## the last block, already holds the hash of it's parent, and that way is conntected to all to previous blocks.

## H_last => is a cryptographic commitment to the entire sequence.

#======================

💡 Extra notes
"Block" is really just a "batch" with a parent pointer. Everything else — "blockchain", "chain of blocks", "distributed ledger" — is decorated with the mystique of Bitcoin. Strip the branding and it's:

an append-only log of batches, where each batch contains the hash of the previous batch.

That's it. That's what we've built. The rest of the book is adding the distribution and agreement layers on top.

Why Time is uint64. Unix timestamps are non-negative integers (seconds since 1970-01-01 UTC). uint64 gives us plenty of range (up to ~584 billion years), and it's the type Bitcoin uses. Fits neatly into the BlockHeader without any JSON encoding weirdness.

Why does genesis have parent: 00000...? Because there's no previous block. Using a zero-value Hash{} is the standard convention. You can always recognize the genesis block by its all-zero parent hash.

Hashing JSON vs hashing binary. Some production blockchains hash a compact binary encoding (e.g., Ethereum's RLP) for efficiency. Our book uses JSON for readability. Both work; JSON is friendlier for a teaching repo, and it's fine for a toy chain. Just remember: whatever encoding you pick, it must be identical on every node, or hashes diverge.

What BlockFS's Key is not. It's not new information — it's a cache of Block.Hash(). If you edit block.db and change only the Key (but not the block itself), a reader would think "this block's hash is X" and store it, but when they later re-hash the block they'd get Y ≠ X. The book will introduce validation in a later chapter (Chapter 9-ish) that catches this. For now, trust the file.

#======================

# peer-to-peer :: p2p layer

every node (participant) gets a full copy of block.db -- keep it in sync with every other node copy.

changes move between nodes,
   - automatically
      the node syncs on its own
   - bidirectionally
      A => B and B => A
   - replicated:
      every node is a full copy, not a subset of the chain

#======================

# database should be shared between all nodes.

   so each node, needs to auto serve, it's block.db blocks => to the rest of the network.

   so: when morteza's node is down, other nodes do the transactions.

# inbound sync:
   babayaga's node => subscribe to other nodes and pull any blocks it's missing, replay locally and update it's state.

# master database: (mining)
   each node should be able to register tx and produce blocks. then broadcast blocks to other nodes.
   - for now: any node can add txs and write blocks.

# bidirectional sync:
   so: nodes send each other:
      - latest block height
      - request missing blocks
      - accept the missing blocks others have , that they do not.

#======================

# peer-to-peer => p2p protocol

   - serve blocks auto to peers
   - pull blocks from peers
   - produce blocks indpendently
   - bidirectional sync
   - join the network and discover peers

#======================

# bootstrap node:
   is just a server, to connect nodes together.

   now, we can have multiple nodes, acting as bootstrap nodes! and there is no one single entry point, you connect to any of them and sync with the network.

#======================

# idea of block number:
   each node has a Height

   node1_height > node2_height:
      node1 is ahead of node 2

   i have blocks 1 to 10
   you have 1 to 15
   give me the 5 that i do not have.

# if we add a block_number to blockHeader, we change the hash of that block.

#======================

# peer 2 peer (network)

Therefore, every P2P protocol hardcodes or configures a few well-known, reliable addresses called Bootstrap Nodes (or DNS seeds in Bitcoin).

A new node needs a "seed" node to connect to first.
This seed node is called the Bootstrap Node.


#======================

# fork problem:

node 1
G => A => B

node 2
G => C => D

here our common last block is G (genesis)

but: there is a fork , from G

both A and C are child of G

# so there are couple of things that we need to figuare out:

   - we need to send the latest_common_block to a peer
   - we need to solve the 2 competing blocks problem, which one comes after G , (A or C)
   - also we need to maintain the cryptographic relation between the blocks on chain.

#======================

# proof of work::

the idea is:
   - node doe not decide to append blocks, when ever they want
   - they compete to earn the right to create the next block.

# consensus:

   a set of rules that determine:
      - let distributed nodes to agree on the next valid block
      - and accepted chain history.

# there is a protocol that nodes follow, for deciding who adds the next block, not just every body adding blocks left and right!!

## for things that we dealing with, in terms of p2p sync:
   - p2p syncing rule
      how nodes exchange blockchain data
   - validation rules:
      which transactions and blocks are acceptable
   - peer trust
      how nodes assess info recieved from other peers
   - block creation
      who gets to create the next block

## so, who gets to create the next block and also which chain is the valid chain, between all the versions that nodes have.

## what is mining?

mining = competing to find a valid proof of work block.

a miner:: is a node that:
   - receives transactions
   - groups transactions into a candidate block
   - searches for a block hash that satisfies the network's difficulty requirement
   - announces the valid block to other nodes

## in bitcoin:
   - the first miner to find a valid solution -- can propose the next block and, in bitcoin, she receives a block reward.

#======================

# miners need to compete!!

# this is where the cryptographic puzzle comes in!!

# the idea of mining puzzle is something like this:

target: find a hash value, smaller that target or basically the has h starts with (some number of zeros) 00 sdfsfdfsdfsdfsd

the more zeros it has => it is harder to come up witn!

now, proof of work => is just keep trying and bruteforcing, your way to the answer::

something like this:

- hash the block header
- change the nonce, each time, 0, 1, 2, ...
   and rehash the nonce and header again
- until:: 00e4c1.... (you won!)

## the miner who gets a winner => brodcasts the block to the network.

   - to get an input that produces a certain hash, like the one starts with 00... => is very defficult and you have to brutte force it.
   - but: each node can use the nonce + block header to hash the block again
   - every nodes verifies the block_hash.
   - if valid: the miner earns the block reward + fees
   - block gets attached to the previous block.

Miner (finding):    ~10^22 guesses   → hours of electricity
Everyone else (verifying): 1 hash    → microseconds

so:

blockHeader:
   nonce: 0
   parent:
   ...

miner increments nonce and hashes this header.
untill hash < target (00...)
then braodcast it to other nodes.

other nodes the already have bunch of rules:
   one of them is has < target

   and if all checks out, they add the block to their chain.

#======================

# even with proof of wrok:: fork stil can happen?

if two miners solve the puzzle at the same time => and broadcast the two nodes with the same parent into the network?

#======================

BlockHash = Hash(
   BlockHeader,
   transaction,
   Nonce,
)

Hash(same input) = same output

## the reason we need nonce?

   block content is the same,
   but we want to get different hashes => to find: hash < 00asd
   so, we put nonce in blockHeader and increment it, starting from 0, 1, 2, 3, ...
   so the block hash changes each time.

#======================

## again => nodes do not call each other to confirm the hash!! the already have the rules, they just hash the block and compare it to the rules.

# some of the rules:
   - block transactions are valid
   - the parent hash connects to their accepted chain
   - the block hash satisfies difficulty, 0000 or 00000000 , more zero = harder to mine !!
   - the proof of work is valid

#======================

# Bitcoin solves, occasional fork => with using the chain with greatest accumulated proof of work.

#======================

# understand the mining loop:

- receive a pending block
- gnerate a nonce (increment)
- construct a block using that nonce
- calculate the block hash
- check whether the hash is valid
- if invalid, repeat
- if valid, return the mined block.

## so, the parent_hash and transactions, does not change.
## each incremented nonce, create a new candidate block, with a new hash, => until we hit 6 zeros, 000_000

#======================

# transaction propagation:
   every miner node, recives pending transactions from other nodes, tries mining and including them in chain.

# coordinated mining:
   when one miner succeed to mine a pending block, other should stop mining the same block and include this one in their chain.

# every node => should share their pending transactions. so everyone can mine!

# who every finishes the mining first, adds that block!! and other nodes should stop processing the same block.

#======================

# each node makes pool of waiting for mine transactions.
# each broadcases the pool of txs, to other miners.

# on sync => nodes share the pendingTxs list, so all miners have access to the same pool of transactions, and one at a time gets mined and added to the entire network block.db  !!
# so: any node can mine the same tx.

/tx/add => should only accept a tx. and add it to the pool !! do not create a block!

# if one node wins => others should stop processing the same block.

# pendingTx => should not effect the state.Balances

# we only apply tx to state.Balances => after mining and inclding inside the block.db
# we archive the tx after is included inside a mined block. so the same transaction => can not be mined twice.

# miner can pack the pendingTxs into one block and mine the whole thing.

# also, transactions included in a block, should be archived.

#======================

consensus failure:

   - forking
   - halting

# bitcoin solves fork by:
   chain selection favors => the valid chain with the most accumulated work.
   it means more mined blocks, after the fork.

# something like xrp => uses halting:
   stop producing new blocks, until validators can reach an agreement.


The trade-off:
- Forking: availability can continue, but different histories may temporarily exist.
- Halting: consistency is prioritized, but the system sacrifices availability.
This is an important distributed systems design decision.

#======================

Safety + Liveness + Fault Tolerance

safety:
   nothing incorrect happens

liveness:
   the system continues to make progress

fault tolerance
   the system can handle failures

#======================

small eth => wei

1 ETH = 1,000,000,000,000,000,000 wei

The blockchain protocol creates new cryptocurrency and credits it to the miner who successfully produces a block.

#======================

Only the miner who creates a new block receives the reward.

# block reward:
   mine a block, protocol gives you reward in your balance.

#======================

# hard fork?

fork:
   - it could be the chain block history is different
   - if coulf be also that rules or block data structure has changed

protocol incompatibility vs chain split

A hard fork occurs when the rule change causes incompatibility between participants, such that nodes running different rules can no longer agree on the same valid blockchain.

#======================

A blockchain is a shared database governed by rules that every participant must follow. When those rules change, old blocks may no longer be valid under the new software. Since nodes are independently operated, upgrades require coordination. Migration preserves the old transaction history by replaying transactions into a new blockchain and allowing the new consensus rules to produce valid blocks.

#======================

# signature

# private key
   only owner should have this, we sign transactions with it
# public key
   - shared with everyone
   - used to verify signature

# any blockchain node, can verify sig => using public key.

#======================

# asymmetric cryptography: (two keys instead of one password)

- we use a mathematically related pair of keys::
   - public key
   - private key

# we can sign a tx with a private key

# then: any node can verify sig by usig the public key.

# we can not go from public key to => private key

# private key stays on local machine.

# and node only receives::
   - signature
   - info that it needs to verify it.

# the main idea that prevents other dudes from spending my money is, this:

   address :: value
   balances: {
      "0001234": 30 tmc,
      "0001456": 0 tmc,
   }

   the address that is mapped to value, will be derived from sinature.
   and signature, give public address.
   adn signature, is produced using, private key.

   so if you use a different private key, and you don't have mine => signature produces a different address.

   so: you get to spend money from a different account and not mine.

#======================

# process:

wallet generate a random => private_key

then: derives a => public key from it

wallet signs tx with a private key => get a signature

node, verify sig using => public key

and: address is derived from signature verification.

#======================

# basically, signature gives us two things:
   - who signed sig, and the account address to change it's balance
   - if tx data changes, sig is not valid anymore, cause: we hash the whole thing with a private key to get a sig

#======================

Now the book introduces the specific cryptographic algorithm used by Ethereum and Bitcoin: ECDSA.

ECDSA = Elliptic Curve Digital Signature Algorithm

#======================

How these pieces fit together
Keystore
Generate, load, store, encrypt keys

Wallet
Manage and protect the user's keys

Crypto operations
Sign and verify using cryptographic algorithms

#======================

private_key => public_key => account_address

account_address => shows up on chain

state.Balances = {
   "0x2323dfs": 10,
}

#======================

Ethereum derives an account address from the public key using Keccak-256.

# process of getting an ETH address => from a public key

Public Key
    ↓
Remove the uncompressed public-key prefix
    ↓
Keccak-256 hash
    ↓
Take the last 20 bytes
    ↓
Ethereum Address

#======================

the address is 20 bytes, not 32 bytes. The Keccak-256 hash is 32 bytes, but Ethereum takes only its last 20 bytes.

0x22ba1F80452E6220c7cc6ea2D1e3EEDDaC5F694A

## about the above address::

   0x indicates hexadecimal representation.

   The remaining 40 hexadecimal characters represent 20 bytes.

   The mixed uppercase/lowercase letters provide an address checksum format (EIP-55).

#======================

# file generated by, new-account , keystore::

{
   # account_address (not encrypted)
   "address":"eb698b87894d918666f05465bdca59795ef9b5bb",

   # info, about how private_key is encrypted!!
   "crypto":{
      "cipher":"aes-128-ctr",
      # encrypted private_key
      "ciphertext":"e8f1871c42b061ef7619b7393ecbfdb447b132fb21a8144aca1e9715a7314777",
      "cipherparams":{
         "iv":"dcb235ca70775f06a43aafbfcab2d8a8"
      },
      "kdf":"scrypt",
      "kdfparams":{
         "dklen":32,
         "n":262144,
         "p":1,
         "r":8,"salt":"d2ef777f69ed7098eaf40e648cecef3acc74f4d2a737b3bc2e58dbcb4924deca"
      },"mac":"fe4a51057655e16f4ca205f7505e65a18d645ede7e36417fbd4d68f9fbb6970c"
   },
   "id":"97b7f1da-1d8e-44cb-b454-40aabee5d304",
   "version":3
}


Password → Scrypt → Encryption Key → AES-128-CTR → Encrypted Private Key

There are two different cryptographic operations here:
- Scrypt derives an encryption key from the user's password. It is intentionally expensive to make password guessing harder.
- AES-128-CTR uses that derived key to encrypt the private key.

# basically:: using passphrase we also decrypt our private_key!! which is encrypted inside keystore file!!


#======================

type PublicKey struct {
    elliptic.Curve
    X, Y *big.Int
}

type PrivateKey struct {
    PublicKey
    D *big.Int
}

type ecdsaSignature struct {
    R, S *big.Int
}

#======================

# digital signature:

# proof of identity or ownership

some how the wallet/node/ person , who is sending the transaction, it should also produce some proof that every node can verify.

# called message integrity!
# digital sig => is tied to content of tx and private key!! the whole thing is hashed together! , so change tx content and signature does not match anymore!

#======================

# this is the signature

signature = Sign(Keccak256(message), privateKey)

# signature output:

R

S

V


R and S are the two main ECDSA signature values.

V is the recovery identifier used by Ethereum's signature-recovery mechanism.

# sig is gonna be a 65bytes slice

#======================

Alice has private key a and public key A.

She signs message "hello" → signature s.

Bob has private key b and public key B. He cannot produce a signature that recovers to A — only Alice can.

Anyone who sees (hello, s) runs Verify, gets A back, and knows Alice signed it.

#======================

# account
   nonce

we increment account.nonce , per tx
so: nonce can stop a replay attack!!

# so, you can not send the same transaction with nonce = 1 , twice!

# we count txs on, state.account.nonce

# also: put a nonce on tx.nonce
   so we can replay txs for an account and increment accounts[0x1234] = 2

#======================

denial-of-service problem. Anyone can send a million tiny TXs, and every node has to mine them, store them, sync them, apply them — for free. The network grinds to a halt.

 In Ethereum, fees are the only block reward (Ethereum abolished its base block reward in 2022). In Bitcoin, both exist (base reward halves every 210,000 blocks; fees grow to compensate).

The spam test's economics: Andrej starts with 1000 TBB. He sends 4 TXs of 200 each. Without fees: 1000 - 800 = 200 remaining. With fees: 1000 - (800 + 4×50) = 1000 - 1000 = 0. He can't send a 5th TX. No balance. That's the whole point — spam becomes unaffordable.

#======================
#======================
#======================
#======================
#======================
#======================
#======================
#======================

-->
