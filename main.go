// main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/vicpoo/APItemperatura/Temperatura/infrastructure"
)

// Middleware para manejar CORS (Cross-Origin Resource Sharing)
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Permitir solicitudes desde cualquier origen (CORS)
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Si es una solicitud OPTIONS, responder inmediatamente
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// Continuar con la siguiente función en la pila de middlewares
		c.Next()
	}
}

// Middleware para registrar las solicitudes entrantes
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Request: %s %s", c.Request.Method, c.Request.URL.Path)
		// Llamar a la siguiente función en la pila
		c.Next()
	}
}

func main() {
	// Configurar Gin
	r := gin.Default()

	// Agregar los middlewares globalmente
	r.Use(CORS())   // Middleware para manejar CORS
	r.Use(Logger()) // Middleware para registrar las solicitudes

	// Inicializar el hub de WebSocket y el consumidor
	hub := infrastructure.NewHub()
	go hub.Run()

	// Configurar servicio de mensajería
	messagingService := infrastructure.NewMessagingService(hub)
	defer messagingService.Close()

	// Configurar rutas
	infrastructure.SetupRoutes(r, hub)

	// Iniciar consumidor de RabbitMQ para temperatura
	if err := messagingService.ConsumeTemperatureMessages(); err != nil {
		log.Fatalf("Failed to start RabbitMQ consumer: %v", err)
	}

	// Manejar señales para apagado limpio
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar servidor en una goroutine
	go func() {
		if err := r.Run(":8001"); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Println("Server started on port 8001")
	log.Println("Temperature consumer started")

	// Esperar señal de apagado
	<-sigChan
	log.Println("Shutting down server...")
}
