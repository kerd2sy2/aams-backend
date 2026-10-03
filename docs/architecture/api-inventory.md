# AAMS Backend — Complete API Inventory

This document represents the frozen API compatibility baseline established in **Phase 0**.  
**Rules:** Zero route renaming, zero method changes, zero payload restructuring during refactoring.

---

## 1. Authentication & System Public Routes (`auth` & `system`)

| HTTP Method | Route URL | Current Handler | Service / Function | Auth / Role | Target Module |
|---|---|---|---|---|---|
| `GET` | `/` | Root inline handler | Static status check | Public | `system` |
| `GET` | `/api/v1` | Root inline handler | Version info | Public | `system` |
| `GET` | `/api/v1/health` | Health inline handler | App health status | Public | `system` |
| `GET` | `/api/v1/system/performance` | Metrics inline handler | Runtime MemStats & DB stats | Public (Admin Dashboard) | `system` |
| `POST` | `/api/v1/login` | `authHandler.Login` | `AuthService.Login` | Public (Rate Limited) | `auth` |
| `POST` | `/api/v1/auth/login` | `authHandler.Login` | `AuthService.Login` | Public (Rate Limited) | `auth` |
| `POST` | `/api/v1/refresh` | `authHandler.RefreshToken` | `AuthService.RefreshToken` | Public | `auth` |
| `POST` | `/api/v1/auth/refresh` | `authHandler.RefreshToken` | `AuthService.RefreshToken` | Public | `auth` |
| `POST` | `/api/refresh` | `authHandler.RefreshToken` | `AuthService.RefreshToken` | Public (Fallback) | `auth` |
| `POST` | `/auth/refresh` | `authHandler.RefreshToken` | `AuthService.RefreshToken` | Public (Fallback) | `auth` |
| `POST` | `/api/v1/auth/google/login` | `authHandler.GoogleLogin` | `AuthService.GoogleLogin` | Public | `auth` |
| `POST` | `/api/v1/auth/request-otp` | `otpHandler.RequestOTP` | `OTPService.RequestOTP` | Public | `auth` |
| `POST` | `/api/v1/auth/verify-otp` | `otpHandler.VerifyOTP` | `OTPService.VerifyOTP` | Public | `auth` |
| `GET` | `/api/v1/settings/public` | `settingHandler.GetPublicSettings` | `SettingService.GetPublicSettings` | Public | `system` |
| `GET` | `/api/v1/public/doc/:id` | `investigationHandler.GetPublicByID`| `InvestigationService.GetPublicDoc` | Public (QR Scanner) | `hr_legal` |
| `GET` | `/api/v1/public/investigations/:id`| `investigationHandler.GetPublicByID`| `InvestigationService.GetPublicDoc` | Public (QR Scanner) | `hr_legal` |
| `GET` | `/api/v1/me` | `authHandler.Me` | `AuthService.GetMe` | JWT Required | `auth` |
| `POST` | `/api/v1/auth/google/link` | `authHandler.LinkGoogle` | `AuthService.LinkGoogle` | JWT Required | `auth` |
| `POST` | `/api/v1/auth/google/unlink` | `authHandler.UnlinkGoogle` | `AuthService.UnlinkGoogle` | JWT Required | `auth` |

---

## 2. Employees & Delegates Management (`employee`)

