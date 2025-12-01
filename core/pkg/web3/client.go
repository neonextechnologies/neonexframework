package web3

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Network represents a blockchain network
type Network string

const (
	NetworkEthereum      Network = "ethereum"
	NetworkPolygon       Network = "polygon"
	NetworkBSC           Network = "bsc"
	NetworkArbitrum      Network = "arbitrum"
	NetworkOptimism      Network = "optimism"
	NetworkAvalanche     Network = "avalanche"
	NetworkFantom        Network = "fantom"
	NetworkGoerli        Network = "goerli"        // Testnet
	NetworkSepolia       Network = "sepolia"       // Testnet
	NetworkMumbai        Network = "mumbai"        // Polygon Testnet
	NetworkBSCTestnet    Network = "bsc-testnet"
)

// NetworkConfig network configuration
type NetworkConfig struct {
	Network    Network
	ChainID    *big.Int
	RPCURL     string
	WSURL      string
	Explorer   string
	NativeCoin string
}

// Web3Client blockchain client
type Web3Client struct {
	config      *NetworkConfig
	client      *ethclient.Client
	wsClient    *ethclient.Client
	chainID     *big.Int
	mu          sync.RWMutex
}

// TransactionStatus transaction status
type TransactionStatus string

const (
	TxStatusPending   TransactionStatus = "pending"
	TxStatusConfirmed TransactionStatus = "confirmed"
	TxStatusFailed    TransactionStatus = "failed"
)

// Transaction transaction details
type Transaction struct {
	Hash        common.Hash
	From        common.Address
	To          *common.Address
	Value       *big.Int
	Gas         uint64
	GasPrice    *big.Int
	Nonce       uint64
	Data        []byte
	Status      TransactionStatus
	BlockNumber *big.Int
	Timestamp   time.Time
}

// Wallet wallet management
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    common.Address
}

// Web3Manager manages Web3 connections
type Web3Manager struct {
	clients map[Network]*Web3Client
	mu      sync.RWMutex
}

// DefaultNetworkConfigs default network configurations
var DefaultNetworkConfigs = map[Network]*NetworkConfig{
	NetworkEthereum: {
		Network:    NetworkEthereum,
		ChainID:    big.NewInt(1),
		RPCURL:     "https://mainnet.infura.io/v3/YOUR_API_KEY",
		Explorer:   "https://etherscan.io",
		NativeCoin: "ETH",
	},
	NetworkPolygon: {
		Network:    NetworkPolygon,
		ChainID:    big.NewInt(137),
		RPCURL:     "https://polygon-rpc.com",
		Explorer:   "https://polygonscan.com",
		NativeCoin: "MATIC",
	},
	NetworkBSC: {
		Network:    NetworkBSC,
		ChainID:    big.NewInt(56),
		RPCURL:     "https://bsc-dataseed.binance.org",
		Explorer:   "https://bscscan.com",
		NativeCoin: "BNB",
	},
	NetworkGoerli: {
		Network:    NetworkGoerli,
		ChainID:    big.NewInt(5),
		RPCURL:     "https://goerli.infura.io/v3/YOUR_API_KEY",
		Explorer:   "https://goerli.etherscan.io",
		NativeCoin: "GoerliETH",
	},
	NetworkSepolia: {
		Network:    NetworkSepolia,
		ChainID:    big.NewInt(11155111),
		RPCURL:     "https://sepolia.infura.io/v3/YOUR_API_KEY",
		Explorer:   "https://sepolia.etherscan.io",
		NativeCoin: "SepoliaETH",
	},
}

// NewWeb3Manager creates a new Web3 manager
func NewWeb3Manager() *Web3Manager {
	return &Web3Manager{
		clients: make(map[Network]*Web3Client),
	}
}

// Connect connects to a blockchain network
func (m *Web3Manager) Connect(config *NetworkConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, err := ethclient.Dial(config.RPCURL)
	if err != nil {
		return fmt.Errorf("failed to connect to network %s: %w", config.Network, err)
	}

	web3Client := &Web3Client{
		config:  config,
		client:  client,
		chainID: config.ChainID,
	}

	// Connect WebSocket if available
	if config.WSURL != "" {
		wsClient, err := ethclient.Dial(config.WSURL)
		if err == nil {
			web3Client.wsClient = wsClient
		}
	}

	m.clients[config.Network] = web3Client
	return nil
}

