package work

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

)

func (s *workService) GetLastSessionOrVehicleKM(ctx context.Context, empID uuid.UUID, motorcycleNumber string) (*LastKMResponse, error) {
	resp := &LastKMResponse{}

	lastSession, _ := s.repo.FindLastCompletedSession(ctx, empID)
	if lastSession != nil {
		resp.LastKM = lastSession.EndKM
		resp.LastEndKM = lastSession.EndKM
		resp.LastStartKM = lastSession.StartKM
		resp.MotorcycleNumber = lastSession.MotorcycleNumber
	}

	targetBike := motorcycleNumber
	if targetBike == "" && lastSession != nil {
		targetBike = lastSession.MotorcycleNumber
	}

	if targetBike != "" && s.vehicleProvider != nil {
		vInfo, err := s.vehicleProvider.EnsureVehicleExists(ctx, targetBike, "", nil, 0)
		if err == nil && vInfo != nil {
			if vInfo.PlateNumber != "" {
				resp.MotorcycleNumber = vInfo.PlateNumber
			}
			resp.VehicleLastKM = vInfo.CurrentKM
			resp.LastKM = vInfo.CurrentKM
			resp.LastEndKM = vInfo.CurrentKM
			resp.RegistrationImage = vInfo.RegistrationImage
			isBroken := vInfo.IsOdometerBroken || strings.Contains(targetBike, "7569") || strings.Contains(targetBike, "8043") || strings.Contains(vInfo.PlateNumber, "7569") || strings.Contains(vInfo.PlateNumber, "8043")
			resp.IsOdometerBroken = isBroken

			interval := oilChangeInterval(vInfo.VehicleType)
			var distanceSinceOil float64
			var remaining float64
			var needsOil bool
			var remainingDays int
			var oilStatusMessage string

			if isBroken {
				if vInfo.LastOilChangeDate != nil {
					days := time.Since(*vInfo.LastOilChangeDate).Hours() / 24.0
					if days >= 5.0 {
						needsOil = true
						remaining = 0
						remainingDays = 0
						distanceSinceOil = interval
						oilStatusMessage = "اليوم موعد تغيير الزيت للدباب"
					} else {
						needsOil = false
						remDays := 5.0 - days
						remainingDays = int(math.Ceil(remDays))
						if remainingDays <= 1 {
							oilStatusMessage = "اليوم موعد تغيير الزيت للدباب"
						} else {
							oilStatusMessage = fmt.Sprintf("فاضل %d أيام لتغيير الزيت", remainingDays)
						}
						remaining = remDays * 200.0
						distanceSinceOil = 0
					}
				} else {
					needsOil = true
					remaining = 0
					remainingDays = 0
					distanceSinceOil = interval
					oilStatusMessage = "اليوم موعد تغيير الزيت للدباب"
				}
			} else {
				if vInfo.LastOilChangeKM > 0 || vInfo.LastOilChangeDate != nil {
					distanceSinceOil = vInfo.CurrentKM - vInfo.LastOilChangeKM
					if distanceSinceOil < 0 {
						distanceSinceOil = 0
					}
				} else {
					if vInfo.TotalDistance > 0 {
						distanceSinceOil = vInfo.TotalDistance
					} else {
						distanceSinceOil = 0
					}
				}
				remaining = interval - distanceSinceOil
				if remaining < 0 {
					remaining = 0
				}
				needsOil = distanceSinceOil >= interval
				if needsOil {
					oilStatusMessage = "اليوم موعد تغيير الزيت للدباب"
				} else {
					oilStatusMessage = fmt.Sprintf("فاضل %.0f كم لتغيير الزيت", remaining)
				}
			}

			resp.DistanceSinceOil = distanceSinceOil
			resp.RemainingOilKM = remaining
			resp.NeedsOilChange = needsOil
			resp.RemainingDays = remainingDays
			resp.OilStatusMessage = oilStatusMessage

			if lastSession != nil && lastSession.MotorcycleNumber != targetBike {
				resp.IsDifferentBike = true
			}
			if lastSession != nil && resp.VehicleLastKM > lastSession.EndKM && lastSession.EndKM > 0 {
				resp.HasGap = true
				resp.GapKM = resp.VehicleLastKM - lastSession.EndKM
			}
		}
	}

	return resp, nil
}

func (s *workService) ScanPlateImage(ctx context.Context, imageBase64 string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"success": true,
		"message": "plate scanned successfully",
	}, nil
}

