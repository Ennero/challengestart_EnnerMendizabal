package handlers

import (
	// "bytes"
	"encoding/json"
	"log"

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
		config.ContentType = "application/json" // Asume JSON si es una plantilla (lo más común)

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

// ExecuteMock es el endpoint genérico que intenta hacer coincidir y ejecutar un mock.
func ExecuteMock(c *fiber.Ctx) error {

	// Obtener la ruta, método y parámetros de la solicitud
	reqPath := c.Path()
	reqMethod := c.Method()
	reqQueryParams := c.Queries()

	// Obtener los headers de la solicitud
	reqHeaders := make(map[string]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		reqHeaders[strings.ToLower(string(key))] = string(value) // Convertir a minúsculas por cualquier cosa :)
	})

	// Obtener el body de la solicitud
	var reqBody models.RequestBody
	if len(c.Body()) > 0 && strings.Contains(strings.ToLower(c.Get("Content-Type")), "application/json") {
		if err := json.Unmarshal(c.Body(), &reqBody); err != nil {
			log.Printf("Advertencia: No se pudo parsear el cuerpo JSON de la solicitud para %s %s: %v", reqMethod, reqPath, err)
			// Continuar sin el body parseado si hay error que puede ser un JSON mal formado
		}
	}

	// Obtener todas las configuraciones de mocks
	allConfigs := storage.GetAllMockConfigurations()

	// Iterar sobre todas las configuraciones de mocks
	for _, config := range allConfigs {
		// 1. Coincidencia de Ruta y Método
		if !matchPath(reqPath, config.Path) || !matchMethod(reqMethod, config.Method) {
			continue
		}

		// 2. Coincidencia de Query Params (implementación básica, puedes mejorarla)
		if !matchQueryParams(reqQueryParams, config.QueryParams) {
			continue
		}

		// 3. Coincidencia de Body Params (implementación básica, puedes mejorarla para nested JSON)
		if !matchBodyParams(reqBody, config.BodyParams) {
			continue
		}

		// 4. Coincidencia de Headers (implementación básica)
		if !matchHeaders(reqHeaders, config.Headers) {
			continue
		}

		// Si se llega aquí, encontramos una coincidencia
		c.Set("Content-Type", config.ContentType)
		return c.Status(config.ResponseStatusCode).JSON(config.ResponseBody) // O SendString, Send para otros Content-Types
	}

	// Si no se encuentra ninguna coincidencia
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Mock no encontrado para la solicitud", "path": reqPath, "method": reqMethod})
}

// matchPath verifica si la ruta de la solicitud coincide con la ruta configurada.
// Podrías expandir esto para manejar patrones de URL más complejos.
func matchPath(requestPath, configPath string) bool {
	// Simple comparación directa por ahora.
	// Para /api/v1/productos/:id, necesitarías lógica más avanzada (ej. regex o path parameters en Fiber).
	return requestPath == configPath
}

// matchMethod verifica si el método HTTP de la solicitud coincide con el configurado.
func matchMethod(requestMethod, configMethod string) bool {
	return strings.EqualFold(requestMethod, configMethod)
}

// matchQueryParams verifica si los parámetros de la URL de la solicitud
// contienen todos los parámetros configurados con sus valores correspondientes.
func matchQueryParams(requestParams, configParams map[string]string) bool {
	if len(configParams) == 0 {
		return true // Si no hay parámetros configurados, cualquier query params coinciden
	}
	for key, value := range configParams {
		if reqVal, ok := requestParams[key]; !ok || reqVal != value {
			return false
		}
	}
	return true
}

// matchBodyParams verifica si los parámetros del body de la solicitud
// contienen todos los parámetros configurados con sus valores correspondientes.
// Nota: Esta es una implementación básica y no maneja JSON anidado complejo.
func matchBodyParams(requestBody, configBody map[string]interface{}) bool {
	if len(configBody) == 0 {
		return true
	}
	if len(requestBody) == 0 {
		return false // Si hay parámetros configurados pero no hay body en la request
	}

	for key, configVal := range configBody {
		reqVal, ok := requestBody[key]
		if !ok {
			return false // El parámetro configurado no está en el body de la solicitud
		}
		// Comparación de valores (simple por ahora, para tipos básicos)
		if reqVal != configVal {
			return false
		}
	}
	return true
}

// matchHeaders verifica si los encabezados de la solicitud
// contienen todos los encabezados configurados con sus valores correspondientes.
func matchHeaders(requestHeaders, configHeaders map[string]string) bool {
	if len(configHeaders) == 0 {
		return true
	}
	if len(requestHeaders) == 0 {
		return false // Si hay headers configurados pero no hay headers en la request
	}

	for key, value := range configHeaders {
		// Normalizar a minúsculas para la comparación de claves
		reqVal, ok := requestHeaders[strings.ToLower(key)]
		if !ok || reqVal != value {
			return false
		}
	}
	return true
}