// GetClient gets a client for a network
func (m *Web3Manager) GetClient(network Network) (*Web3Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[network]
	if !exists {
		return nil, fmt.Errorf("client not found for network: %s", network)
	}

	return client, nil
}

// Disconnect disconnects from a network
func (m *Web3Manager) Disconnect(network Network) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[network]
	if !exists {
		return fmt.Errorf("client not found for network: %s", network)
	}

	client.client.Close()
	if client.wsClient != nil {
		client.wsClient.Close()
	}

	delete(m.clients, network)
	return nil
}

// CreateWallet creates a new wallet
func CreateWallet() (*Wallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKeyECDSA,
		Address:    address,
	}, nil
}

// ImportWallet imports a wallet from private key
func ImportWallet(privateKeyHex string) (*Wallet, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKeyECDSA,
		Address:    address,
	}, nil
}

// GetBalance gets account balance
func (c *Web3Client) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	balance, err := c.client.BalanceAt(ctx, address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

// GetNonce gets account nonce
func (c *Web3Client) GetNonce(ctx context.Context, address common.Address) (uint64, error) {
	nonce, err := c.client.PendingNonceAt(ctx, address)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}
	return nonce, nil
}

// EstimateGas estimates gas for a transaction
func (c *Web3Client) EstimateGas(ctx context.Context, msg interface{}) (uint64, error) {
	// Type assertion for ethereum.CallMsg would go here
	// Simplified for this example
	return 21000, nil
}

// SuggestGasPrice suggests gas price
func (c *Web3Client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %w", err)
	}
	return gasPrice, nil
}

// SendTransaction sends a transaction
func (c *Web3Client) SendTransaction(ctx context.Context, wallet *Wallet, to common.Address, value *big.Int, data []byte) (*Transaction, error) {
	nonce, err := c.GetNonce(ctx, wallet.Address)
	if err != nil {
		return nil, err
	}

	gasPrice, err := c.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	gasLimit := uint64(21000)
	if len(data) > 0 {
		gasLimit = uint64(100000) // Higher for contract interaction
	}

	tx := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, data)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainID), wallet.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = c.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return &Transaction{
		Hash:      signedTx.Hash(),
		From:      wallet.Address,
		To:        &to,
		Value:     value,
		Gas:       gasLimit,
		GasPrice:  gasPrice,
		Nonce:     nonce,
		Data:      data,
		Status:    TxStatusPending,
		Timestamp: time.Now(),
	}, nil
}

// GetTransaction gets transaction by hash
func (c *Web3Client) GetTransaction(ctx context.Context, hash common.Hash) (*Transaction, error) {
	tx, isPending, err := c.client.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	transaction := &Transaction{
		Hash:     tx.Hash(),
		To:       tx.To(),
		Value:    tx.Value(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Nonce:    tx.Nonce(),
		Data:     tx.Data(),
	}

	if isPending {
		transaction.Status = TxStatusPending
	} else {
		receipt, err := c.client.TransactionReceipt(ctx, hash)
		if err != nil {
			return nil, fmt.Errorf("failed to get receipt: %w", err)
		}

		if receipt.Status == 1 {
			transaction.Status = TxStatusConfirmed
		} else {
			transaction.Status = TxStatusFailed
		}

		transaction.BlockNumber = receipt.BlockNumber
	}

	return transaction, nil
}

// WaitForTransaction waits for transaction confirmation
func (c *Web3Client) WaitForTransaction(ctx context.Context, hash common.Hash, confirmations uint64) (*types.Receipt, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			receipt, err := c.client.TransactionReceipt(ctx, hash)
			if err != nil {
				continue
			}

			currentBlock, err := c.client.BlockNumber(ctx)
			if err != nil {
				continue
			}

			if currentBlock-receipt.BlockNumber.Uint64() >= confirmations {
				return receipt, nil
			}
		}
	}
}

// GetBlockNumber gets current block number
func (c *Web3Client) GetBlockNumber(ctx context.Context) (uint64, error) {
	blockNumber, err := c.client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get block number: %w", err)
	}
	return blockNumber, nil
}

// GetBlock gets block by number
func (c *Web3Client) GetBlock(ctx context.Context, blockNumber *big.Int) (*types.Block, error) {
	block, err := c.client.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}
	return block, nil
}

// Close closes the client connection
func (c *Web3Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		c.client.Close()
	}
	if c.wsClient != nil {
		c.wsClient.Close()
	}
}
