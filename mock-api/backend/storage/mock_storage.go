package storage

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"sync"

	"backend/models"
)

// Constante con el nmobre del archivo de almacenamiento de mocks
const mocksFileName = "config/mocks.json"

// Variables globales para almacenar las configuraciones de mocks
var (
	mockConfigurations = make(map[string]models.MockConfig)
	mutex             	sync.RWMutex
)

// InitMockStorage inicializa el almacenamiento y carga las configuraciones existentes desde el archivo.
func InitMockStorage() {
	mutex.Lock()
	defer mutex.Unlock()

	// Intenta leer el archivo de mocks
	data, err := os.ReadFile(mocksFileName)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Archivo de mocks '%s' no encontrado. Iniciando con almacenamiento vacío.", mocksFileName)
			return // El archivo no existe, no hay nada que cargar
		}
		log.Printf("Error al leer el archivo de mocks '%s': %v", mocksFileName, err)
		return // Otro error de lectura
	}

	// Si el archivo existe y se leyó, intenta deserializar el JSON
	if err := json.Unmarshal(data, &mockConfigurations); err != nil {

		// Si hay un error de deserialización, loguea el error y comienza con un almacenamiento vacío
		log.Printf("Error al deserializar mocks desde '%s': %v. Iniciando con almacenamiento vacío.", mocksFileName, err)
		mockConfigurations = make(map[string]models.MockConfig)
	} else {
		log.Printf("Mocks cargados exitosamente desde '%s'. Total: %d", mocksFileName, len(mockConfigurations))
	}
}


// saveMocksToFile guarda las configuraciones de mocks en el archivo JSON
func saveMocksToFile() error {
	// Serializa el mapa de configuraciones a JSON
	data, err := json.MarshalIndent(mockConfigurations, "", "  ")
	if err != nil {
		log.Printf("Error al serializar mocks a JSON: %v", err)
		return err
	}

	// Escribe el JSON al archivo
	if err := os.WriteFile(mocksFileName, data, 0644); err != nil { 
		log.Printf("Error al escribir mocks en el archivo '%s': %v", mocksFileName, err)
		return err
	}

	log.Printf("Mocks guardados exitosamente en '%s'. Total: %d", mocksFileName, len(mockConfigurations))
	return nil
}





// AddMockConfig agrega una nueva configuración de mock al almacenamiento.
func AddMockConfig(config models.MockConfig) error {
	mutex.Lock()
	defer mutex.Unlock()
	mockConfigurations[config.Id] = config
	err := saveMocksToFile() // Guarda las configuraciones en el archivo después de agregar

	if err != nil {
		log.Printf("Error al guardar la configuración del mock '%s': %v", config.Id, err)
		return err
	}

	return nil
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

	// Convertir el mapa a un slice para facilitar la iteración
	configs := make([]models.MockConfig, 0, len(mockConfigurations))
	for _, config := range mockConfigurations {
		configs = append(configs, config)
	}

	// Ordenar las configuraciones por prioridad antes de devolverlas
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Priority > configs[j].Priority
	})

	return configs
}

// DeleteMockConfig elimina una configuración de mock por su ID y la guarda.
func DeleteMockConfig(id string) bool {
	mutex.Lock()
	defer mutex.Unlock()
	_, exists := mockConfigurations[id]
	if exists {
		delete(mockConfigurations, id)
		saveMocksToFile() // Guarda después de eliminar
		return true
	}
	return false
}

