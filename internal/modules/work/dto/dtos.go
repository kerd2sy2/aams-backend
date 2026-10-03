package dto

import "time"

type StartWorkRequest struct {
	EmployeeID       string  `json:"employee_id" binding:"required,uuid"`
	StartKM          float64 `json:"start_km" binding:"gte=0"`
	StartKMImage     string  `json:"start_km_image"`
	StartPlateImage  string  `json:"start_plate_image"`
	ApplicationID    string  `json:"application_id"`
	ApplicationType  string  `json:"application_type"`
	VehicleType      string  `json:"vehicle_type"`
	MotorcycleNumber string  `json:"motorcycle_number"`
	Notes            string  `json:"notes"`
}

type EndWorkRequest struct {
	EmployeeID      string  `json:"employee_id" binding:"required,uuid"`
	EndKM           float64 `json:"end_km" binding:"gte=0"`
	EndKMImage      string  `json:"end_km_image"`
	OrdersCount     int     `json:"orders_count"`
	FuelCost        float64 `json:"fuel_cost"`
	ApplicationID   string  `json:"application_id"`
	ApplicationType string  `json:"application_type"`
	Notes           string  `json:"notes"`
	IsReviewed      *bool   `json:"is_reviewed"`
	ReviewNotes     string  `json:"review_notes"`
}

type ReviewWorkSessionRequest struct {
	IsReviewed  bool     `json:"is_reviewed"`
	ReviewNotes string   `json:"review_notes"`
	OrdersCount *int     `json:"orders_count,omitempty"`
	EndKM       *float64 `json:"end_km,omitempty"`
	StartKM     *float64 `json:"start_km,omitempty"`
	FuelCost    *float64 `json:"fuel_cost,omitempty"`
}

type UpdateWorkSessionRequest struct {
	EmployeeID      string     `json:"employee_id"`
	StartKM         float64    `json:"start_km"`
	StartKMImage    string     `json:"start_km_image"`
	StartPlateImage string     `json:"start_plate_image"`
	EndKM           float64    `json:"end_km"`
	EndKMImage      string     `json:"end_km_image"`
	OrdersCount     int        `json:"orders_count"`
	FuelCost        float64    `json:"fuel_cost"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	ApplicationType string     `json:"application_type"`
	Notes           string     `json:"notes"`
}

type OilChangeCheckResponse struct {
	NeedsOilChange          bool    `json:"needs_oil_change"`
	CurrentDistance         float64 `json:"current_distance"`
	LastOilChangeDistance   float64 `json:"last_oil_change_distance"`
	DistanceSinceOil        float64 `json:"distance_since_oil"`
	RemainingDistance       float64 `json:"remaining_distance"`
	Percentage              float64 `json:"percentage"`
	Interval                float64 `json:"interval"`
	VehicleType             string  `json:"vehicle_type"`
	HasActiveMaintenanceReq bool    `json:"has_active_maintenance_request"`
	LastOilChangeDate       string  `json:"last_oil_change_date"`
}

type LastKMResponse struct {
	LastKM           float64 `json:"last_km"`
	VehicleLastKM    float64 `json:"vehicle_last_km"`
	IsDifferentBike  bool    `json:"is_different_bike"`
	MotorcycleNumber string  `json:"motorcycle_number"`
	HasGap           bool    `json:"has_gap"`
	GapKM            float64 `json:"gap_km"`
}
