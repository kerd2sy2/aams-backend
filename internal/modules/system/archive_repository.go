package system

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (r *gormSystemRepository) GetArchivedItems(ctx context.Context, filter ArchiveFilter) ([]ArchivedItemDTO, int64, ArchiveStatsDTO, error) {
	items := make([]ArchivedItemDTO, 0)
	var total int64
	var stats ArchiveStatsDTO

	// Stats counts across all archivable entities
	r.db.WithContext(ctx).Table("employees").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalEmployees)
	r.db.WithContext(ctx).Table("vehicles").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalVehicles)
	r.db.WithContext(ctx).Table("branches").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalBranches)
	r.db.WithContext(ctx).Table("employee_documents").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalDocuments)
	r.db.WithContext(ctx).Table("work_sessions").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalWorkSessions)
	r.db.WithContext(ctx).Table("leave_requests").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalLeaves)
	r.db.WithContext(ctx).Table("maintenance_requests").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalMaintenance)
	r.db.WithContext(ctx).Table("violations").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalViolations)
	r.db.WithContext(ctx).Table("support_tickets").Unscoped().Where("deleted_at IS NOT NULL").Count(&stats.TotalTickets)
	stats.GrandTotal = stats.TotalEmployees + stats.TotalVehicles + stats.TotalBranches + stats.TotalDocuments +
		stats.TotalWorkSessions + stats.TotalLeaves + stats.TotalMaintenance + stats.TotalViolations + stats.TotalTickets

	likeOp := r.likeOperator()
	typeQuery := filter.Type
	if typeQuery == "" || typeQuery == "all" || typeQuery == "employees" {
		var emps []struct {
			ID         uuid.UUID `gorm:"column:id"`
			Name       string    `gorm:"column:name"`
			NationalID string    `gorm:"column:national_id"`
			Phone      string    `gorm:"column:phone"`
			DeletedAt  time.Time `gorm:"column:deleted_at"`
			CreatedAt  time.Time `gorm:"column:created_at"`
		}
		q := r.db.WithContext(ctx).Table("employees").Unscoped().Where("deleted_at IS NOT NULL")
		if filter.BranchID != nil {
			q = q.Where("branch_id = ?", *filter.BranchID)
		}
		if filter.Search != "" {
			s := "%" + filter.Search + "%"
			searchCond := fmt.Sprintf("name %s ? OR national_id %s ? OR phone %s ?", likeOp, likeOp, likeOp)
			q = q.Where(searchCond, s, s, s)
		}
		_ = q.Order("deleted_at DESC").Limit(50).Scan(&emps).Error
		for _, e := range emps {
			items = append(items, ArchivedItemDTO{
				ID:         e.ID,
				Type:       "employees",
				TypeName:   "مندوب",
				Title:      e.Name,
				Subtitle:   e.NationalID,
				Details:    e.Phone,
				ArchivedAt: e.DeletedAt,
				CreatedAt:  e.CreatedAt,
			})
		}
	}

	total = int64(len(items))
	return items, total, stats, nil
}

func (r *gormSystemRepository) RestoreItem(ctx context.Context, itemType string, id uuid.UUID) error {
	table := r.getTableName(itemType)
	if table == "" {
		return errors.New("نوع العنصر غير صالح")
	}
	return r.db.WithContext(ctx).Table(table).Unscoped().Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *gormSystemRepository) PermanentDeleteItem(ctx context.Context, itemType string, id uuid.UUID) error {
	table := r.getTableName(itemType)
	if table == "" {
		return errors.New("نوع العنصر غير صالح")
	}
	return r.db.WithContext(ctx).Table(table).Unscoped().Where("id = ?", id).Delete(nil).Error
}

func (r *gormSystemRepository) BulkRestoreItems(ctx context.Context, itemType string, ids []uuid.UUID) error {
	table := r.getTableName(itemType)
	if table == "" {
		return errors.New("نوع العنصر غير صالح")
	}
	return r.db.WithContext(ctx).Table(table).Unscoped().Where("id IN ?", ids).Update("deleted_at", nil).Error
}

func (r *gormSystemRepository) BulkPermanentDeleteItems(ctx context.Context, itemType string, ids []uuid.UUID) error {
	table := r.getTableName(itemType)
	if table == "" {
		return errors.New("نوع العنصر غير صالح")
	}
	return r.db.WithContext(ctx).Table(table).Unscoped().Where("id IN ?", ids).Delete(nil).Error
}

func (r *gormSystemRepository) getTableName(itemType string) string {
	switch itemType {
	case "employees":
		return "employees"
	case "vehicles":
		return "vehicles"
	case "branches":
		return "branches"
	case "documents":
		return "employee_documents"
	case "work_sessions":
		return "work_sessions"
	case "leaves":
		return "leave_requests"
	case "maintenance":
		return "maintenance_requests"
	case "violations":
		return "violations"
	case "tickets":
		return "support_tickets"
	default:
		return ""
	}
}