| HTTP Method | Route URL | Current Handler | Service / Function | Target Module |
|---|---|---|---|---|
| `POST` | `/api/v1/employees` | `empHandler.Create` | `EmployeeService.Create` | `employee` |
| `GET` | `/api/v1/employees` | `empHandler.GetAll` | `EmployeeService.GetAll` | `employee` |
| `GET` | `/api/v1/employees/search` | `empHandler.Search` | `EmployeeService.Search` | `employee` |
| `GET` | `/api/v1/employees/working` | `empHandler.GetWorking` | `EmployeeService.GetWorking` | `employee` |
| `GET` | `/api/v1/employees/:id` | `empHandler.GetByID` | `EmployeeService.GetByID` | `employee` |
| `PUT` | `/api/v1/employees/:id` | `empHandler.Update` | `EmployeeService.Update` | `employee` |
| `DELETE` | `/api/v1/employees/:id` | `empHandler.Delete` | `EmployeeService.Delete` | `employee` |
| `POST` | `/api/v1/employees/batch-oil-setup` | `empHandler.BatchSetOilChange`| `EmployeeService.BatchSetOil` | `employee` |
| `POST` | `/api/v1/employees/me/change-password` | `empHandler.ChangeMyPassword` | `EmployeeService.ChangeMyPassword` | `employee` |
| `POST` | `/api/v1/employees/me/phone` | `empHandler.SetMyPhone` | `EmployeeService.SetMyPhone` | `employee` |
| `POST` | `/api/v1/employees/me/location` | `empHandler.SetMyLocation` | `EmployeeService.SetMyLocation` | `employee` |
| `GET` | `/api/v1/employees/locations` | `empHandler.GetLocations` | `EmployeeService.GetLocations` | `employee` |
| `PUT` | `/api/v1/employees/:id/phone` | `empHandler.SetPhone` | `EmployeeService.SetPhone` | `employee` |
| `POST` | `/api/v1/employees/:id/reset-password` | `empHandler.ResetPassword` | `EmployeeService.ResetPassword` | `employee` |
| `GET` | `/api/v1/employees/:id/barcode` | `empHandler.GetBarcode` | `EmployeeService.GetBarcode` | `employee` |
| `GET` | `/api/v1/employees/:id/qrcode` | `empHandler.GetQRCode` | `EmployeeService.GetQRCode` | `employee` |
| `GET` | `/api/v1/employees/:id/print-card` | `empHandler.GetPrintCard` | `EmployeeService.GetPrintCard` | `employee` |
| `POST` | `/api/v1/upload` | `empHandler.UploadImage` | `StorageService.UploadImage` | `system` (Storage) |
| `POST` | `/api/v1/upload-file` | `empHandler.UploadFile` | `StorageService.UploadFile` | `system` (Storage) |
| `GET` | `/api/v1/documents` | `docHandler.GetAll` | `DocumentService.GetAll` | `employee` |
| `GET` | `/api/v1/documents/expiring` | `docHandler.GetExpiringSoon`| `DocumentService.GetExpiring` | `employee` |
| `GET` | `/api/v1/documents/:id` | `docHandler.GetByID` | `DocumentService.GetByID` | `employee` |
| `POST` | `/api/v1/documents` | `docHandler.Create` | `DocumentService.Create` | `employee` |
| `PUT` | `/api/v1/documents/:id` | `docHandler.Update` | `DocumentService.Update` | `employee` |
| `DELETE` | `/api/v1/documents/:id` | `docHandler.Delete` | `DocumentService.Delete` | `employee` |
| `GET` | `/api/v1/bank-accounts` | `bankHandler.GetAll` | `BankAccountService.GetAll` | `employee` |
| `POST` | `/api/v1/bank-accounts` | `bankHandler.Create` | `BankAccountService.Create` | `employee` |
| `PUT` | `/api/v1/bank-accounts/:id` | `bankHandler.Update` | `BankAccountService.Update` | `employee` |
| `DELETE` | `/api/v1/bank-accounts/:id` | `bankHandler.Delete` | `BankAccountService.Delete` | `employee` |

---

## 3. Work Sessions & Shifts (`work`)

| HTTP Method | Route URL | Current Handler | Service / Function | Target Module |
|---|---|---|---|---|
| `POST` | `/api/v1/work/start` | `workHandler.StartWork` | `WorkService.StartWork` | `work` |
| `POST` | `/api/v1/work/end` | `workHandler.EndWork` | `WorkService.EndWork` | `work` |
| `PUT` | `/api/v1/work/:id` | `workHandler.UpdateWorkSession` | `WorkService.UpdateSession` | `work` |
| `PUT` | `/api/v1/work/:id/review` | `workHandler.ReviewWorkSession` | `WorkService.ReviewSession` | `work` |
| `PUT` | `/api/v1/work/sessions/:id/review` | `workHandler.ReviewWorkSession` | `WorkService.ReviewSession` | `work` |
| `GET` | `/api/v1/work/sessions/:id` | `workHandler.GetSessionByID` | `WorkService.GetSessionByID` | `work` |
| `GET` | `/api/v1/work/active` | `workHandler.GetActiveSession` | `WorkService.GetActiveSession` | `work` |
| `GET` | `/api/v1/work/last-km` | `workHandler.GetLastKM` | `WorkService.GetLastKM` | `work` |
| `POST` | `/api/v1/work/scan-plate` | `workHandler.ScanPlate` | `WorkService.ScanPlate (OCR)` | `work` |
| `GET` | `/api/v1/work/today-count` | `workHandler.TodayCount` | `WorkService.TodayCount` | `work` |
| `GET` | `/api/v1/work/check-oil` | `workHandler.CheckOilChange` | `WorkService.CheckOilChange` | `work` |

