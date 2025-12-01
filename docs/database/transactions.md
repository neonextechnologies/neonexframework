# Database Transactions

NeonEx Framework provides robust transaction management through the **TxManager** for handling database operations that require atomicity, consistency, isolation, and durability (ACID).

## Table of Contents

- [Overview](#overview)
- [Transaction Manager](#transaction-manager)
- [Basic Usage](#basic-usage)
- [Advanced Patterns](#advanced-patterns)
- [Nested Transactions](#nested-transactions)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)
- [Performance Optimization](#performance-optimization)
- [Troubleshooting](#troubleshooting)

## Overview

Transactions ensure that a series of database operations either all succeed or all fail together, maintaining data consistency.

### When to Use Transactions

✅ **Use transactions for:**
- Creating related records (user + profile)
- Financial operations (debit + credit)
- Multi-step business logic
- Data integrity requirements

❌ **Avoid transactions for:**
- Single read operations
- Independent operations
- Long-running processes

## Transaction Manager

### TxManager Structure

```go
// core/pkg/database/transaction.go
package database

import (
    "context"
    "gorm.io/gorm"
)

type TxManager struct {
    db *gorm.DB
}

func NewTxManager(db *gorm.DB) *TxManager {
    return &TxManager{db: db}
}
```

### Initialization

```go
import (
    "neonexcore/pkg/database"
    "gorm.io/gorm"
)

// Initialize in your service
func NewUserService(db *gorm.DB) *UserService {
    return &UserService{
        txManager: database.NewTxManager(db),
    }
}
```

## Basic Usage

### WithTransaction Method

The recommended way to use transactions in NeonEx:

```go
func (s *UserService) CreateUserWithProfile(ctx context.Context, userData *CreateUserRequest) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Create user
        user := &User{
            Name:  userData.Name,
            Email: userData.Email,
        }
        if err := tx.Create(user).Error; err != nil {
            return err // Transaction will rollback
        }
        
        // Create profile
        profile := &Profile{
            UserID: user.ID,
            Bio:    userData.Bio,
        }
        if err := tx.Create(profile).Error; err != nil {
            return err // Transaction will rollback
        }
        
        return nil // Transaction will commit
    })
}
```

### Manual Transaction Control

For more control, use manual Begin/Commit/Rollback:

```go
func (s *OrderService) ProcessOrder(ctx context.Context, orderData *OrderRequest) error {
    tx := s.txManager.BeginTx(ctx)
    
    // Create order
    order := &Order{
        UserID: orderData.UserID,
        Total:  orderData.Total,
    }
    if err := tx.Create(order).Error; err != nil {
        s.txManager.Rollback(tx)
        return err
    }
    
    // Create order items
    for _, item := range orderData.Items {
        orderItem := &OrderItem{
            OrderID:   order.ID,
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
        }
        if err := tx.Create(orderItem).Error; err != nil {
            s.txManager.Rollback(tx)
            return err
        }
    }
    
    // Update inventory
    if err := s.updateInventory(tx, orderData.Items); err != nil {
        s.txManager.Rollback(tx)
        return err
    }
    
    return s.txManager.Commit(tx)
}
```

### Using Repository WithTx

Repositories can work with transactions:

```go
func (s *UserService) TransferBalance(ctx context.Context, fromUserID, toUserID uint, amount float64) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Use repository with transaction
        userRepo := s.userRepo.WithTx(tx)
        
        // Get users
        fromUser, err := userRepo.FindByID(ctx, fromUserID)
        if err != nil {
            return err
        }
        
        toUser, err := userRepo.FindByID(ctx, toUserID)
        if err != nil {
            return err
        }
        
        // Check balance
        if fromUser.Balance < amount {
            return errors.NewBadRequest("Insufficient balance")
        }
        
        // Update balances
        fromUser.Balance -= amount
        toUser.Balance += amount
        
        if err := userRepo.Update(ctx, fromUser); err != nil {
            return err
        }
        
        if err := userRepo.Update(ctx, toUser); err != nil {
            return err
        }
        
        return nil
    })
}
```

## Advanced Patterns

### Transaction with Multiple Services

```go
type OrderService struct {
    txManager       *database.TxManager
    orderRepo       *OrderRepository
    inventoryRepo   *InventoryRepository
    paymentService  *PaymentService
    notificationSvc *notification.Manager
}

func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    var order *Order
    
    err := s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // 1. Create order
        order = &Order{
            UserID: req.UserID,
            Total:  req.Total,
            Status: "pending",
        }
        if err := tx.Create(order).Error; err != nil {
            return fmt.Errorf("failed to create order: %w", err)
        }
        
        // 2. Create order items and update inventory
        for _, item := range req.Items {
            orderItem := &OrderItem{
                OrderID:   order.ID,
                ProductID: item.ProductID,
                Quantity:  item.Quantity,
                Price:     item.Price,
            }
            if err := tx.Create(orderItem).Error; err != nil {
                return fmt.Errorf("failed to create order item: %w", err)
            }
            
            // Decrease inventory
            result := tx.Model(&Inventory{}).
                Where("product_id = ? AND quantity >= ?", item.ProductID, item.Quantity).
                UpdateColumn("quantity", gorm.Expr("quantity - ?", item.Quantity))
            
            if result.Error != nil {
                return fmt.Errorf("failed to update inventory: %w", result.Error)
            }
            if result.RowsAffected == 0 {
                return errors.NewBadRequest("Insufficient inventory")
            }
        }
        
        // 3. Process payment
        if err := s.paymentService.ChargeWithTx(tx, order.ID, req.Total); err != nil {
            return fmt.Errorf("payment failed: %w", err)
        }
        
        // 4. Update order status
        order.Status = "confirmed"
        if err := tx.Save(order).Error; err != nil {
            return fmt.Errorf("failed to update order: %w", err)
        }
        
        return nil
    })
    
    if err != nil {
        return nil, err
    }
    
    // Send notification outside transaction (async)
    go s.notificationSvc.SendEmail(ctx, order.UserID, "Order Confirmed", "Your order has been confirmed")
    
    return order, nil
}
```

### TransactionalOperation Helper

```go
// Using the TransactionalOperation helper
func (s *UserService) SetupNewUser(ctx context.Context, req *SetupUserRequest) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        op := database.NewTransactionalOperation(tx)
        
        return op.Execute(
            // Operation 1: Create user
            func(tx *gorm.DB) error {
                user := &User{
                    Name:  req.Name,
                    Email: req.Email,
                }
                return tx.Create(user).Error
            },
            
            // Operation 2: Assign role
            func(tx *gorm.DB) error {
                userRole := &UserRole{
                    UserID: req.UserID,
                    RoleID: req.RoleID,
                }
                return tx.Create(userRole).Error
            },
            
            // Operation 3: Create settings
            func(tx *gorm.DB) error {
                settings := &UserSettings{
                    UserID: req.UserID,
                    Theme:  "default",
                }
                return tx.Create(settings).Error
            },
        )
    })
}
```

## Nested Transactions

### Savepoints

GORM supports savepoints for nested transaction-like behavior:

```go
func (s *OrderService) ComplexOrder(ctx context.Context) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Main transaction work
        order := &Order{Total: 100}
        if err := tx.Create(order).Error; err != nil {
            return err
        }
        
        // Create savepoint
        tx = tx.SavePoint("sp1")
        
        // Try optional operation
        if err := s.tryOptionalOperation(tx); err != nil {
            // Rollback to savepoint (not entire transaction)
            tx.RollbackTo("sp1")
            // Continue with main transaction
        }
        
        // Continue main transaction
        return s.finalizeOrder(tx, order)
    })
}
```

### Nested WithTransaction

```go
func (s *UserService) CreateUserWithAudit(ctx context.Context, req *CreateUserRequest) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Create user
        user := &User{Name: req.Name, Email: req.Email}
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        
        // Nested transaction for audit (will use same tx in GORM)
        return s.createAuditLog(ctx, tx, user.ID, "user_created")
    })
}

func (s *UserService) createAuditLog(ctx context.Context, tx *gorm.DB, userID uint, action string) error {
    audit := &AuditLog{
        UserID: userID,
        Action: action,
        Time:   time.Now(),
    }
    return tx.Create(audit).Error
}
```

## Error Handling

### Proper Error Handling Pattern

```go
func (s *ProductService) UpdateProductWithHistory(ctx context.Context, productID uint, updates map[string]interface{}) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Find product
        var product Product
        if err := tx.First(&product, productID).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return errors.NewNotFound("Product not found")
            }
            return fmt.Errorf("database error: %w", err)
        }
        
        // Save history
        history := &ProductHistory{
            ProductID: productID,
            OldData:   product,
            Timestamp: time.Now(),
        }
        if err := tx.Create(history).Error; err != nil {
            return fmt.Errorf("failed to create history: %w", err)
        }
        
        // Update product
        if err := tx.Model(&product).Updates(updates).Error; err != nil {
            return fmt.Errorf("failed to update product: %w", err)
        }
        
        return nil
    })
}
```

### Custom Error Types

```go
var (
    ErrInsufficientFunds = errors.New("insufficient_funds", "Insufficient funds", 400)
    ErrInvalidTransfer   = errors.New("invalid_transfer", "Invalid transfer", 400)
)

func (s *WalletService) Transfer(ctx context.Context, fromID, toID uint, amount float64) error {
    err := s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        var fromWallet, toWallet Wallet
        
        if err := tx.First(&fromWallet, fromID).Error; err != nil {
            return err
        }
        
        if fromWallet.Balance < amount {
            return ErrInsufficientFunds
        }
        
        if err := tx.First(&toWallet, toID).Error; err != nil {
            return err
        }
        
        fromWallet.Balance -= amount
        toWallet.Balance += amount
        
        if err := tx.Save(&fromWallet).Error; err != nil {
            return err
        }
        
        return tx.Save(&toWallet).Error
    })
    
    if err != nil {
        // Log transaction failure
        log.Error("Transfer failed", logger.Fields{
            "from":   fromID,
            "to":     toID,
            "amount": amount,
            "error":  err.Error(),
        })
        return err
    }
    
    return nil
}
```

### Panic Recovery

```go
func (s *OrderService) ProcessOrderSafe(ctx context.Context, req *OrderRequest) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic in transaction: %v", r)
        }
    }()
    
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Transaction operations that might panic
        return s.processOrderInternal(tx, req)
    })
}
```

## Best Practices

### 1. Keep Transactions Short

```go
// ❌ Bad: Long-running transaction
func (s *UserService) BadExample(ctx context.Context) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        user := &User{}
        tx.Create(user)
        
        // External API call in transaction - BAD!
        time.Sleep(5 * time.Second)
        
        return nil
    })
}

// ✅ Good: Short transaction, external calls outside
func (s *UserService) GoodExample(ctx context.Context) error {
    // Prepare data first
    data := s.prepareData()
    
    // Quick transaction
    err := s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        return tx.Create(&User{Name: data.Name}).Error
    })
    
    if err != nil {
        return err
    }
    
    // External calls after transaction
    s.sendWelcomeEmail(data.Email)
    return nil
}
```

### 2. Use Context for Cancellation

```go
func (s *OrderService) ProcessOrder(ctx context.Context, orderID uint) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Check context cancellation
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        // Process order steps...
        
        return nil
    })
}
```

### 3. Avoid Nested Transactions When Possible

```go
// ✅ Good: Single transaction
func (s *UserService) CreateUserComplete(ctx context.Context, req *CreateUserRequest) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Create user
        user := &User{Name: req.Name}
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        
        // Create profile (same transaction)
        profile := &Profile{UserID: user.ID}
        return tx.Create(profile).Error
    })
}
```

### 4. Use Proper Isolation Levels

```go
// Set isolation level when needed
func (s *AccountService) ConcurrentUpdate(ctx context.Context) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Set isolation level
        tx = tx.Set("gorm:isolation_level", "SERIALIZABLE")
        
        // Critical operations
        return s.updateAccount(tx)
    })
}
```

### 5. Log Transaction Events

```go
func (s *OrderService) CreateOrderLogged(ctx context.Context, req *OrderRequest) error {
    log := logger.WithContext(ctx)
    log.Info("Starting order transaction", logger.Fields{"user_id": req.UserID})
    
    err := s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Transaction operations
        return s.processOrder(tx, req)
    })
    
    if err != nil {
        log.Error("Order transaction failed", logger.Fields{"error": err.Error()})
        return err
    }
    
    log.Info("Order transaction completed successfully")
    return nil
}
```

## Performance Optimization

### Batch Operations in Transactions

```go
func (s *ProductService) BulkImport(ctx context.Context, products []*Product) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Use batch insert instead of individual inserts
        return tx.CreateInBatches(products, 100).Error
    })
}
```

### Optimistic Locking

```go
type Account struct {
    gorm.Model
    Balance float64
    Version int `gorm:"default:0"` // Version field for optimistic locking
}

func (s *AccountService) UpdateBalance(ctx context.Context, accountID uint, amount float64) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        var account Account
        if err := tx.First(&account, accountID).Error; err != nil {
            return err
        }
        
        oldVersion := account.Version
        account.Balance += amount
        account.Version++
        
        // Update with version check
        result := tx.Model(&account).
            Where("id = ? AND version = ?", accountID, oldVersion).
            Updates(map[string]interface{}{
                "balance": account.Balance,
                "version": account.Version,
            })
        
        if result.RowsAffected == 0 {
            return errors.NewConflict("Account was modified by another transaction")
        }
        
        return result.Error
    })
}
```

### Read-Only Transactions

```go
// For read-only operations, use regular queries (no transaction needed)
func (s *ReportService) GenerateReport(ctx context.Context) (*Report, error) {
    // No transaction needed for reads
    var data []ReportData
    if err := s.db.Find(&data).Error; err != nil {
        return nil, err
    }
    
    return s.processReport(data), nil
}
```

## Troubleshooting

### Common Issues

**Issue: Deadlock detected**

```go
// Solution: Always acquire locks in the same order
func (s *TransferService) Transfer(ctx context.Context, fromID, toID uint, amount float64) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Always lock accounts in order (smaller ID first)
        id1, id2 := fromID, toID
        if id1 > id2 {
            id1, id2 = id2, id1
        }
        
        var account1, account2 Account
        tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&account1, id1)
        tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&account2, id2)
        
        // Perform transfer
        return s.doTransfer(tx, fromID, toID, amount)
    })
}
```

**Issue: Transaction timeout**

```go
import "time"

func (s *OrderService) ProcessWithTimeout(ctx context.Context) error {
    // Set transaction timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        // Operations must complete within 5 seconds
        return s.quickOperation(tx)
    })
}
```

**Issue: Lost updates**

```go
// Use pessimistic locking for critical sections
func (s *InventoryService) ReserveStock(ctx context.Context, productID uint, quantity int) error {
    return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
        var inventory Inventory
        
        // Lock row for update
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&inventory, "product_id = ?", productID).Error; err != nil {
            return err
        }
        
        if inventory.Quantity < quantity {
            return errors.NewBadRequest("Insufficient stock")
        }
        
        inventory.Quantity -= quantity
        return tx.Save(&inventory).Error
    })
}
```

## Summary

NeonEx Framework's transaction management provides:

✅ **Simple API** with `WithTransaction` method  
✅ **ACID guarantees** for data consistency  
✅ **Automatic rollback** on errors  
✅ **Context support** for cancellation  
✅ **Repository integration** with `WithTx`  
✅ **Production-ready** error handling

For more information:
- [Database Configuration](configuration.md)
- [Repository Pattern](repository.md)
- [Models](models.md)
