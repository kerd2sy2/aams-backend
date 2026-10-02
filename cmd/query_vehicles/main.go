package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Vehicle struct {
	ID                string `gorm:"primaryKey"`
	PlateNumber       string
	VehicleType       string
	RegistrationImage string
	Status            string
}

func main() {
	dsn := "host=localhost user=aams_user password=Qx7Kp2Mz9Wr4Tv8Yn3La6Se1Dg5Fh0Jc dbname=delivery_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	var list []Vehicle
	db.Table("vehicles").Find(&list)
	fmt.Printf("Total vehicles in DB: %d\n", len(list))
	for i, v := range list {
		fmt.Printf("%d: %s (Type: %s, Status: %s, HasRegImage: %t)\n", i+1, v.PlateNumber, v.VehicleType, v.Status, v.RegistrationImage != "")
	}
}