---

## 4. Fleet & Vehicles (`fleet`)

| HTTP Method | Route URL | Current Handler | Service / Function | Target Module |
|---|---|---|---|---|
| `GET` | `/api/v1/vehicles` | `vehicleHandler.GetAll` | `VehicleService.GetAll` | `fleet` |
| `POST` | `/api/v1/vehicles` | `vehicleHandler.Create` | `VehicleService.Create` | `fleet` |
| `GET` | `/api/v1/vehicles/check-km` | `vehicleHandler.CheckKM` | `VehicleService.CheckKM` | `fleet` |
| `GET` | `/api/v1/vehicles/:id` | `vehicleHandler.GetByID` | `VehicleService.GetByID` | `fleet` |
| `PUT` | `/api/v1/vehicles/:id` | `vehicleHandler.Update` | `VehicleService.Update` | `fleet` |
| `DELETE` | `/api/v1/vehicles/:id` | `vehicleHandler.Delete` | `VehicleService.Delete` | `fleet` |
| `POST` | `/api/v1/vehicles/:id/oil-change`| `vehicleHandler.RecordOilChange`| `VehicleService.RecordOilChange` | `maintenance` |

---

## 5. Maintenance & Repairs (`maintenance`)

| HTTP Method | Route URL | Current Handler | Service / Function | Target Module |
|---|---|---|---|---|
| `GET` | `/api/v1/maintenance/logs` | `maintHandler.GetAllLogs` | `MaintenanceService.GetAllLogs` | `maintenance` |
| `GET` | `/api/v1/maintenance/employee-logs` | `maintHandler.GetEmployeeLogs` | `MaintenanceService.GetEmployeeLogs` | `maintenance` |
| `GET` | `/api/v1/maintenance-requests` | `maintRequestHandler.GetAll` | `MaintenanceRequestService.GetAll` | `maintenance` |
| `POST` | `/api/v1/maintenance-requests` | `maintRequestHandler.Create` | `MaintenanceRequestService.Create` | `maintenance` |
| `PUT` | `/api/v1/maintenance-requests/:id` | `maintRequestHandler.Update` | `MaintenanceRequestService.Update` | `maintenance` |
| `DELETE` | `/api/v1/maintenance-requests/:id` | `maintRequestHandler.Delete` | `MaintenanceRequestService.Delete` | `maintenance` |

---

## 6. Inventory & Warehouse (`inventory`)

| HTTP Method | Route URL | Current Handler | Service / Function | Target Module |
|---|---|---|---|---|
| `GET` | `/api/v1/inventory/items` | `invHandler.GetItems` | `InventoryService.GetItems` | `inventory` |
| `GET` | `/api/v1/inventory/items/:id` | `invHandler.GetItemByID` | `InventoryService.GetItemByID` | `inventory` |
| `GET` | `/api/v1/inventory/barcode` | `invHandler.FindByBarcode` | `InventoryService.FindByBarcode` | `inventory` |
| `POST` | `/api/v1/inventory/items` | `invHandler.CreateItem` | `InventoryService.CreateItem` | `inventory` |
| `PUT` | `/api/v1/inventory/items/:id` | `invHandler.UpdateItem` | `InventoryService.UpdateItem` | `inventory` |
| `DELETE` | `/api/v1/inventory/items/:id` | `invHandler.DeleteItem` | `InventoryService.DeleteItem` | `inventory` |
| `POST` | `/api/v1/inventory/add-stock` | `invHandler.AddStock` | `InventoryService.AddStock` | `inventory` |
| `POST` | `/api/v1/inventory/remove-stock` | `invHandler.RemoveStock` | `InventoryService.RemoveStock` | `inventory` |
| `POST` | `/api/v1/inventory/dispense-oil` | `invHandler.DispenseOil` | `InventoryService.DispenseOil` | `inventory` |
| `GET` | `/api/v1/inventory/transactions` | `invHandler.GetTransactions` | `InventoryService.GetTransactions` | `inventory` |
| `DELETE` | `/api/v1/inventory/transactions` | `invHandler.DeleteAllTransactions`| `InventoryService.DeleteTransactions` | `inventory` |
| `GET` | `/api/v1/inventory/purchases` | `invHandler.GetPurchaseInvoices` | `InventoryService.GetPurchases` | `inventory` |
| `GET` | `/api/v1/inventory/purchases/:id` | `invHandler.GetPurchaseInvoiceByID`| `InventoryService.GetPurchaseByID` | `inventory` |
| `POST` | `/api/v1/inventory/purchases` | `invHandler.CreatePurchaseInvoice`| `InventoryService.CreatePurchase` | `inventory` |
| `DELETE` | `/api/v1/inventory/purchases/:id` | `invHandler.DeletePurchaseInvoice`| `InventoryService.DeletePurchase` | `inventory` |