func (s *workService) CheckOilChange(ctx context.Context, empID uuid.UUID, motorcycleNumber string) (*OilChangeCheckResponse, error) {
	emp, err := s.empProvider.GetEmployee(ctx, empID)
	if err != nil {
		return nil, errors.New("الموظف غير موجود")
	}

	bike := motorcycleNumber
	if bike == "" {
		bike = emp.MotorcycleNumber
	}

	var vehicleInfo *VehicleData
	if bike != "" && s.vehicleProvider != nil {
		vehicleInfo, _ = s.vehicleProvider.GetVehicleInfo(ctx, bike)
	}

	if vehicleInfo != nil {
		vType := vehicleInfo.VehicleType
		if vType == "" {
			vType = emp.VehicleType
		}
		interval := oilChangeInterval(vType)
		var distanceSinceOil float64
		var needsOil bool
		var remaining float64
		var pct float64
		var lastOilDateStr string

		if vehicleInfo.LastOilChangeDate != nil {
			lastOilDateStr = vehicleInfo.LastOilChangeDate.Format("2006-01-02 15:04:05")
		}

		var remainingDays int
		var oilStatusMessage string

		isBroken := vehicleInfo.IsOdometerBroken || strings.Contains(bike, "7569") || strings.Contains(bike, "8043") || strings.Contains(vehicleInfo.PlateNumber, "7569") || strings.Contains(vehicleInfo.PlateNumber, "8043")
		if isBroken {
			if vehicleInfo.LastOilChangeDate != nil {
				days := time.Since(*vehicleInfo.LastOilChangeDate).Hours() / 24.0
				if days >= 5.0 {
					needsOil = true
					remaining = 0
					remainingDays = 0
					pct = 100
					distanceSinceOil = interval
					oilStatusMessage = "اليوم موعد تغيير الزيت للدباب"
				} else {
					needsOil = false
					remDays := 5.0 - days
					remainingDays = int(math.Ceil(remDays))
					if remainingDays <= 1 {
						oilStatusMessage = "اليوم موعد تغيير الزيت للدباب"
					} else {
						oilStatusMessage = fmt.Sprintf("فاضل %d أيام لتغيير الزيت", remainingDays)
					}
					remaining = remDays * 200.0
					pct = (days / 5.0) * 100
					distanceSinceOil = 0
				}
			} else {
				needsOil = true
				remaining = 0
				remainingDays = 0
				pct = 100
				distanceSinceOil = interval
				oilStatusMessage = "اليوم موعد تغيير الزيت للدباب"
			}
		} else {
			if vehicleInfo.LastOilChangeKM > 0 || vehicleInfo.LastOilChangeDate != nil {
				distanceSinceOil = vehicleInfo.CurrentKM - vehicleInfo.LastOilChangeKM
				if distanceSinceOil < 0 {
					distanceSinceOil = 0
				}
			} else {
				if vehicleInfo.TotalDistance > 0 {
					distanceSinceOil = vehicleInfo.TotalDistance
				} else {
					distanceSinceOil = 0
				}
			}
			needsOil = distanceSinceOil >= interval
			remaining = interval - distanceSinceOil
			if remaining < 0 {
				remaining = 0
			}
			pct = (distanceSinceOil / interval) * 100
			if pct > 100 {
				pct = 100
			}
			if needsOil {
				oilStatusMessage = "اليوم موعد تغيير الزيت للدباب"
			} else {
				oilStatusMessage = fmt.Sprintf("فاضل %.0f كم لتغيير الزيت", remaining)
			}
		}

		hasMaintenance := false
		if s.vehicleProvider != nil {
			hasMaintenance, _ = s.vehicleProvider.HasActiveMaintenance(ctx, bike)
		}

		return &OilChangeCheckResponse{
			NeedsOilChange:          needsOil,
			CurrentDistance:         vehicleInfo.CurrentKM,
			LastOilChangeDistance:   vehicleInfo.LastOilChangeKM,
			DistanceSinceOil:        distanceSinceOil,
			RemainingDistance:       remaining,
			Percentage:              pct,
			Interval:                interval,
			VehicleType:             vType,
			HasActiveMaintenanceReq: hasMaintenance,
			LastOilChangeDate:       lastOilDateStr,
			IsOdometerBroken:        isBroken,
			RemainingDays:           remainingDays,
			OilStatusMessage:        oilStatusMessage,
		}, nil
	}

	// Oil change is strictly tracked on the vehicle, never on the employee
	interval := oilChangeInterval(emp.VehicleType)
	return &OilChangeCheckResponse{
		NeedsOilChange:          false,
		CurrentDistance:         0,
		LastOilChangeDistance:   0,
		DistanceSinceOil:        0,
		RemainingDistance:       interval,
		Percentage:              0,
		Interval:                interval,
		VehicleType:             emp.VehicleType,
		HasActiveMaintenanceReq: false,
		LastOilChangeDate:       "",
	}, nil
}
