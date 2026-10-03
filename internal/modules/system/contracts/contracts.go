package contracts

import (
	"context"

	"github.com/google/uuid"
)

// IAuditLogger is a system contract for logging admin/user actions across modules.
type IAuditLogger interface {
	Log(ctx context.Context, adminName, action, details, ipAddress string, branchID *uuid.UUID) error
}

// INotificationSender is a system contract for dispatching notifications.
type INotificationSender interface {
	SendNotification(ctx context.Context, branchID, adminID, employeeID *uuid.UUID, title, body, notifType string) error
}
