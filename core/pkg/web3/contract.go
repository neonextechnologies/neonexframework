package web3

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ContractManager manages smart contract interactions
type ContractManager struct {
	client    *Web3Client
	contracts map[common.Address]*Contract
	mu        sync.RWMutex
}

// Contract smart contract wrapper
type Contract struct {
	Address  common.Address
	ABI      abi.ABI
	Instance *bind.BoundContract
}

// ContractEvent contract event
type ContractEvent struct {
	Name        string
	Address     common.Address
	BlockNumber uint64
	TxHash      common.Hash
	Data        map[string]interface{}
}

// NewContractManager creates a new contract manager
func NewContractManager(client *Web3Client) *ContractManager {
	return &ContractManager{
		client:    client,
		contracts: make(map[common.Address]*Contract),
	}
}

// LoadContract loads a contract
func (m *ContractManager) LoadContract(address common.Address, abiJSON string) (*Contract, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Parse ABI
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Create bound contract
	instance := bind.NewBoundContract(address, parsedABI, m.client.client, m.client.client, m.client.client)

	contract := &Contract{
		Address:  address,
		ABI:      parsedABI,
		Instance: instance,
	}

	m.contracts[address] = contract
	return contract, nil
}

// GetContract gets a loaded contract
func (m *ContractManager) GetContract(address common.Address) (*Contract, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contract, exists := m.contracts[address]
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", address.Hex())
	}

	return contract, nil
}

// CallMethod calls a contract read method
func (m *ContractManager) CallMethod(ctx context.Context, contractAddress common.Address, methodName string, args ...interface{}) ([]interface{}, error) {
	contract, err := m.GetContract(contractAddress)
	if err != nil {
		return nil, err
	}

	// Pack method call
	data, err := contract.ABI.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack method: %w", err)
	}

	// Call contract
	msg := map[string]interface{}{
		"to":   contractAddress,
		"data": data,
	}

	// This would use ethereum.CallMsg in real implementation
	_ = msg

	// Unpack results
	method, exists := contract.ABI.Methods[methodName]
	if !exists {
		return nil, fmt.Errorf("method not found: %s", methodName)
	}

	// For this example, return empty results
	results := make([]interface{}, len(method.Outputs))
	return results, nil
}

// SendMethod sends a contract write transaction
func (m *ContractManager) SendMethod(ctx context.Context, wallet *Wallet, contractAddress common.Address, methodName string, value *big.Int, args ...interface{}) (*Transaction, error) {
	contract, err := m.GetContract(contractAddress)
	if err != nil {
		return nil, err
	}

	// Pack method call
	data, err := contract.ABI.Pack(methodName, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack method: %w", err)
	}

	// Send transaction
	return m.client.SendTransaction(ctx, wallet, contractAddress, value, data)
}

// DeployContract deploys a new contract
func (m *ContractManager) DeployContract(ctx context.Context, wallet *Wallet, abiJSON string, bytecode []byte, args ...interface{}) (*Transaction, common.Address, error) {
	// Parse ABI
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Pack constructor arguments
	var data []byte
	if len(args) > 0 {
		constructorArgs, err := parsedABI.Pack("", args...)
		if err != nil {
			return nil, common.Address{}, fmt.Errorf("failed to pack constructor args: %w", err)
		}
		data = append(bytecode, constructorArgs...)
	} else {
		data = bytecode
	}

	// Get nonce
	nonce, err := m.client.GetNonce(ctx, wallet.Address)
	if err != nil {
		return nil, common.Address{}, err
	}

	// Get gas price
	gasPrice, err := m.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, common.Address{}, err
	}

	// Create contract creation transaction
	tx := types.NewContractCreation(nonce, big.NewInt(0), 3000000, gasPrice, data)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(m.client.chainID), wallet.PrivateKey)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = m.client.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Calculate contract address
	contractAddress := crypto.CreateAddress(wallet.Address, nonce)

	transaction := &Transaction{
		Hash:      signedTx.Hash(),
		From:      wallet.Address,
		To:        nil, // Contract creation
		Value:     big.NewInt(0),
		Gas:       3000000,
		GasPrice:  gasPrice,
		Nonce:     nonce,
		Data:      data,
		Status:    TxStatusPending,
	}

	return transaction, contractAddress, nil
}

