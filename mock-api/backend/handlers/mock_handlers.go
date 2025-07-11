package handlers

import (
	// "bytes"
	"encoding/json"
	"log"

	"strings"

	"backend/models"
	"backend/storage"

	"github.com/gofiber/fiber/v2"
)



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
