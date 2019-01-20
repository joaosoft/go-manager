package manager

import "fmt"

// AppConfig ...
type AppConfig struct {
	Manager *ManagerConfig `json:"Manager"`
}

// ManagerConfig ...
type ManagerConfig struct {
	Log struct {
		Level string `json:"level"`
	} `json:"logger"`
}

// NewConfig ...
func NewConfig() (*ManagerConfig, error) {
	appConfig := &AppConfig{}
	if _, err := NewSimpleConfig(fmt.Sprintf("/config/app.%s.json", GetEnv()), appConfig); err != nil {
		log.Error(err.Error())

		return &ManagerConfig{}, err
	}

	return appConfig.Manager, nil
}