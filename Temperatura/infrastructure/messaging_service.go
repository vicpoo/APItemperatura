package infrastructure

import (
	"log"

	"github.com/streadway/amqp"
)

type MessagingService struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	hub  *Hub
}

func NewMessagingService(hub *Hub) *MessagingService {
	conn, err := amqp.Dial("amqp://reyhades:reyhades@44.223.218.9:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
		return nil
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
		return nil
	}

	// Declarar el intercambio
	err = ch.ExchangeDeclare(
		"sensor_data", // nombre del intercambio
		"topic",       // tipo de intercambio
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %s", err)
		return nil
	}

	return &MessagingService{
		conn: conn,
		ch:   ch,
		hub:  hub,
	}
}

func (ms *MessagingService) ConsumeTemperatureMessages() error {
	q, err := ms.ch.QueueDeclare(
		"temperature_queue", // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return err
	}

	// Vincular la cola al exchange con el routing key correcto
	err = ms.ch.QueueBind(
		q.Name,               // queue name
		"sensor.temperature", // routing key
		"sensor_data",        // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := ms.ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack (false para confirmar manualmente)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			log.Printf("Received a temperature message: %s", msg.Body)

			// Enviar el mensaje directamente a todos los clientes WebSocket
			ms.hub.broadcast <- msg.Body

			msg.Ack(false) // Confirmar procesamiento exitoso
		}
	}()

	return nil
}

func (ms *MessagingService) Close() {
	ms.ch.Close()
	ms.conn.Close()
}
