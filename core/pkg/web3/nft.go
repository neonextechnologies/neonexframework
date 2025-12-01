package web3

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

// NFTManager manages NFT operations
type NFTManager struct {
	client          *Web3Client
	contractManager *ContractManager
	mu              sync.RWMutex
}

// NFTMetadata NFT metadata structure
type NFTMetadata struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Image       string                 `json:"image"`
	ExternalURL string                 `json:"external_url,omitempty"`
	Attributes  []NFTAttribute         `json:"attributes,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// NFTAttribute NFT attribute
type NFTAttribute struct {
	TraitType   string      `json:"trait_type"`
	Value       interface{} `json:"value"`
	DisplayType string      `json:"display_type,omitempty"`
}

// NFT NFT details
type NFT struct {
	TokenID      *big.Int
	ContractAddr common.Address
	Owner        common.Address
	TokenURI     string
	Metadata     *NFTMetadata
}

// NewNFTManager creates a new NFT manager
func NewNFTManager(client *Web3Client, contractManager *ContractManager) *NFTManager {
	return &NFTManager{
		client:          client,
		contractManager: contractManager,
	}
}

// GetNFT gets NFT details
func (m *NFTManager) GetNFT(ctx context.Context, contractAddress common.Address, tokenID *big.Int) (*NFT, error) {
	// Get owner
	owner, err := m.contractManager.ERC721OwnerOf(ctx, contractAddress, tokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner: %w", err)
	}

	// Get token URI
	results, err := m.contractManager.CallMethod(ctx, contractAddress, "tokenURI", tokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokenURI: %w", err)
	}

	tokenURI := ""
	if len(results) > 0 {
		if uri, ok := results[0].(string); ok {
			tokenURI = uri
		}
	}

	nft := &NFT{
		TokenID:      tokenID,
		ContractAddr: contractAddress,
		Owner:        owner,
		TokenURI:     tokenURI,
	}

	return nft, nil
}

// GetNFTsByOwner gets all NFTs owned by an address
func (m *NFTManager) GetNFTsByOwner(ctx context.Context, contractAddress, owner common.Address) ([]*NFT, error) {
	// Get balance
	balance, err := m.contractManager.ERC721BalanceOf(ctx, contractAddress, owner)
	if err != nil {
		return nil, err
	}

	nfts := make([]*NFT, 0, balance.Int64())

	// This would iterate through tokens in real implementation
	// For now, return empty list
	return nfts, nil
}

// MintNFT mints a new NFT
func (m *NFTManager) MintNFT(ctx context.Context, wallet *Wallet, contractAddress, to common.Address, tokenID *big.Int, tokenURI string) (*Transaction, error) {
	return m.contractManager.SendMethod(ctx, wallet, contractAddress, "mint", big.NewInt(0), to, tokenID, tokenURI)
}

// TransferNFT transfers NFT
func (m *NFTManager) TransferNFT(ctx context.Context, wallet *Wallet, contractAddress common.Address, from, to common.Address, tokenID *big.Int) (*Transaction, error) {
	return m.contractManager.ERC721TransferFrom(ctx, wallet, contractAddress, from, to, tokenID)
}

// ApproveNFT approves NFT transfer
func (m *NFTManager) ApproveNFT(ctx context.Context, wallet *Wallet, contractAddress common.Address, to common.Address, tokenID *big.Int) (*Transaction, error) {
	return m.contractManager.SendMethod(ctx, wallet, contractAddress, "approve", big.NewInt(0), to, tokenID)
}

// SetApprovalForAll sets approval for all NFTs
func (m *NFTManager) SetApprovalForAll(ctx context.Context, wallet *Wallet, contractAddress common.Address, operator common.Address, approved bool) (*Transaction, error) {
	return m.contractManager.SendMethod(ctx, wallet, contractAddress, "setApprovalForAll", big.NewInt(0), operator, approved)
}

// TokenManager manages token operations
type TokenManager struct {
	client          *Web3Client
	contractManager *ContractManager
	mu              sync.RWMutex
}

// Token token details
type Token struct {
	Address     common.Address
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply *big.Int
}

// NewTokenManager creates a new token manager
func NewTokenManager(client *Web3Client, contractManager *ContractManager) *TokenManager {
	return &TokenManager{
		client:          client,
		contractManager: contractManager,
	}
}

// GetTokenInfo gets token information
func (m *TokenManager) GetTokenInfo(ctx context.Context, tokenAddress common.Address) (*Token, error) {
	token := &Token{
		Address: tokenAddress,
	}

	// Get name
	if results, err := m.contractManager.CallMethod(ctx, tokenAddress, "name"); err == nil && len(results) > 0 {
		if name, ok := results[0].(string); ok {
			token.Name = name
		}
	}

	// Get symbol
	if results, err := m.contractManager.CallMethod(ctx, tokenAddress, "symbol"); err == nil && len(results) > 0 {
		if symbol, ok := results[0].(string); ok {
			token.Symbol = symbol
		}
	}

	// Get decimals
	if results, err := m.contractManager.CallMethod(ctx, tokenAddress, "decimals"); err == nil && len(results) > 0 {
		if decimals, ok := results[0].(uint8); ok {
			token.Decimals = decimals
		}
	}

	// Get total supply
	if results, err := m.contractManager.CallMethod(ctx, tokenAddress, "totalSupply"); err == nil && len(results) > 0 {
		if supply, ok := results[0].(*big.Int); ok {
			token.TotalSupply = supply
		}
	}

	return token, nil
}

// GetBalance gets token balance
func (m *TokenManager) GetBalance(ctx context.Context, tokenAddress, account common.Address) (*big.Int, error) {
	return m.contractManager.ERC20BalanceOf(ctx, tokenAddress, account)
}

// Transfer transfers tokens
func (m *TokenManager) Transfer(ctx context.Context, wallet *Wallet, tokenAddress, to common.Address, amount *big.Int) (*Transaction, error) {
	return m.contractManager.ERC20Transfer(ctx, wallet, tokenAddress, to, amount)
}

// Approve approves token spending
func (m *TokenManager) Approve(ctx context.Context, wallet *Wallet, tokenAddress, spender common.Address, amount *big.Int) (*Transaction, error) {
	return m.contractManager.ERC20Approve(ctx, wallet, tokenAddress, spender, amount)
}

// GetAllowance gets token allowance
func (m *TokenManager) GetAllowance(ctx context.Context, tokenAddress, owner, spender common.Address) (*big.Int, error) {
	return m.contractManager.ERC20Allowance(ctx, tokenAddress, owner, spender)
}

// TransferFrom transfers tokens from another address
func (m *TokenManager) TransferFrom(ctx context.Context, wallet *Wallet, tokenAddress, from, to common.Address, amount *big.Int) (*Transaction, error) {
	return m.contractManager.SendMethod(ctx, wallet, tokenAddress, "transferFrom", big.NewInt(0), from, to, amount)
}

// GasEstimator estimates gas for transactions
type GasEstimator struct {
	client *Web3Client
}

// NewGasEstimator creates a new gas estimator
func NewGasEstimator(client *Web3Client) *GasEstimator {
	return &GasEstimator{
		client: client,
	}
}

// EstimateTransfer estimates gas for transfer
func (g *GasEstimator) EstimateTransfer(ctx context.Context) (uint64, error) {
	return 21000, nil // Standard transfer gas
}

// EstimateTokenTransfer estimates gas for token transfer
func (g *GasEstimator) EstimateTokenTransfer(ctx context.Context) (uint64, error) {
	return 65000, nil // Approximate gas for ERC20 transfer
}

// EstimateContractCall estimates gas for contract call
func (g *GasEstimator) EstimateContractCall(ctx context.Context, data []byte) (uint64, error) {
	baseGas := uint64(21000)
	dataGas := uint64(len(data)) * 68 // 68 gas per byte of data
	return baseGas + dataGas, nil
}

// GetGasPrice gets current gas price
func (g *GasEstimator) GetGasPrice(ctx context.Context) (*big.Int, error) {
	return g.client.SuggestGasPrice(ctx)
}

// CalculateTxCost calculates transaction cost
func (g *GasEstimator) CalculateTxCost(ctx context.Context, gasLimit uint64) (*big.Int, error) {
	gasPrice, err := g.GetGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	cost := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	return cost, nil
}
