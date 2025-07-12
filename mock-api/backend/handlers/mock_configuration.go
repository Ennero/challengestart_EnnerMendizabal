package handlers

import (
	// "bytes"
	"encoding/json"

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

	// Generar un ID único para la configuración si no se proporciona
	if config.Id == "" {
		config.Id = uuid.New().String()
	}

	// Normalización de campos
	config.Path = strings.TrimSpace(config.Path)
	config.Method = strings.ToUpper(strings.TrimSpace(config.Method))
	config.ContentType = strings.TrimSpace(config.ContentType)

	// Validaciones básicas
	if config.Path == "" || config.Method == "" || config.ResponseStatusCode == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Campos Path, Method y ResponseStatusCode son requeridos"})
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

	// Validación del ResponseBody
	if config.ContentType == "" {
		// Asignar un Content-Type por defecto si no se proporciona y hay ResponseBody
		if config.ResponseBody != nil {
			config.ContentType = "application/json" // Por defecto un JSON
		} else {
			config.ContentType = "text/plain" // Sin body, o es una string simple
		}
	}

	// Si no es una plantilla y el Content-Type es JSON, validar que ResponseBody sea JSON válido
	if !config.IsTemplate && strings.Contains(strings.ToLower(config.ContentType), "application/json") {

		// Intentar unmarshal ResponseBody para asegurar que es un JSON válido
		var raw json.RawMessage
		if rbBytes, err := json.Marshal(config.ResponseBody); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ResponseBody no es un formato JSON válido.", "details": err.Error()})
		} else if err := json.Unmarshal(rbBytes, &raw); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "El ResponseBody no es un JSON válido o la estructura no coincide con el Content-Type.", "details": err.Error()})
		}

		// Reasignar para mantener consistencia
		config.ResponseBody = raw

	} else if config.IsTemplate && config.ContentType == "" {

		// Si es una plantilla, pero no se especificó un ContentType, asigna uno por defecto
		config.ContentType = "application/json"

	} else if config.IsTemplate {
		// Asegurar de que ResponseBody es un string si se marca como plantilla
		if _, ok := config.ResponseBody.(string)
		!ok {
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