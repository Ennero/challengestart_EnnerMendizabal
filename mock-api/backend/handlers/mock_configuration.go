package handlers

import (
	// "bytes"
	"encoding/json"
	"regexp"
	"strings"

	"backend/models"
	"backend/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid" // Para generar los IDs únicos
)

// ConfigureMock maneja la solicitud POST /configure-mock
func ConfigureMock(c *fiber.Ctx) error {
	var config models.MockConfig

	// Parsear el cuerpo de la solicitud a la estructura MockConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No se pudo parsear la configuración del mock", "details": err.Error()})
	}

	// Normalización de campos
	config.Path = strings.TrimSpace(config.Path)
	config.Method = strings.ToUpper(strings.TrimSpace(config.Method))
	config.ContentType = strings.TrimSpace(config.ContentType)

	// Generar un ID único para la configuración si no se proporciona
	if config.Id == "" {
		config.Id = uuid.New().String()
	}

	// VALIDACIONES
	// Validaciones de campos requeridos
	if config.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El campo 'path' es requerido y no puede estar vacío."})
	}
	if config.Method == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El campo 'method' es requerido y no puede estar vacío."})
	}
	if config.ResponseStatusCode == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El campo 'responseStatusCode' es requerido y no puede ser 0."})
	}

	// Validación de formato de Path
	var pathRegex = regexp.MustCompile(`^(/[\w.-]*)*(/[\w.-]*)?$`) // Permite /path/to/resource, /resource, etc.
	if !pathRegex.MatchString(config.Path) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El campo 'path' tiene un formato URL inválido. Ejemplos válidos: /api/v1/users, /hello-world."})
	}

	// Validación de método HTTP valido
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
		"CONNECT": true,
		"TRACE":   true,
	}
	if !validMethods[config.Method] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Método HTTP inválido. Los métodos permitidos son: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, CONNECT, TRACE."})
	}

	// Inicializar mapas vacíos por cualquier cosa
	if config.QueryParams == nil {
		config.QueryParams = make(map[string]string)
	}
	if config.BodyParams == nil {
		config.BodyParams = make(map[string]interface{})
	}
	if config.Headers == nil {
		config.Headers = make(map[string]string)
	}


	// Validación y asignación de Content-Type por defecto
	validContentTypes := map[string]bool{
		"application/json": true,
		"text/plain":       true,
		"text/html":        true, 
		"application/xml":  true, 
		"application/octet-stream": true, // Para archivos binarios
	}


	// Validación del ResponseBody
	if config.ContentType == "" {

		// Asignar un Content-Type por defecto si no se proporciona
		if config.ResponseBody != nil {
			config.ContentType = "application/json" // Por defecto JSON si hay body
		} else {
			config.ContentType = "text/plain" // Por defecto texto plano si no hay body
		}
	} else {
		if !validContentTypes[config.ContentType] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Content-Type inválido. Los Content-Types permitidos son: " + strings.Join(getKeys(validContentTypes), ", ") + "."})
		}
	}

	// Validación del ResponseBody basado en Content-Type y IsTemplate
	if !config.IsTemplate {
		if config.ResponseBody == nil {
			// Si no hay ResponseBody y Content-Type es JSON, asignar un objeto JSON vacío
			if config.ContentType == "application/json" {
				config.ResponseBody = map[string]interface{}{}
			}

			// Si ResponseBody es nil y no es JSON, se asume que no hay contenido de respuesta
		} else if config.ContentType == "application/json" {

			// Intentar serializar y deserializar para validar que ResponseBody sea JSON válido
			rbBytes, err := json.Marshal(config.ResponseBody)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El 'responseBody' no pudo ser serializado a JSON.", "details": err.Error()})
			}
			var temp interface{}
			if err := json.Unmarshal(rbBytes, &temp); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El 'responseBody' no es un JSON válido o la estructura no coincide con 'application/json'.", "details": err.Error()})
			}
			config.ResponseBody = temp // Asegurar que ResponseBody sea un objeto JSON válido
		} else {
			// Si no es una plantilla y no es JSON, asegurar que ResponseBody sea un string
			if _, ok := config.ResponseBody.(string); !ok {
				// Intentar convertir a string si no lo es 
				rbBytes, err := json.Marshal(config.ResponseBody)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El 'responseBody' no pudo ser convertido a string para el 'Content-Type' especificado.", "details": err.Error()})
				}
				config.ResponseBody = string(rbBytes)
			}
		}
	} else { // config.IsTemplate es true
		// Si es una plantilla, ResponseBody debe ser un string
		if _, ok := config.ResponseBody.(string); !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Si 'isTemplate' es verdadero, 'responseBody' debe ser un string que contenga la plantilla."})
		}
	}

	// Agregar la configuración del mock al almacenamiento
	err := storage.AddMockConfig(config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "No se pudo guardar la configuración del mock en el almacenamiento persistente.", "details": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Configuración de mock guardada exitosamente", "id": config.Id})
}

// getKeys es una función auxiliar para obtener las claves de un mapa de booleanos
func getKeys(m map[string]bool) []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

// GetMockConfigurations maneja la solicitud GET /configure-mock
func GetMockConfigurations(c *fiber.Ctx) error {
	configs := storage.GetAllMockConfigurations()
	return c.Status(fiber.StatusOK).JSON(configs)
}

// DeleteMockConfiguration maneja la solicitud DELETE /configure-mock/:id
func DeleteMockConfiguration(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID del mock es requerido"})
	}

	if deleted := storage.DeleteMockConfig(id); !deleted {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Configuración de mock no encontrada"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Configuración de mock eliminada exitosamente"})
}