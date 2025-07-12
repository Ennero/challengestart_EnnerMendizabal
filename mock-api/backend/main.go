package main

import (
	"log"
	"os"
	"strings"

	"backend/handlers"
	"backend/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Configuración de Fiber con un manejador de errores global
	app := fiber.New(fiber.Config{

		// ErrorHandler es una función que se ejecuta cuando un handler retorna un error
		ErrorHandler: func(c *fiber.Ctx, err error) error {

			// Determinar el código de estado HTTP
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				// Si el error es un *fiber.Error usa el código de estado de ese error.
				code = e.Code
			}

			// Establecer el Content-Type de la respuesta como JSON
			c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			// Devolver una respuesta JSON consistente con el error
			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
				"code":    code, // Código de estado HTTP
			})
		},
	})

	// Middleware para logging de solicitudes
	app.Use(logger.New())

	// Middleware para CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173", // Ajusta esto a la URL de tu frontend
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
