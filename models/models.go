// models/models.go
package models

import "time"

type ClickStat struct {
	Timestamp time.Time `json:"ts"`
	Count     int       `json:"v"`
}
