package main

import (
	"log"
	"os"
	"strings"

	"backend/handlers"
	"backend/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/cors" 

)



func main() {
	app := fiber.New()

	// Middleware para logging de solicitudes
	app.Use(logger.New())


	// Middleware para CORS
    app.Use(cors.New(cors.Config{
        AllowOrigins: "http://localhost:5173", 
        AllowHeaders: "Origin, Content-Type, Accept",
    }))

	// Inicializar el almacenamiento de mocks
	storage.InitMockStorage()

	// Rutas para la gestión de configuraciones de mocks
	app.Post("/configure-mock", handlers.ConfigureMock)
	app.Get("/configure-mock", handlers.GetMockConfigurations)
	app.Delete("/configure-mock/:id", handlers.DeleteMockConfiguration)

	// Endpoint Genérico para la ejecución de mocks
	app.All("/*", handlers.ExecuteMock)

	// Iniciar el servidor Fiber
	port := os.Getenv("PORT")

	// Puerto por defecto si no se especifica
	if port == "" {
		port = "3000" 
	}

	// Asegurarse de que el puerto tenga el prefijo correcto
	if !strings.HasPrefix(port, ":") {
		port = ":" + port 
	}

	log.Printf("Servidor escuchando en el puerto %s", port)
	if err := app.Listen(port); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
