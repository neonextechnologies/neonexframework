# Blockchain/Web3 Support Package

NeonexCore Web3 package provides comprehensive blockchain integration with support for Ethereum and EVM-compatible chains, smart contract interaction, NFT management, wallet integration, and Web3 authentication.

## Features

- **Multi-Chain Support**: Ethereum, Polygon, BSC, Arbitrum, Optimism, Avalanche, Fantom
- **Wallet Management**: Create, import, and manage wallets
- **Transaction Management**: Send, track, and confirm transactions
- **Smart Contract Interaction**: Deploy and interact with smart contracts
- **NFT Support**: ERC-721 operations and metadata management
- **Token Operations**: ERC-20 token transfers and approvals
- **Web3 Authentication**: Sign-in with Ethereum, WalletConnect, MetaMask
- **Gas Estimation**: Accurate gas price and limit estimation
- **Event Listening**: Watch and query contract events
- **Network Management**: Support for mainnet and testnet networks

## Installation

```go
import "github.com/neonexcore/pkg/web3"
```

Required dependencies:
```bash
go get github.com/ethereum/go-ethereum
```

## Quick Start

### Connect to Network

```go
// Create Web3 manager
manager := web3.NewWeb3Manager()

// Connect to Ethereum mainnet
config := web3.DefaultNetworkConfigs[web3.NetworkEthereum]
config.RPCURL = "https://mainnet.infura.io/v3/YOUR_API_KEY"

err := manager.Connect(config)
if err != nil {
    log.Fatal(err)
}

// Get client
client, err := manager.GetClient(web3.NetworkEthereum)
```

### Wallet Management

```go
// Create new wallet
wallet, err := web3.CreateWallet()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Address: %s\n", wallet.Address.Hex())
fmt.Printf("Private Key: %s\n", hex.EncodeToString(crypto.FromECDSA(wallet.PrivateKey)))

// Import existing wallet
privateKey := "your_private_key_hex"
wallet, err = web3.ImportWallet(privateKey)
if err != nil {
    log.Fatal(err)
}

// Get balance
balance, err := client.GetBalance(ctx, wallet.Address)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Balance: %s ETH\n", web3.WeiToEther(balance))
```

### Send Transaction

```go
ctx := context.Background()

// Prepare transaction
toAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
value := big.NewInt(1000000000000000000) // 1 ETH in wei

// Send transaction
tx, err := client.SendTransaction(ctx, wallet, toAddress, value, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Transaction hash: %s\n", tx.Hash.Hex())

// Wait for confirmation
receipt, err := client.WaitForTransaction(ctx, tx.Hash, 1)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Transaction confirmed in block: %d\n", receipt.BlockNumber.Uint64())
```

## Smart Contract Interaction

### Load Contract

```go
contractManager := web3.NewContractManager(client)

// Contract ABI (JSON)
abiJSON := `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"type":"function"}]`

// Load contract
contractAddress := common.HexToAddress("0x...")
contract, err := contractManager.LoadContract(contractAddress, abiJSON)
if err != nil {
    log.Fatal(err)
}
```

### Call Read Methods

```go
// Call view/pure function
results, err := contractManager.CallMethod(
    ctx, 
    contractAddress, 
    "balanceOf", 
    wallet.Address,
)
if err != nil {
    log.Fatal(err)
}

balance := results[0].(*big.Int)
fmt.Printf("Token balance: %s\n", balance.String())
```

### Call Write Methods

```go
// Send transaction to contract
amount := big.NewInt(1000000000000000000) // 1 token

tx, err := contractManager.SendMethod(
    ctx,
    wallet,
    contractAddress,
    "transfer",
    big.NewInt(0), // ETH value
    toAddress,
    amount,
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Transaction: %s\n", tx.Hash.Hex())
```

### Deploy Contract

```go
// Contract bytecode and ABI
abiJSON := `[...]`
bytecode := common.FromHex("0x60806040...")

// Constructor arguments
args := []interface{}{
    "MyToken",
    "MTK",
    big.NewInt(1000000),
}

// Deploy
tx, contractAddress, err := contractManager.DeployContract(
    ctx,
    wallet,
    abiJSON,
    bytecode,
    args...,
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Contract deployed at: %s\n", contractAddress.Hex())
```

