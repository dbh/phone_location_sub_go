package main

import "gorm.io/gorm"

type PhoneGeo struct {
	gorm.Model
	DeviceId  string  `json:"device_id" `
	Name      string  `json:"device_name" `
	Latitude  float64 `json:"latitude" `
	Longitude float64 `json:"longitude" `
	Speed     float32 `json:"speed" `
	Timestamp uint64  `json:"timestamp" `
}
