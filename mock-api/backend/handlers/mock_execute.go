package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"text/template"

	"strings"

	"backend/models"
	"backend/storage"

	"github.com/gofiber/fiber/v2"
)

// ExecuteMock es el endpoint genérico que intenta hacer coincidir y ejecutar un mock.
func ExecuteMock(c *fiber.Ctx) error {

	reqPath := c.Path()
	reqMethod := c.Method()
	reqQueryParams := c.Queries()

	reqHeaders := make(map[string]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		reqHeaders[strings.ToLower(string(key))] = string(value)
	})

	var reqBody models.RequestBody
	if len(c.Body()) > 0 && strings.Contains(strings.ToLower(c.Get("Content-Type")), "application/json") {
		if err := json.Unmarshal(c.Body(), &reqBody); err != nil {
			log.Printf("Advertencia: No se pudo parsear el cuerpo JSON de la solicitud para %s %s: %v", reqMethod, reqPath, err)
			// Continuar sin el body parseado si hay error que puede ser un JSON mal formado
		}
	}

	allConfigs := storage.GetAllMockConfigurations() // Ya ordenadas por prioridad

	for _, config := range allConfigs {
		// 1. Coincidencia de Ruta y Método
		// Normaliza la ruta de la solicitud para la comparación
		if !matchPath(reqPath, config.Path) || !matchMethod(reqMethod, config.Method) {
			log.Printf("Skipping mock %s: Path '%s' (request) != '%s' (config) OR Method '%s' (request) != '%s' (config)",
				config.Id, reqPath, config.Path, reqMethod, config.Method)
			continue
		}

		// 2. Coincidencia de Query Params
		if !matchQueryParams(reqQueryParams, config.QueryParams) {
			continue
		}

		// 3. Coincidencia de Body Params
		if !matchBodyParams(reqBody, config.BodyParams) {
			continue
		}

		// 4. Coincidencia de Headers
		if !matchHeaders(reqHeaders, config.Headers) {
			continue
		}

		// Si se llega aquí, encontramos una coincidencia.
		// Ahora, procesamos la respuesta, incluyendo las plantillas.

		c.Set("Content-Type", config.ContentType)
		finalResponseBody := config.ResponseBody

		if config.IsTemplate {
			// Si es una plantilla, necesitamos procesarla
			templateString, ok := config.ResponseBody.(string)
			if !ok {
				// Esto no debería ocurrir si ConfigureMock valida correctamente,
				// pero es una buena práctica añadir un fallback o log.
				log.Printf("Error: ResponseBody no es un string a pesar de IsTemplate=true para mock %s", config.Id)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Configuración de mock inválida: el cuerpo de la plantilla no es un string."})
			}

			tmpl, err := template.New("response").Parse(templateString)
			if err != nil {
				log.Printf("Error al parsear plantilla para mock %s: %v", config.Id, err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al procesar la plantilla de respuesta."})
			}

			// Prepara los datos que estarán disponibles para la plantilla
			templateData := fiber.Map{
				"Request": fiber.Map{
					"Path":    reqPath,
					"Method":  reqMethod,
					"Query":   reqQueryParams,
					"Headers": reqHeaders,
					"Body":    reqBody, // Este será un map[string]interface{}
				},
				// Puedes añadir otras variables globales o funciones aquí
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, templateData); err != nil {
				log.Printf("Error al ejecutar plantilla para mock %s: %v", config.Id, err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error al ejecutar la plantilla de respuesta."})
			}

			// El resultado de la plantilla es un string.
			// Si el Content-Type es JSON, necesitamos intentar parsearlo de nuevo a interface{}
			if strings.Contains(strings.ToLower(config.ContentType), "application/json") {
				var parsedTemplateBody interface{}
				if err := json.Unmarshal(buf.Bytes(), &parsedTemplateBody); err != nil {
					log.Printf("Advertencia: La salida de la plantilla no es un JSON válido para mock %s: %v", config.Id, err)
					// Si la salida de la plantilla no es JSON, enviamos como string simple o error
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "La plantilla de respuesta JSON generó un JSON inválido.", "details": err.Error()})
				}
				finalResponseBody = parsedTemplateBody
			} else {
				// Si no es JSON, simplemente lo enviamos como string (ej. text/plain, text/html)
				return c.Status(config.ResponseStatusCode).SendString(buf.String())
			}
		}

		// Enviar la respuesta final (estática o procesada de la plantilla como JSON)
		return c.Status(config.ResponseStatusCode).JSON(finalResponseBody)
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