---

## 7. Delivery Platform Targets & Excel Imports (`target`)

| HTTP Method | Route URL | Current Handler | Service / Function | Target Module |
|---|---|---|---|---|
| `GET` | `/api/v1/target/dashboard` | `targetHandler.GetDashboardSummary` | `TargetService.GetDashboard` | `target` |
| `GET` | `/api/v1/target/identifiers` | `targetHandler.ListIdentifiers` | `TargetService.ListIdentifiers` | `target` |
| `GET` | `/api/v1/target/identifiers/:id` | `targetHandler.GetIdentifierDetails`| `TargetService.GetIdentifier` | `target` |
| `GET` | `/api/v1/target/drivers` | `targetHandler.ListDrivers` | `TargetService.ListDrivers` | `target` |
| `GET` | `/api/v1/target/alerts` | `targetHandler.ListAlerts` | `TargetService.ListAlerts` | `target` |
| `PATCH` | `/api/v1/target/alerts/resolve-all` | `targetHandler.ResolveAllAlerts` | `TargetService.ResolveAllAlerts` | `target` |
| `PATCH` | `/api/v1/target/alerts/:id/resolve` | `targetHandler.ResolveAlert` | `TargetService.ResolveAlert` | `target` |
| `GET` | `/api/v1/target/settings` | `targetHandler.GetTargetSettings` | `TargetService.GetSettings` | `target` |
| `GET` | `/api/v1/target/batches` | `targetHandler.ListImportBatches` | `TargetService.ListBatches` | `target` |
| `POST` | `/api/v1/target/import/preview` | `targetHandler.PreviewExcelImport` | `TargetService.PreviewExcel` | `target` |
| `POST` | `/api/v1/target/import/confirm` | `targetHandler.ConfirmExcelImport` | `TargetService.ConfirmExcel` | `target` |
| `DELETE` | `/api/v1/target/batches/:id` | `targetHandler.DeleteImportBatch` | `TargetService.DeleteBatch` | `target` |
| `DELETE` | `/api/v1/target/batches/date/:orderDate` | `targetHandler.DeleteSheetByDate` | `TargetService.DeleteByDate` | `target` |
| `POST` | `/api/v1/target/identifiers` | `targetHandler.CreateIdentifier` | `TargetService.CreateIdentifier` | `target` |
| `PUT` | `/api/v1/target/identifiers/:id` | `targetHandler.UpdateIdentifier` | `TargetService.UpdateIdentifier` | `target` |
| `DELETE` | `/api/v1/target/identifiers` | `targetHandler.DeleteAllIdentifiers`| `TargetService.DeleteIdentifiers` | `target` |
| `DELETE` | `/api/v1/target/identifiers/wipe-all` | `targetHandler.DeleteAllIdentifiers`| `TargetService.DeleteIdentifiers` | `target` |
| `DELETE` | `/api/v1/target/identifiers/:id` | `targetHandler.DeleteIdentifier` | `TargetService.DeleteIdentifier` | `target` |
| `PUT` | `/api/v1/target/settings` | `targetHandler.UpdateTargetSettings`| `TargetService.UpdateSettings` | `target` |
| `POST` | `/api/v1/admin/target/import/preview` | `targetHandler.PreviewExcelImport` | `TargetService.PreviewExcel` | `target` |
| `POST` | `/api/v1/admin/target/import/confirm` | `targetHandler.ConfirmExcelImport` | `TargetService.ConfirmExcel` | `target` |
| `GET` | `/api/v1/admin/target/import/batches` | `targetHandler.ListImportBatches` | `TargetService.ListBatches` | `target` |
| `DELETE` | `/api/v1/admin/target/import/batches/:id` | `targetHandler.DeleteImportBatch` | `TargetService.DeleteBatch` | `target` |
| `DELETE` | `/api/v1/admin/target/import/batches/date/:orderDate` | `targetHandler.DeleteSheetByDate` | `TargetService.DeleteByDate` | `target` |

