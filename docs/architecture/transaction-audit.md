# AAMS Backend — Transaction Audit & UnitOfWork Architecture

This document audits all database transactions across the current codebase and establishes the official **UnitOfWork & TransactionSession Architecture** for future phases.

---

## 1. Current Codebase Transaction Audit

| File | Function / Location | Tables Touched | Current Transaction Mechanism | Cross-Domain Risk |
|---|---|---|---|---|
| `internal/service/services.go` | `WorkService.StartWork` | `work_sessions`, `employees`, `vehicles` | Sequential GORM calls (`db.Create`, `db.Model.Update`) without explicit `db.Transaction` block | **High:** Partial failure during start work leaves session half-created. |
| `internal/service/services.go` | `WorkService.EndWork` | `work_sessions`, `vehicles` | `r.db.Transaction(...)` inside repository layer | **Medium:** Transaction is hidden inside monolithic repository method. |
| `internal/service/services.go` | `VehicleService.RecordOilChange` | `oil_changes`, `vehicles`, `inventory_items` | Direct call to `r.db.Create(&oilChange)` followed by `r.db.Model(&vehicle).Update` | **Critical:** If vehicle update fails, oil change record remains created without inventory deduction atomicity. |
| `internal/service/services.go` | `InventoryService.DispenseOil` | `inventory_items`, `stock_transactions` | Implicit non-transactional sequential execution | **Medium:** Stock count may decrement without transaction log on crash. |
| `internal/service/services.go` | `CustodyService.AddExpense` | `custody_expenses`, `custody_days` | Updates expense and calculates custody day total | **Medium:** Concurrency risk if multiple expenses posted simultaneously. |
| `internal/service/target_service.go` | `TargetService.ConfirmExcelImport` | `targets`, `target_logs`, `platform_mappings` | `tx := s.repo.Begin()` with explicit loop `tx.Commit()` | **High:** Long transaction holding locks during heavy Excel row iterations. |

---

## 2. Identified Anti-Patterns in Current Code

1. **Transaction Leakage in Repositories:** Some transactions are initiated inside `gorm_repo.go` while others are initiated inside `target_service.go`.
2. **Missing Atomicity on Cross-Domain Mutations:** `RecordOilChange` touches 3 domains (`maintenance`, `fleet`, `inventory`) without a single enclosing atomic transaction.
3. **Implicit ORM Coupling:** Business services directly access `*gorm.DB` or `repository.DB` handles.

---

## 3. The Approved Future Model: `UnitOfWork & TransactionSession`

```text
Primary Business Service (Transaction Coordinator & Owner)
               │
               ▼
   UnitOfWork.Do(ctx, func(tx database.TransactionSession) error {
               │
               ├── 1. PrimaryModuleRepo.CreateRecord(tx, entity)
               │
               ├── 2. ForeignModuleContractA.ExecuteDeduction(tx, req)
               │
               └── 3. ForeignModuleContractB.UpdateOdometer(tx, req)
   })
               │
               ▼
   UnitOfWork Executes Commit() OR Rollback() Atomically
```

### Architectural Rules:
1. **Service is Coordinator:** The primary service that accepts the command owns the transaction lifetime.
2. **UnitOfWork Handles Lifecycle:** `UnitOfWork` is the exclusive owner of `tx.Commit()` and `tx.Rollback()`.
3. **Type-Safe Abstraction:** `database.TransactionSession` exposes query execution capabilities without exposing `*gorm.DB` struct internals to domain contracts.
4. **No Nested Autonomous Transactions:** Participating modules MUST use the passed `TransactionSession` and are strictly forbidden from calling `Begin()` or `Commit()` independently.