// WatchEvents watches for contract events
func (m *ContractManager) WatchEvents(ctx context.Context, contractAddress common.Address, eventName string, fromBlock uint64) (<-chan *ContractEvent, error) {
	contract, err := m.GetContract(contractAddress)
	if err != nil {
		return nil, err
	}

	// Get event from ABI
	event, exists := contract.ABI.Events[eventName]
	if !exists {
		return nil, fmt.Errorf("event not found: %s", eventName)
	}

	_ = event // Use event for filtering

	// Create event channel
	eventChan := make(chan *ContractEvent, 100)

	// Start watching (simplified)
	go func() {
		defer close(eventChan)
		<-ctx.Done()
	}()

	return eventChan, nil
}

// GetPastEvents gets past contract events
func (m *ContractManager) GetPastEvents(ctx context.Context, contractAddress common.Address, eventName string, fromBlock, toBlock uint64) ([]*ContractEvent, error) {
	contract, err := m.GetContract(contractAddress)
	if err != nil {
		return nil, err
	}

	// Get event from ABI
	event, exists := contract.ABI.Events[eventName]
	if !exists {
		return nil, fmt.Errorf("event not found: %s", eventName)
	}

	_ = event // Use event for filtering

	// This would use FilterLogs in real implementation
	events := make([]*ContractEvent, 0)

	return events, nil
}

// ERC20 standard interface helpers

// ERC20Transfer transfers ERC20 tokens
func (m *ContractManager) ERC20Transfer(ctx context.Context, wallet *Wallet, tokenAddress common.Address, to common.Address, amount *big.Int) (*Transaction, error) {
	return m.SendMethod(ctx, wallet, tokenAddress, "transfer", big.NewInt(0), to, amount)
}

// ERC20BalanceOf gets ERC20 token balance
func (m *ContractManager) ERC20BalanceOf(ctx context.Context, tokenAddress common.Address, account common.Address) (*big.Int, error) {
	results, err := m.CallMethod(ctx, tokenAddress, "balanceOf", account)
	if err != nil {
		return nil, err
	}

	if len(results) > 0 {
		if balance, ok := results[0].(*big.Int); ok {
			return balance, nil
		}
	}

	return big.NewInt(0), nil
}

// ERC20Approve approves spending
func (m *ContractManager) ERC20Approve(ctx context.Context, wallet *Wallet, tokenAddress common.Address, spender common.Address, amount *big.Int) (*Transaction, error) {
	return m.SendMethod(ctx, wallet, tokenAddress, "approve", big.NewInt(0), spender, amount)
}

// ERC20Allowance gets allowance
func (m *ContractManager) ERC20Allowance(ctx context.Context, tokenAddress common.Address, owner, spender common.Address) (*big.Int, error) {
	results, err := m.CallMethod(ctx, tokenAddress, "allowance", owner, spender)
	if err != nil {
		return nil, err
	}

	if len(results) > 0 {
		if allowance, ok := results[0].(*big.Int); ok {
			return allowance, nil
		}
	}

	return big.NewInt(0), nil
}

// ERC721 standard interface helpers

// ERC721OwnerOf gets NFT owner
func (m *ContractManager) ERC721OwnerOf(ctx context.Context, nftAddress common.Address, tokenID *big.Int) (common.Address, error) {
	results, err := m.CallMethod(ctx, nftAddress, "ownerOf", tokenID)
	if err != nil {
		return common.Address{}, err
	}

	if len(results) > 0 {
		if owner, ok := results[0].(common.Address); ok {
			return owner, nil
		}
	}

	return common.Address{}, nil
}

// ERC721TransferFrom transfers NFT
func (m *ContractManager) ERC721TransferFrom(ctx context.Context, wallet *Wallet, nftAddress common.Address, from, to common.Address, tokenID *big.Int) (*Transaction, error) {
	return m.SendMethod(ctx, wallet, nftAddress, "transferFrom", big.NewInt(0), from, to, tokenID)
}

// ERC721BalanceOf gets NFT balance
func (m *ContractManager) ERC721BalanceOf(ctx context.Context, nftAddress common.Address, owner common.Address) (*big.Int, error) {
	results, err := m.CallMethod(ctx, nftAddress, "balanceOf", owner)
	if err != nil {
		return nil, err
	}

	if len(results) > 0 {
		if balance, ok := results[0].(*big.Int); ok {
			return balance, nil
		}
	}

	return big.NewInt(0), nil
}