---

## 8. Branch Custody & Fuel Logs (`custody`)

| HTTP Method | Route URL | Current Handler | Service / Function | Target Module |
|---|---|---|---|---|
| `GET` | `/api/v1/custody` | `custodyHandler.List` | `CustodyService.List` | `custody` |
| `POST` | `/api/v1/custody` | `custodyHandler.Create` | `CustodyService.Create` | `custody` |
| `POST` | `/api/v1/custody/add-amount` | `custodyHandler.AddAmount` | `CustodyService.AddAmount` | `custody` |
| `GET` | `/api/v1/custody/logs` | `custodyHandler.GetLogs` | `CustodyService.GetLogs` | `custody` |
| `DELETE` | `/api/v1/custody/logs/:id` | `custodyHandler.DeleteLog` | `CustodyService.DeleteLog` | `custody` |
| `POST` | `/api/v1/custody/:id/expenses` | `custodyHandler.AddExpense` | `CustodyService.AddExpense` | `custody` |
| `DELETE` | `/api/v1/custody/expenses/:id` | `custodyHandler.DeleteExpense` | `CustodyService.DeleteExpense` | `custody` |
| `GET` | `/api/v1/fuel-logs` | `fuelLogHandler.GetAll` | `FuelLogService.GetAll` | `custody` |
| `POST` | `/api/v1/fuel-logs` | `fuelLogHandler.Create` | `FuelLogService.Create` | `custody` |
| `PUT` | `/api/v1/fuel-logs/:id` | `fuelLogHandler.Update` | `FuelLogService.Update` | `custody` |
| `DELETE` | `/api/v1/fuel-logs/:id` | `fuelLogHandler.Delete` | `FuelLogService.Delete` | `custody` |

---

## 9. HR, Legal, Attendance, Violations & Leaves (`hr_legal`)

| HTTP Method | Route URL | Current Handler | Service / Function | Target Module |
|---|---|---|---|---|
| `POST` | `/api/v1/investigations` | `investigationHandler.Create` | `InvestigationService.Create` | `hr_legal` |
| `GET` | `/api/v1/investigations` | `investigationHandler.GetAll` | `InvestigationService.GetAll` | `hr_legal` |
| `GET` | `/api/v1/investigations/pending-count` | `investigationHandler.PendingCount` | `InvestigationService.PendingCount`| `hr_legal` |
| `GET` | `/api/v1/investigations/:id` | `investigationHandler.GetByID` | `InvestigationService.GetByID` | `hr_legal` |
| `PUT` | `/api/v1/investigations/:id` | `investigationHandler.Update` | `InvestigationService.Update` | `hr_legal` |
| `POST` | `/api/v1/investigations/:id/approve` | `investigationHandler.Approve` | `InvestigationService.Approve` | `hr_legal` |
| `GET` | `/api/v1/attendance` | `attendanceHandler.GetAttendance` | `AttendanceService.GetAttendance` | `hr_legal` |
| `POST` | `/api/v1/attendance/:employee_id` | `attendanceHandler.ToggleAttendance`| `AttendanceService.Toggle` | `hr_legal` |
| `GET` | `/api/v1/violations` | `violationHandler.GetAll` | `TrafficViolationService.GetAll` | `hr_legal` |
| `POST` | `/api/v1/violations` | `violationHandler.Create` | `TrafficViolationService.Create` | `hr_legal` |
| `PUT` | `/api/v1/violations/:id` | `violationHandler.Update` | `TrafficViolationService.Update` | `hr_legal` |
| `DELETE` | `/api/v1/violations/:id` | `violationHandler.Delete` | `TrafficViolationService.Delete` | `hr_legal` |
| `GET` | `/api/v1/leaves` | `leaveHandler.GetAll` | `LeaveService.GetAll` | `hr_legal` |
| `POST` | `/api/v1/leaves` | `leaveHandler.Create` | `LeaveService.Create` | `hr_legal` |
| `PUT` | `/api/v1/leaves/:id/status` | `leaveHandler.UpdateStatus` | `LeaveService.UpdateStatus` | `hr_legal` |
| `DELETE` | `/api/v1/leaves/:id` | `leaveHandler.Delete` | `LeaveService.Delete` | `hr_legal` |