## ERC-20 Token Operations

```go
tokenManager := web3.NewTokenManager(client, contractManager)

// Get token info
tokenAddress := common.HexToAddress("0x...")
token, err := tokenManager.GetTokenInfo(ctx, tokenAddress)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Token: %s (%s)\n", token.Name, token.Symbol)
fmt.Printf("Decimals: %d\n", token.Decimals)

// Get balance
balance, err := tokenManager.GetBalance(ctx, tokenAddress, wallet.Address)

// Transfer tokens
amount := big.NewInt(1000000) // Adjust for decimals
tx, err := tokenManager.Transfer(ctx, wallet, tokenAddress, toAddress, amount)

// Approve spending
tx, err = tokenManager.Approve(ctx, wallet, tokenAddress, spenderAddress, amount)

// Check allowance
allowance, err := tokenManager.GetAllowance(ctx, tokenAddress, wallet.Address, spenderAddress)
```

## NFT Operations

```go
nftManager := web3.NewNFTManager(client, contractManager)

// Get NFT details
nftAddress := common.HexToAddress("0x...")
tokenID := big.NewInt(1)

nft, err := nftManager.GetNFT(ctx, nftAddress, tokenID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Owner: %s\n", nft.Owner.Hex())
fmt.Printf("Token URI: %s\n", nft.TokenURI)

// Transfer NFT
tx, err := nftManager.TransferNFT(
    ctx,
    wallet,
    nftAddress,
    wallet.Address, // from
    toAddress,      // to
    tokenID,
)

// Mint NFT
tokenURI := "ipfs://QmX..."
tx, err = nftManager.MintNFT(ctx, wallet, nftAddress, toAddress, tokenID, tokenURI)

// Approve NFT
tx, err = nftManager.ApproveNFT(ctx, wallet, nftAddress, approvedAddress, tokenID)

// Set approval for all
tx, err = nftManager.SetApprovalForAll(ctx, wallet, nftAddress, operatorAddress, true)
```

## Web3 Authentication

### Sign-In with Ethereum

```go
auth := web3.NewWeb3Auth()

// 1. Generate challenge
address := common.HexToAddress("0x...")
challenge, err := auth.GenerateChallenge(address)
if err != nil {
    log.Fatal(err)
}

// 2. User signs message with wallet (client-side)
// signature := await ethereum.request({method: 'personal_sign', params: [challenge.Message, address]})

// 3. Verify and create session
signature := "0x..." // From client
session, err := auth.Authenticate(ctx, challenge.Nonce, signature, address)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Session ID: %s\n", session.ID)
fmt.Printf("Expires: %s\n", session.ExpiresAt)

// 4. Verify session
session, err = auth.GetSession(session.ID)
if err != nil {
    log.Fatal(err)
}

// 5. Revoke session
err = auth.RevokeSession(session.ID)
```

### MetaMask Integration

```go
metaMask := web3.NewMetaMaskAuth(auth)

// Request challenge
challenge, err := metaMask.RequestChallenge(address)

// User signs with MetaMask
// const signature = await ethereum.request({
//     method: 'personal_sign',
//     params: [challenge.Message, address]
// })

// Verify signature
session, err := metaMask.VerifyMetaMaskSignature(challenge, signature)
```

### WalletConnect

```go
wcManager := web3.NewWalletConnectManager()

// Create connection
chainID := 1 // Ethereum mainnet
connection, err := wcManager.CreateConnection(address, chainID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Session ID: %s\n", connection.SessionID)

// Get connection
connection, err = wcManager.GetConnection(connection.SessionID)

// Disconnect
err = wcManager.DisconnectSession(connection.SessionID)
```

## Contract Events

### Watch Events

```go
// Watch for Transfer events
eventChan, err := contractManager.WatchEvents(
    ctx,
    tokenAddress,
    "Transfer",
    0, // from block
)
if err != nil {
    log.Fatal(err)
}

// Listen for events
for event := range eventChan {
    fmt.Printf("Event: %s\n", event.Name)
    fmt.Printf("Block: %d\n", event.BlockNumber)
    fmt.Printf("Tx: %s\n", event.TxHash.Hex())
}
```

### Get Past Events

