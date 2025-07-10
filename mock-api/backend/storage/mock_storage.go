package storage

import (
	"sync"
	"backend/models"
)

// Variables globales para almacenar las configuraciones de mocks
var (
	mockConfigurations = make(map[string]models.MockConfig)
	mutex             	sync.RWMutex
)

func InitMockStorage() {
	// Inicializar el almacenamiento si es necesario
	// Aquí podrías cargar configuraciones desde un archivo o base de datos
}

// AddMockConfig agrega una nueva configuración de mock al almacenamiento.
func AddMockConfig(config models.MockConfig){
	mutex.Lock()
	defer mutex.Unlock()
	mockConfigurations[config.Id] = config
}

// GetMockConfigByID obtiene una configuración de mock por su ID.
func GetMockConfigByID(id string) (models.MockConfig, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	config, ok := mockConfigurations[id]
	return config, ok
}

// GetAllMockConfigurations obtiene todas las configuraciones de mocks.
func GetAllMockConfigurations() []models.MockConfig {
	mutex.RLock()
	defer mutex.RUnlock()
	configs := make([]models.MockConfig, 0, len(mockConfigurations))
	for _, config := range mockConfigurations {
		configs = append(configs, config)
	}
	return configs
}

// DeleteMockConfig elimina una configuración de mock por su ID.
func DeleteMockConfig(id string) bool {
	mutex.Lock()
	defer mutex.Unlock()
	_, exists := mockConfigurations[id]
	if exists {
		delete(mockConfigurations, id)
		return true
	}
	return false
}

