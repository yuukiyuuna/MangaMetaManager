package models

type AppSettings struct {
	ID                  uint `gorm:"primaryKey"`
	BackupBeforeFlatten bool `json:"backupBeforeFlatten" gorm:"default:true"`
}
