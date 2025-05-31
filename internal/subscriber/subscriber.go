package subscriber

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/segmentio/kafka-go"
	"log"
	"searchengine/internal/config"
	"searchengine/internal/index"
	"searchengine/pkg/model"
)

type Subscriber struct {
	index *index.Index
	cfg   *config.Config
}

func New(cfg *config.Config, index *index.Index) *Subscriber {
	return &Subscriber{
		index: index,
		cfg:   cfg,
	}
}

func (s *Subscriber) Start(ctx context.Context) {
	if s.cfg.EnableNatsSubscriber {
		go s.startNATS(ctx)
		log.Println("[Subscriber] Started NATS subscriber")
	}

	if s.cfg.EnableKafkaSubscriber {
		go s.startKafka(ctx)
		log.Println("[Subscriber] Started Kafka subscriber")
	}
}

func (s *Subscriber) startNATS(ctx context.Context) {
	nc, err := nats.Connect(s.cfg.NatsURL)
	if err != nil {
		log.Printf("[NATS] Connect error: %v\n", err)
		return
	}
	defer nc.Close()

	_, err = nc.Subscribe(s.cfg.NatsSubject, func(msg *nats.Msg) {
		// Обработка сообщения
		s.handleMessage(msg.Data)
	})
	if err != nil {
		log.Printf("[NATS] Subscribe error: %v\n", err)
		return
	}

	<-ctx.Done()
	log.Println("[NATS] Subscriber stopped")
}

func (s *Subscriber) startKafka(ctx context.Context) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{s.cfg.KafkaURL},
		Topic:   s.cfg.KafkaTopic,
		GroupID: "search_engine_consumer",
	})
	defer r.Close()

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("[Kafka] ReadMessage error: %v\n", err)
			break
		}
		s.handleMessage(m.Value)
	}
	log.Println("[Kafka] Subscriber stopped")
}

func (s *Subscriber) handleMessage(data []byte) {
	var docMsg model.DocMsg
	err := json.Unmarshal(data, &docMsg)
	if err != nil {
		log.Printf("[Subscriber] Unmarshal error: %v\n", err)
	}

	log.Printf("[Subscriber] Received message: %s\n", string(data))
	if !docMsg.Deleted {
		err = s.index.Update(docMsg.DocId, docMsg.Document)
		if err != nil {
			log.Printf("[Subscriber] Update error: %v\n", err)
		}
	} else {
		err = s.index.Delete(docMsg.DocId)
	}

}
