package models

// MockConfig representa la configuración de un mock.
type MockConfig struct {
	Id                 string            		`json:"id"` 			
	Path               string            		`json:"path"`
	Method             string            		`json:"method"`
	QueryParams        map[string]string 		`json:"queryParams"` 
	BodyParams         map[string]interface{} 	`json:"bodyParams"` 	
	Headers            map[string]string 		`json:"headers"`
	ResponseStatusCode int               		`json:"responseStatusCode"`
	ResponseBody       interface{}       		`json:"responseBody"` 
	ContentType        string            		`json:"contentType"`
	
	// Agrega campos para lógica condicional si es necesario
	ConditionalLogic string `json:"conditionalLogic,omitempty"` // Ejemplo: JS o Go template
}

// Para facilitar la deserialización de parámetros del body, si es JSON
type RequestBody map[string]interface{}