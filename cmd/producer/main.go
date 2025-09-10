

package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	kfk "github.com/segmentio/kafka-go"
	"github.com/Uva337/WBL0v1/internal/models"
)

func main() {
	count := flag.Int("n", 10, "number of messages to generate")
	flag.Parse()

	brokers := envOr("KAFKA_BROKERS", "localhost:9092")
	topic := envOr("KAFKA_TOPIC", "orders")

	ensureTopicExists(brokers, topic)

	w := &kfk.Writer{
		Addr:         kfk.TCP(brokers),
		Topic:        topic,
		Balancer:     &kfk.LeastBytes{},
		RequiredAcks: kfk.RequireOne,
		Async:        false,
	}
	defer w.Close()

	log.Printf("generating and sending %d messages to topic '%s'...", *count, topic)
	var messages []kfk.Message

	for i := 0; i < *count; i++ {
		order := generateFakeOrder()
		log.Printf("Generated OrderUID: %s", order.OrderUID)
		data, err := json.Marshal(order)
		if err != nil {
			log.Printf("failed to marshal fake order: %v", err)
			continue
		}
		messages = append(messages, kfk.Message{Value: data})
	}

	if err := w.WriteMessages(context.Background(), messages...); err != nil {
		log.Fatalf("failed to write messages: %v", err)
	}

	log.Printf("%d messages sent successfully", len(messages))
}


func ensureTopicExists(brokerAddress, topicName string) {
	conn, err := kfk.Dial("tcp", brokerAddress)
	if err != nil {
		log.Fatalf("failed to connect to kafka: %v", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Fatalf("failed to get kafka controller: %v", err)
	}

	controllerConn, err := kfk.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Fatalf("failed to connect to kafka controller: %v", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kfk.TopicConfig{
		{
			Topic:             topicName,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		
		if e, ok := err.(kfk.Error); ok && e == kfk.TopicAlreadyExists {
			log.Printf("topic '%s' already exists. proceeding...", topicName)
			return
		}
		log.Fatalf("failed to create kafka topic: %v", err)
	}

	log.Printf("topic '%s' created successfully or already exists.", topicName)
}

func generateFakeOrder() models.Order {
	var items []models.Item
	for i := 0; i < gofakeit.Number(1, 5); i++ {
		items = append(items, models.Item{
			ChrtID:      gofakeit.Number(1000000, 9999999),
			TrackNumber: gofakeit.UUID(),
			Price:       gofakeit.Number(100, 5000),
			RID:         gofakeit.UUID(),
			Name:        gofakeit.Word(),
			Sale:        gofakeit.Number(10, 80),
			Size:        "0",
			TotalPrice:  gofakeit.Number(100, 5000),
			NmID:        gofakeit.Number(1000000, 9999999),
			Brand:       gofakeit.Company(),
			Status:      202,
		})
	}

	return models.Order{
		OrderUID:    gofakeit.UUID(),
		TrackNumber: gofakeit.UUID(),
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Address().Address,
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		},
		Payment: models.Payment{
			Transaction:  gofakeit.UUID(),
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       gofakeit.Number(1000, 10000),
			PaymentDT:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: gofakeit.Number(100, 1000),
			GoodsTotal:   gofakeit.Number(100, 5000),
			CustomFee:    0,
		},
		Items:             items,
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        gofakeit.Username(),
		DeliveryService:   "meest",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