---

## 10. Support Tickets & Archive (`ticket` & `system`)

| HTTP Method | Route URL | Current Handler | Service / Function | Target Module |
|---|---|---|---|---|
| `GET` | `/api/v1/tickets` | `ticketHandler.GetAll` | `TicketService.GetAll` | `ticket` |
| `POST` | `/api/v1/tickets` | `ticketHandler.Create` | `TicketService.Create` | `ticket` |
| `PUT` | `/api/v1/tickets/:id` | `ticketHandler.Update` | `TicketService.Update` | `ticket` |
| `DELETE` | `/api/v1/tickets/:id` | `ticketHandler.Delete` | `TicketService.Delete` | `ticket` |
| `GET` | `/api/v1/archive` | `archiveHandler.GetArchived` | `ArchiveService.GetArchived` | `system` |
| `POST` | `/api/v1/archive/restore` | `archiveHandler.Restore` | `ArchiveService.Restore` | `system` |
| `DELETE` | `/api/v1/archive/permanent` | `archiveHandler.PermanentDelete` | `ArchiveService.PermanentDelete` | `system` |
| `POST` | `/api/v1/archive/restore-bulk` | `archiveHandler.BulkRestore` | `ArchiveService.BulkRestore` | `system` |
| `DELETE` | `/api/v1/archive/permanent-bulk` | `archiveHandler.BulkPermanentDelete`| `ArchiveService.BulkPermanentDelete`| `system` |

---

## 11. System, Admin Users, Roles & Notifications (`auth` & `system`)