```go
// Get historical events
events, err := contractManager.GetPastEvents(
    ctx,
    tokenAddress,
    "Transfer",
    0,      // from block
    100000, // to block
)
if err != nil {
    log.Fatal(err)
}

for _, event := range events {
    fmt.Printf("Event at block %d: %+v\n", event.BlockNumber, event.Data)
}
```

## Gas Management

```go
gasEstimator := web3.NewGasEstimator(client)

// Estimate gas for transfer
gasLimit, err := gasEstimator.EstimateTransfer(ctx)

// Estimate gas for token transfer
gasLimit, err = gasEstimator.EstimateTokenTransfer(ctx)

// Get current gas price
gasPrice, err := gasEstimator.GetGasPrice(ctx)
fmt.Printf("Gas price: %s Gwei\n", web3.WeiToGwei(gasPrice))

// Calculate transaction cost
cost, err := gasEstimator.CalculateTxCost(ctx, gasLimit)
fmt.Printf("Estimated cost: %s ETH\n", web3.WeiToEther(cost))
```

## Network Configuration

### Supported Networks

```go
// Mainnet networks
web3.NetworkEthereum   // Ethereum mainnet
web3.NetworkPolygon    // Polygon mainnet
web3.NetworkBSC        // Binance Smart Chain
web3.NetworkArbitrum   // Arbitrum
web3.NetworkOptimism   // Optimism
web3.NetworkAvalanche  // Avalanche C-Chain
web3.NetworkFantom     // Fantom Opera

// Testnet networks
web3.NetworkGoerli     // Ethereum Goerli
web3.NetworkSepolia    // Ethereum Sepolia
web3.NetworkMumbai     // Polygon Mumbai
web3.NetworkBSCTestnet // BSC Testnet
```

### Custom Network

```go
customConfig := &web3.NetworkConfig{
    Network:    "custom",
    ChainID:    big.NewInt(1234),
    RPCURL:     "https://custom-rpc.example.com",
    WSURL:      "wss://custom-ws.example.com",
    Explorer:   "https://explorer.example.com",
    NativeCoin: "CUSTOM",
}

err := manager.Connect(customConfig)
```

## Best Practices

1. **Private Key Security**: Never hardcode private keys, use environment variables
2. **Gas Management**: Always estimate gas and set appropriate limits
3. **Transaction Confirmation**: Wait for confirmations before considering tx final
4. **Error Handling**: Handle network errors and retry logic
5. **Rate Limiting**: Implement rate limiting for RPC requests
6. **Connection Management**: Reuse clients, close when done
7. **Testnet First**: Test on testnets before mainnet deployment
8. **Signature Verification**: Always verify signatures on server-side
9. **Session Management**: Set appropriate session expiration times
10. **Event Monitoring**: Use WebSocket for real-time events

## Utility Functions

```go
// Convert Wei to Ether
ether := web3.WeiToEther(weiAmount)

// Convert Ether to Wei
wei := web3.EtherToWei(etherAmount)

// Convert Wei to Gwei
gwei := web3.WeiToGwei(weiAmount)

// Format address
checksum := web3.ToChecksumAddress(address)

// Validate address
isValid := web3.IsValidAddress(addressString)
```

## Complete Example

See `examples/web3_example.go` for comprehensive examples including:
- Network connection and management
- Wallet creation and import
- Sending transactions
- Smart contract deployment and interaction
- ERC-20 token operations
- NFT management
- Web3 authentication
- Event listening
- Gas estimation

## Architecture

- **Web3Manager**: Multi-network connection manager
- **Web3Client**: Blockchain client for single network
- **ContractManager**: Smart contract interaction layer
- **TokenManager**: ERC-20 token operations
- **NFTManager**: ERC-721 NFT operations
- **Web3Auth**: Web3 authentication system
- **WalletConnectManager**: WalletConnect integration
- **GasEstimator**: Gas price and limit estimation

## Security Considerations

- Private keys are never logged or exposed
- All signatures verified server-side
- Session management with expiration
- Challenge-based authentication
- Nonce for replay protection
- Address checksum validation
- Transaction verification before execution

## Performance

- Connection pooling for multiple requests
- Caching for frequently accessed data
- Batch requests where possible
- WebSocket for real-time updates
- Efficient event filtering
- Minimal RPC calls

## Thread Safety

All components are thread-safe with proper synchronization using RWMutex.
