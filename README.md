# invoice-chain

### Build

#### Invoice Chain Ledger Server

```sh
cd cmd
go get github.com/dgraph-io/badger
go get github.com/spf13/viper
go get github.com/sithu/invoice-chain.git
go build -o qbchain
```

#### Build CLI tool

```sh
$ cd cli
$ go build -o qb
```

# Usage

## Generate Key Pair

```sh
./qb genkeys
```

## Submit a New Transaction

```sh
./qb submit
```

## Starting a node

You can start as many nodes as you want with the following command

`./qbchain -port=<port-number>`


## Endpoints


### Requesting the Blockchain of a node

* `GET 127.0.0.1:8000/chain`

### Mining some coins

* `GET 127.0.0.1:8000/mine`

### Adding a new transaction

* `POST 127.0.0.1:8000/transactions/new`

* __Body__: A transaction to be added

  ```json
  {
    "sender": "sender-address-te33412uywq89234g",
    "recipient": "recipient-address-j3h45jk23hjk543gf",
    "amount": 1000
  }
  ```

### Register a new node in the network
Currently you must add each new node to each running node.

* `POST 127.0.0.1:8000/nodes/register`

* __Body__: A list of nodes to add

  ```json
  {
     "nodes": ["http://127.0.0.1:8001", <more-nodes>]
  }
  ```

### Resolving Blockchain differences in each node

* `GET 127.0.0.1:8000/nodes/resolve`