| HTTP Method | Route URL | Current Handler | Service / Function | Target Module |
|---|---|---|---|---|
| `GET` | `/api/v1/users` | `adminHandler.GetAll` | `AdminService.GetAll` | `auth` |
| `POST` | `/api/v1/users` | `adminHandler.Create` | `AdminService.Create` | `auth` |
| `PUT` | `/api/v1/users/:id` | `adminHandler.Update` | `AdminService.Update` | `auth` |
| `DELETE` | `/api/v1/users/:id` | `adminHandler.Delete` | `AdminService.Delete` | `auth` |
| `POST` | `/api/v1/users/change-password` | `adminHandler.ChangePassword` | `AdminService.ChangePassword` | `auth` |
| `GET` | `/api/v1/roles` | `roleHandler.GetAll` | `RoleService.GetAll` | `auth` |
| `POST` | `/api/v1/roles` | `roleHandler.Create` | `RoleService.Create` | `auth` |
| `GET` | `/api/v1/roles/:id` | `roleHandler.GetByID` | `RoleService.GetByID` | `auth` |
| `PUT` | `/api/v1/roles/:id` | `roleHandler.Update` | `RoleService.Update` | `auth` |
| `DELETE` | `/api/v1/roles/:id` | `roleHandler.Delete` | `RoleService.Delete` | `auth` |
| `GET` | `/api/v1/permissions` | `roleHandler.GetPermissions` | `RoleService.GetPermissions` | `auth` |
| `GET` | `/api/v1/branches` | `branchHandler.GetAll` | `BranchService.GetAll` | `system` |
| `GET` | `/api/v1/branches/:id` | `branchHandler.GetByID` | `BranchService.GetByID` | `system` |
| `POST` | `/api/v1/branches` | `branchHandler.Create` | `BranchService.Create` | `system` |
| `PUT` | `/api/v1/branches/:id` | `branchHandler.Update` | `BranchService.Update` | `system` |
| `DELETE` | `/api/v1/branches/:id` | `branchHandler.Delete` | `BranchService.Delete` | `system` |
| `GET` | `/api/v1/settings` | `settingHandler.GetSettings` | `SettingService.GetSettings` | `system` |
| `PUT` | `/api/v1/settings` | `settingHandler.UpdateSettings` | `SettingService.UpdateSettings` | `system` |
| `GET` | `/api/v1/dashboard` | `dashHandler.GetStats` | `DashboardService.GetStats` | `system` |
| `GET` | `/api/v1/reports` | `reportHandler.GetReports` | `ReportService.GetReports` | `system` |
| `GET` | `/api/v1/reports/export` | `reportHandler.ExportReports` | `ReportService.ExportReports` | `system` |
| `GET` | `/api/v1/reports/daily` | `reportHandler.GetDailyReport` | `ReportService.GetDailyReport` | `system` |
| `GET` | `/api/v1/reports/daily/export` | `reportHandler.ExportDailyReport` | `ReportService.ExportDailyReport` | `system` |
| `GET` | `/api/v1/audit-logs` | `auditHandler.GetLogs` | `AuditService.GetLogs` | `system` |
| `DELETE` | `/api/v1/audit-logs/clear` | `auditHandler.ClearLogs` | `AuditService.ClearLogs` | `system` |
| `DELETE` | `/api/v1/audit-logs/bulk` | `auditHandler.BulkDeleteLogs` | `AuditService.BulkDelete` | `system` |
| `DELETE` | `/api/v1/audit-logs/:id` | `auditHandler.DeleteLog` | `AuditService.DeleteLog` | `system` |
| `GET` | `/api/v1/notifications` | `notifHandler.GetMyNotifications` | `NotificationService.GetMy` | `system` |
| `PUT` | `/api/v1/notifications/read-all` | `notifHandler.MarkAllAsRead` | `NotificationService.MarkAllRead`| `system` |
| `PUT` | `/api/v1/notifications/:id/read` | `notifHandler.MarkAsRead` | `NotificationService.MarkRead` | `system` |
| `POST` | `/api/v1/notifications/broadcast` | `notifHandler.SendBroadcast` | `NotificationService.SendBroadcast`| `system` |
| `GET` | `/api/v1/notifications/broadcasts` | `notifHandler.GetBroadcasts` | `NotificationService.GetBroadcasts`| `system` |
| `DELETE` | `/api/v1/notifications/broadcasts/:id` | `notifHandler.DeleteBroadcast` | `NotificationService.DeleteBroadcast`| `system` |
| `GET` | `/api/v1/notifications/broadcasts/:id/votes` | `notifHandler.GetBroadcastVotes`| `NotificationService.GetVotes` | `system` |
| `GET` | `/api/v1/notifications/employee/broadcasts` | `notifHandler.GetEmployeeBroadcasts` | `NotificationService.GetEmpBroadcasts` | `system` |
| `GET` | `/api/v1/notifications/employee/unread` | `notifHandler.GetEmployeeUnreadBroadcasts`| `NotificationService.GetEmpUnread` | `system` |
| `POST` | `/api/v1/notifications/employee/read/:id` | `notifHandler.MarkEmployeeBroadcastRead` | `NotificationService.MarkEmpRead` | `system` |
| `POST` | `/api/v1/notifications/employee/read-all` | `notifHandler.MarkAllEmployeeBroadcastsRead`| `NotificationService.MarkAllEmpRead`| `system` |
| `POST` | `/api/v1/notifications/employee/vote/:id` | `notifHandler.SubmitVote` | `NotificationService.SubmitVote` | `system` |
| `POST` | `/api/v1/employees/me/push-token` | `notifHandler.SaveEmployeePushToken` | `NotificationService.SaveToken` | `system` |
| `GET` | `/api/v1/otp-requests` | `otpHandler.GetOTPList` | `OTPService.GetOTPList` | `auth` |
| `POST` | `/api/v1/otp-requests/:id/cancel` | `otpHandler.CancelOTP` | `OTPService.CancelOTP` | `auth` |
