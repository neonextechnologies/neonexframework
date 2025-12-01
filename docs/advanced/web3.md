# Web3 & Blockchain

Integrate blockchain and Web3 features into your NeonEx applications. Learn wallet connections, smart contracts, token operations, and NFT support.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [Wallet Connection](#wallet-connection)
- [Smart Contracts](#smart-contracts)
- [Token Operations](#token-operations)
- [NFT Support](#nft-support)
- [Blockchain Authentication](#blockchain-authentication)
- [Best Practices](#best-practices)

## Introduction

NeonEx provides comprehensive Web3 integration with:

- **Ethereum Support**: Mainnet, testnets, and layer 2
- **Wallet Connect**: MetaMask, WalletConnect, Coinbase Wallet
- **Smart Contracts**: Deploy and interact with contracts
- **ERC-20 Tokens**: Token transfers and management
- **ERC-721 NFTs**: Mint, transfer, and query NFTs
- **Blockchain Auth**: Sign-in with Ethereum

## Quick Start

```go
package main

import (
    "context"
    "math/big"
    "neonex/core/pkg/web3"
    "github.com/ethereum/go-ethereum/common"
)

func main() {
    // Connect to Ethereum node
    client, err := web3.NewClient(&web3.Config{
        RPCURL:     "https://mainnet.infura.io/v3/YOUR_KEY",
        PrivateKey: "0x...",
    })
    
    if err != nil {
        panic(err)
    }
    
    // Get account balance
    ctx := context.Background()
    address := common.HexToAddress("0x...")
    
    balance, err := client.BalanceAt(ctx, address, nil)
    if err != nil {
        panic(err)
    }
    
    println("Balance:", web3.WeiToEther(balance), "ETH")
}
```

## Wallet Connection

### Client Setup

```go
type Web3Config struct {
    RPCURL      string
    PrivateKey  string
    ChainID     int64
    GasLimit    uint64
    MaxGasPrice *big.Int
}

config := &Web3Config{
    RPCURL:      "https://mainnet.infura.io/v3/YOUR_KEY",
    PrivateKey:  os.Getenv("PRIVATE_KEY"),
    ChainID:     1, // Mainnet
    GasLimit:    300000,
    MaxGasPrice: big.NewInt(100000000000), // 100 gwei
}

client, err := web3.NewClient(config)
```

### Multiple Networks

```go
type NetworkConfig struct {
    Name    string
    ChainID int64
    RPCURL  string
}

var networks = map[string]NetworkConfig{
    "mainnet": {
        Name:    "Ethereum Mainnet",
        ChainID: 1,
        RPCURL:  "https://mainnet.infura.io/v3/YOUR_KEY",
    },
    "goerli": {
        Name:    "Goerli Testnet",
        ChainID: 5,
        RPCURL:  "https://goerli.infura.io/v3/YOUR_KEY",
    },
    "polygon": {
        Name:    "Polygon Mainnet",
        ChainID: 137,
        RPCURL:  "https://polygon-rpc.com",
    },
}

func GetClient(network string) (*web3.Client, error) {
    config := networks[network]
    return web3.NewClient(&web3.Config{
        RPCURL:  config.RPCURL,
        ChainID: config.ChainID,
    })
}
```

### Wallet Operations

```go
type Wallet struct {
    client     *web3.Client
    privateKey *ecdsa.PrivateKey
    address    common.Address
}

func NewWallet(client *web3.Client, privateKeyHex string) (*Wallet, error) {
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        return nil, err
    }
    
    publicKey := privateKey.Public()
    publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
    address := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    return &Wallet{
        client:     client,
        privateKey: privateKey,
        address:    address,
    }, nil
}

func (w *Wallet) GetBalance(ctx context.Context) (*big.Int, error) {
    return w.client.BalanceAt(ctx, w.address, nil)
}

func (w *Wallet) SendETH(ctx context.Context, to common.Address, amount *big.Int) (common.Hash, error) {
    nonce, err := w.client.PendingNonceAt(ctx, w.address)
    if err != nil {
        return common.Hash{}, err
    }
    
    gasPrice, err := w.client.SuggestGasPrice(ctx)
    if err != nil {
        return common.Hash{}, err
    }
    
    tx := types.NewTransaction(nonce, to, amount, 21000, gasPrice, nil)
    
    chainID, _ := w.client.NetworkID(ctx)
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), w.privateKey)
    if err != nil {
        return common.Hash{}, err
    }
    
    err = w.client.SendTransaction(ctx, signedTx)
    if err != nil {
        return common.Hash{}, err
    }
    
    return signedTx.Hash(), nil
}
```

## Smart Contracts

### Contract Deployment

```solidity
// SimpleStorage.sol
pragma solidity ^0.8.0;

contract SimpleStorage {
    uint256 private value;
    
    event ValueChanged(uint256 newValue);
    
    function setValue(uint256 newValue) public {
        value = newValue;
        emit ValueChanged(newValue);
    }
    
    function getValue() public view returns (uint256) {
        return value;
    }
}
```

```go
func DeployContract(client *web3.Client, wallet *Wallet, bytecode string) (common.Address, error) {
    ctx := context.Background()
    
    nonce, _ := client.PendingNonceAt(ctx, wallet.address)
    gasPrice, _ := client.SuggestGasPrice(ctx)
    
    // Create contract creation transaction
    data := common.Hex2Bytes(bytecode)
    
    tx := types.NewContractCreation(
        nonce,
        big.NewInt(0),
        3000000,
        gasPrice,
        data,
    )
    
    chainID, _ := client.NetworkID(ctx)
    signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainID), wallet.privateKey)
    
    err := client.SendTransaction(ctx, signedTx)
    if err != nil {
        return common.Address{}, err
    }
    
    // Wait for deployment
    receipt, err := bind.WaitMined(ctx, client, signedTx)
    if err != nil {
        return common.Address{}, err
    }
    
    return receipt.ContractAddress, nil
}
```

### Contract Interaction

```go
type Contract struct {
    client   *web3.Client
    address  common.Address
    abi      abi.ABI
}

func NewContract(client *web3.Client, address common.Address, abiJSON string) (*Contract, error) {
    contractABI, err := abi.JSON(strings.NewReader(abiJSON))
    if err != nil {
        return nil, err
    }
    
    return &Contract{
        client:  client,
        address: address,
        abi:     contractABI,
    }, nil
}

// Call read-only method
func (c *Contract) Call(ctx context.Context, method string, args ...interface{}) ([]interface{}, error) {
    data, err := c.abi.Pack(method, args...)
    if err != nil {
        return nil, err
    }
    
    msg := ethereum.CallMsg{
        To:   &c.address,
        Data: data,
    }
    
    result, err := c.client.CallContract(ctx, msg, nil)
    if err != nil {
        return nil, err
    }
    
    return c.abi.Unpack(method, result)
}

// Send transaction
func (c *Contract) Transact(ctx context.Context, wallet *Wallet, method string, args ...interface{}) (common.Hash, error) {
    data, err := c.abi.Pack(method, args...)
    if err != nil {
        return common.Hash{}, err
    }
    
    nonce, _ := c.client.PendingNonceAt(ctx, wallet.address)
    gasPrice, _ := c.client.SuggestGasPrice(ctx)
    
    tx := types.NewTransaction(
        nonce,
        c.address,
        big.NewInt(0),
        300000,
        gasPrice,
        data,
    )
    
    chainID, _ := c.client.NetworkID(ctx)
    signedTx, _ := types.SignTx(tx, types.NewEIP155Signer(chainID), wallet.privateKey)
    
    err = c.client.SendTransaction(ctx, signedTx)
    return signedTx.Hash(), err
}
```

### Event Listening

```go
func WatchEvents(client *web3.Client, contractAddress common.Address) {
    query := ethereum.FilterQuery{
        Addresses: []common.Address{contractAddress},
        FromBlock: big.NewInt(0),
        ToBlock:   nil,
    }
    
    logs := make(chan types.Log)
    sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
    if err != nil {
        log.Fatal(err)
    }
    
    for {
        select {
        case err := <-sub.Err():
            log.Fatal(err)
        case vLog := <-logs:
            // Parse event
            event := parseEvent(vLog)
            log.Printf("Event: %+v", event)
        }
    }
}
```

## Token Operations

### ERC-20 Interface

```go
const ERC20ABI = `[
    {
        "constant": true,
        "inputs": [],
        "name": "name",
        "outputs": [{"name": "", "type": "string"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "symbol",
        "outputs": [{"name": "", "type": "string"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "decimals",
        "outputs": [{"name": "", "type": "uint8"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "owner", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [
            {"name": "to", "type": "address"},
            {"name": "amount", "type": "uint256"}
        ],
        "name": "transfer",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    }
]`
```

### Token Service

```go
type TokenService struct {
    client   *web3.Client
    contract *Contract
}

func NewTokenService(client *web3.Client, tokenAddress common.Address) (*TokenService, error) {
    contract, err := NewContract(client, tokenAddress, ERC20ABI)
    if err != nil {
        return nil, err
    }
    
    return &TokenService{
        client:   client,
        contract: contract,
    }, nil
}

func (ts *TokenService) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
    result, err := ts.contract.Call(ctx, "balanceOf", address)
    if err != nil {
        return nil, err
    }
    
    return result[0].(*big.Int), nil
}

func (ts *TokenService) Transfer(ctx context.Context, wallet *Wallet, to common.Address, amount *big.Int) (common.Hash, error) {
    return ts.contract.Transact(ctx, wallet, "transfer", to, amount)
}

func (ts *TokenService) GetTokenInfo(ctx context.Context) (*TokenInfo, error) {
    name, _ := ts.contract.Call(ctx, "name")
    symbol, _ := ts.contract.Call(ctx, "symbol")
    decimals, _ := ts.contract.Call(ctx, "decimals")
    
    return &TokenInfo{
        Name:     name[0].(string),
        Symbol:   symbol[0].(string),
        Decimals: decimals[0].(uint8),
    }, nil
}
```

### Token Transfers

```go
func TransferTokens(ctx context.Context, ts *TokenService, wallet *Wallet, to string, amount float64) error {
    // Get token decimals
    info, err := ts.GetTokenInfo(ctx)
    if err != nil {
        return err
    }
    
    // Convert amount to token units
    tokenAmount := new(big.Int)
    tokenAmount.SetString(fmt.Sprintf("%.0f", amount*math.Pow10(int(info.Decimals))), 10)
    
    // Send transaction
    txHash, err := ts.Transfer(ctx, wallet, common.HexToAddress(to), tokenAmount)
    if err != nil {
        return err
    }
    
    log.Printf("Token transfer sent: %s", txHash.Hex())
    
    // Wait for confirmation
    receipt, err := bind.WaitMined(ctx, ts.client, txHash)
    if err != nil {
        return err
    }
    
    if receipt.Status == 1 {
        log.Println("Transfer confirmed")
    } else {
        return fmt.Errorf("transfer failed")
    }
    
    return nil
}
```

## NFT Support

### ERC-721 Interface

```go
const ERC721ABI = `[
    {
        "constant": true,
        "inputs": [{"name": "tokenId", "type": "uint256"}],
        "name": "ownerOf",
        "outputs": [{"name": "", "type": "address"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "tokenId", "type": "uint256"}],
        "name": "tokenURI",
        "outputs": [{"name": "", "type": "string"}],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [
            {"name": "from", "type": "address"},
            {"name": "to", "type": "address"},
            {"name": "tokenId", "type": "uint256"}
        ],
        "name": "transferFrom",
        "outputs": [],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [
            {"name": "to", "type": "address"},
            {"name": "tokenId", "type": "uint256"}
        ],
        "name": "mint",
        "outputs": [],
        "type": "function"
    }
]`
```

### NFT Service

```go
type NFTService struct {
    client   *web3.Client
    contract *Contract
}

type NFTMetadata struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Image       string `json:"image"`
    Attributes  []struct {
        TraitType string `json:"trait_type"`
        Value     string `json:"value"`
    } `json:"attributes"`
}

func (ns *NFTService) GetOwner(ctx context.Context, tokenID *big.Int) (common.Address, error) {
    result, err := ns.contract.Call(ctx, "ownerOf", tokenID)
    if err != nil {
        return common.Address{}, err
    }
    
    return result[0].(common.Address), nil
}

func (ns *NFTService) GetTokenURI(ctx context.Context, tokenID *big.Int) (string, error) {
    result, err := ns.contract.Call(ctx, "tokenURI", tokenID)
    if err != nil {
        return "", err
    }
    
    return result[0].(string), nil
}

func (ns *NFTService) GetMetadata(ctx context.Context, tokenID *big.Int) (*NFTMetadata, error) {
    uri, err := ns.GetTokenURI(ctx, tokenID)
    if err != nil {
        return nil, err
    }
    
    // Fetch metadata from IPFS or HTTP
    resp, err := http.Get(uri)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var metadata NFTMetadata
    if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
        return nil, err
    }
    
    return &metadata, nil
}

func (ns *NFTService) MintNFT(ctx context.Context, wallet *Wallet, to common.Address, tokenID *big.Int) (common.Hash, error) {
    return ns.contract.Transact(ctx, wallet, "mint", to, tokenID)
}

func (ns *NFTService) TransferNFT(ctx context.Context, wallet *Wallet, from, to common.Address, tokenID *big.Int) (common.Hash, error) {
    return ns.contract.Transact(ctx, wallet, "transferFrom", from, to, tokenID)
}
```

### NFT Marketplace

```go
type NFTMarketplace struct {
    nftService *NFTService
    db         *gorm.DB
}

type NFTListing struct {
    ID        int64
    TokenID   string
    Owner     string
    Price     string // wei
    Status    string // listed, sold, cancelled
    CreatedAt time.Time
}

func (nm *NFTMarketplace) ListNFT(ctx context.Context, wallet *Wallet, tokenID *big.Int, price *big.Int) error {
    // Verify ownership
    owner, err := nm.nftService.GetOwner(ctx, tokenID)
    if err != nil {
        return err
    }
    
    if owner != wallet.address {
        return fmt.Errorf("not the owner")
    }
    
    // Create listing
    listing := &NFTListing{
        TokenID: tokenID.String(),
        Owner:   wallet.address.Hex(),
        Price:   price.String(),
        Status:  "listed",
    }
    
    return nm.db.Create(listing).Error
}

func (nm *NFTMarketplace) BuyNFT(ctx context.Context, buyer *Wallet, listingID int64) error {
    var listing NFTListing
    if err := nm.db.First(&listing, listingID).Error; err != nil {
        return err
    }
    
    if listing.Status != "listed" {
        return fmt.Errorf("NFT not available")
    }
    
    // Transfer NFT
    tokenID := new(big.Int)
    tokenID.SetString(listing.TokenID, 10)
    
    txHash, err := nm.nftService.TransferNFT(
        ctx,
        buyer,
        common.HexToAddress(listing.Owner),
        buyer.address,
        tokenID,
    )
    
    if err != nil {
        return err
    }
    
    log.Printf("NFT transfer: %s", txHash.Hex())
    
    // Update listing
    listing.Status = "sold"
    return nm.db.Save(&listing).Error
}
```

## Blockchain Authentication

### Sign-In with Ethereum

```go
type Web3Auth struct {
    sessions map[string]*Session
    mu       sync.RWMutex
}

type Session struct {
    Address   string
    Nonce     string
    ExpiresAt time.Time
}

func (wa *Web3Auth) GenerateNonce(address string) string {
    nonce := uuid.New().String()
    
    wa.mu.Lock()
    defer wa.mu.Unlock()
    
    wa.sessions[address] = &Session{
        Address:   address,
        Nonce:     nonce,
        ExpiresAt: time.Now().Add(5 * time.Minute),
    }
    
    return nonce
}

func (wa *Web3Auth) VerifySignature(address, signature, nonce string) (bool, error) {
    wa.mu.RLock()
    session, exists := wa.sessions[address]
    wa.mu.RUnlock()
    
    if !exists {
        return false, fmt.Errorf("session not found")
    }
    
    if session.Nonce != nonce {
        return false, fmt.Errorf("invalid nonce")
    }
    
    if time.Now().After(session.ExpiresAt) {
        return false, fmt.Errorf("nonce expired")
    }
    
    // Verify signature
    message := fmt.Sprintf("Sign this message to authenticate: %s", nonce)
    hash := crypto.Keccak256Hash([]byte(message))
    
    sig := common.Hex2Bytes(signature)
    if len(sig) != 65 {
        return false, fmt.Errorf("invalid signature length")
    }
    
    // Recover public key
    pubKey, err := crypto.SigToPub(hash.Bytes(), sig)
    if err != nil {
        return false, err
    }
    
    recoveredAddr := crypto.PubkeyToAddress(*pubKey)
    
    return recoveredAddr.Hex() == address, nil
}
```

### HTTP Handler

```go
func (h *Handler) Web3Login(c echo.Context) error {
    var req struct {
        Address string `json:"address"`
    }
    
    if err := c.Bind(&req); err != nil {
        return err
    }
    
    nonce := h.web3Auth.GenerateNonce(req.Address)
    
    return c.JSON(http.StatusOK, map[string]string{
        "nonce": nonce,
    })
}

func (h *Handler) Web3Verify(c echo.Context) error {
    var req struct {
        Address   string `json:"address"`
        Signature string `json:"signature"`
        Nonce     string `json:"nonce"`
    }
    
    if err := c.Bind(&req); err != nil {
        return err
    }
    
    valid, err := h.web3Auth.VerifySignature(req.Address, req.Signature, req.Nonce)
    if err != nil || !valid {
        return echo.NewHTTPError(http.StatusUnauthorized, "Invalid signature")
    }
    
    // Create session
    token, _ := generateJWT(req.Address)
    
    return c.JSON(http.StatusOK, map[string]string{
        "token": token,
    })
}
```

## Best Practices

### 1. Error Handling

```go
func SafeTransfer(ctx context.Context, wallet *Wallet, to common.Address, amount *big.Int) error {
    // Check balance
    balance, err := wallet.GetBalance(ctx)
    if err != nil {
        return fmt.Errorf("failed to get balance: %w", err)
    }
    
    if balance.Cmp(amount) < 0 {
        return fmt.Errorf("insufficient balance")
    }
    
    // Estimate gas
    gasPrice, _ := wallet.client.SuggestGasPrice(ctx)
    gasCost := new(big.Int).Mul(big.NewInt(21000), gasPrice)
    
    total := new(big.Int).Add(amount, gasCost)
    if balance.Cmp(total) < 0 {
        return fmt.Errorf("insufficient balance for gas")
    }
    
    // Send transaction
    txHash, err := wallet.SendETH(ctx, to, amount)
    if err != nil {
        return fmt.Errorf("transaction failed: %w", err)
    }
    
    log.Printf("Transaction sent: %s", txHash.Hex())
    return nil
}
```

### 2. Transaction Monitoring

```go
func WaitForTransaction(client *web3.Client, txHash common.Hash, timeout time.Duration) (*types.Receipt, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return nil, fmt.Errorf("timeout waiting for transaction")
        case <-ticker.C:
            receipt, err := client.TransactionReceipt(ctx, txHash)
            if err == nil {
                return receipt, nil
            }
        }
    }
}
```

### 3. Gas Optimization

```go
func OptimizedTransfer(wallet *Wallet, to common.Address, amount *big.Int) error {
    ctx := context.Background()
    
    // Use EIP-1559 for gas optimization
    tip, _ := wallet.client.SuggestGasTipCap(ctx)
    baseFee, _ := wallet.client.SuggestGasPrice(ctx)
    
    maxPriorityFee := tip
    maxFee := new(big.Int).Add(baseFee, tip)
    
    tx := types.NewTx(&types.DynamicFeeTx{
        ChainID:   big.NewInt(1),
        Nonce:     nonce,
        To:        &to,
        Value:     amount,
        Gas:       21000,
        GasTipCap: maxPriorityFee,
        GasFeeCap: maxFee,
    })
    
    // Sign and send
    signedTx, _ := types.SignTx(tx, types.LatestSignerForChainID(big.NewInt(1)), wallet.privateKey)
    return wallet.client.SendTransaction(ctx, signedTx)
}
```

### 4. Testing

```go
func TestTokenTransfer(t *testing.T) {
    // Use testnet or local node
    client, _ := web3.NewClient(&web3.Config{
        RPCURL: "http://localhost:8545",
    })
    
    wallet, _ := NewWallet(client, "test_private_key")
    tokenService, _ := NewTokenService(client, testTokenAddress)
    
    initialBalance, _ := tokenService.GetBalance(ctx, wallet.address)
    
    // Transfer tokens
    amount := big.NewInt(1000000)
    txHash, err := tokenService.Transfer(ctx, wallet, testRecipient, amount)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, txHash)
    
    // Verify balance changed
    finalBalance, _ := tokenService.GetBalance(ctx, wallet.address)
    expected := new(big.Int).Sub(initialBalance, amount)
    
    assert.Equal(t, expected, finalBalance)
}
```

---

**Next Steps:**
- Learn about [Database](../database/models.md) for storing blockchain data
- Explore [Queue](queue.md) for async blockchain monitoring
- See [Cache](cache.md) for caching blockchain queries

**Related Topics:**
- [Ethereum Documentation](https://ethereum.org/developers)
- [Web3.js](https://web3js.readthedocs.io/)
- [go-ethereum](https://geth.ethereum.org/docs)
