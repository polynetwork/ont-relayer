<h1 align="center">Ont Relayer</h1>

Ont Relayer is an important character of Poly cross-chain interactive protocol which is responsible for relaying cross-chain transaction from and to Ontology.

## Build From Source

### Prerequisites

- [Golang](https://golang.org/doc/install) version 1.14 or later

### Build

```shell
git clone https://github.com/polynetwork/ont-relayer.git
cd ont-relayer
go build -o ont-relayer main.go
```

After building the source code successfully,  you should see the executable program `ont-relayer`. 

## Run Relayer

Before you can run the relayer you will need to create a wallet file of PolyNetwork. After creation, you need to register it as a Relayer to Poly net and get consensus nodes approving your registeration. And then you can send transaction to Poly net and start relaying.

Before running, you need feed the configuration file `config.json`.

```
{
  "AliaJsonRpcAddress":"http://ip:20336", // poly node
  "SideJsonRpcAddress":"http://ontology:20336", // your ontology node
  "SideChainID": 3, // ontology chain id
  "AliaWalletFile": "./wallet1.dat", // poly wallet
  "SideWalletFile": "./wallet2.dat", // ontology wallet
  "DBPath": "boltdb", // DB path
  "ScanInterval": 1, // interval for scanning chains
  "RetryInterval": 1, // interval for retrying sending transactions
  "GasPrice":500, // gas price for ontology
  "GasLimit":200000, // gas limit for ontology
  "SideToAlliForceSyncHeight": 0, // start scanning height of ontology
  "AlliToSideForceSyncHeight": 0 // start scanning height of Poly
}
```

Now, you can start relayer as follow: 

```shell
./ont-relayer --ontpwd pwd  --alliapwd pwd
```

Flag `ontpwd` is the password for your ontology wallet and `alliapwd` is the password for your polly wallet. Relayer will generate logs under `./Log` and check relayer status by view log file.