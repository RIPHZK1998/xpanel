package models

import "time"

// SystemConfig represents a system-wide configuration entry
type SystemConfig struct {
	Key         string    `gorm:"primaryKey;size:255" json:"key"`
	Value       string    `gorm:"type:text;not null" json:"value"`
	Encrypted   bool      `gorm:"default:false" json:"encrypted"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy   *uint     `gorm:"index" json:"updated_by,omitempty"`
	UpdatedUser *User     `gorm:"foreignKey:UpdatedBy" json:"updated_user,omitempty"`
}

// TableName specifies the table name for SystemConfig
func (SystemConfig) TableName() string {
	return "system_config"
}

// SystemConfigResponse is the API response format
type SystemConfigResponse struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"` // Masked if encrypted
	Encrypted   bool      `json:"encrypted"`
	Description string    `json:"description,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   *uint     `json:"updated_by,omitempty"`
}

// ToResponse converts SystemConfig to response format
func (sc *SystemConfig) ToResponse(maskValue bool) SystemConfigResponse {
	value := sc.Value
	if sc.Encrypted && maskValue {
		value = "••••••••" // Mask encrypted values
	}

	return SystemConfigResponse{
		Key:         sc.Key,
		Value:       value,
		Encrypted:   sc.Encrypted,
		Description: sc.Description,
		UpdatedAt:   sc.UpdatedAt,
		UpdatedBy:   sc.UpdatedBy,
	}
}
