# Blockchain in Go

A blockchain implementation in Go, as described in these articles:

1. [Basic Prototype](https://jeiwan.cc/posts/building-blockchain-in-go-part-1/)
2. [Proof-of-Work](https://jeiwan.cc/posts/building-blockchain-in-go-part-2/)
3. [Persistence and CLI](https://jeiwan.cc/posts/building-blockchain-in-go-part-3/)
4. [Transactions 1](https://jeiwan.cc/posts/building-blockchain-in-go-part-4/)
5. [Addresses](https://jeiwan.cc/posts/building-blockchain-in-go-part-5/)
6. [Transactions 2](https://jeiwan.cc/posts/building-blockchain-in-go-part-6/)
7. [Network](https://jeiwan.cc/posts/building-blockchain-in-go-part-7/)

using URPO model

## allowed usage:

### 1. create a tokoin
Can only be done by a owner
### 2, transfer a tokoin
Can only be done by a holder
### 3. edit a tokoin
The owner resend the new tokoin to the holder

Can only be done by the owner
### 4. redeem a tokoin
The visitor(also the holder) send the tokoin back to the owner

Can only be done by the visitor
### 5. discard a tokoin
Can only be done by the owner
