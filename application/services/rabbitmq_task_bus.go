package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

type JobHandler func(context.Context, AsyncJob) error

type JobDispatcher interface {
	Dispatch(job AsyncJob) error
	DispatchDelayed(job AsyncJob, delay time.Duration) error
}

type RabbitMQTaskBus struct {
	cfg            config.MQConfig
	log            *logger.Logger
	queueName      string
	delayQueueName string

	conn      *amqp.Connection
	publishCh *amqp.Channel
	consumeCh *amqp.Channel
	pool      *WorkerPool

	mu       sync.RWMutex
	handlers map[string]JobHandler
}

func NewRabbitMQTaskBus(cfg config.MQConfig, log *logger.Logger) (*RabbitMQTaskBus, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	if cfg.URL == "" {
		return nil, fmt.Errorf("mq url is required when mq is enabled")
	}

	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	publishCh, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open publisher channel: %w", err)
	}

	consumeCh, err := conn.Channel()
	if err != nil {
		_ = publishCh.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("open consumer channel: %w", err)
	}

	queueName := fmt.Sprintf("%s.task.default", normalizeQueuePrefix(cfg.QueuePrefix))
	delayQueueName := queueName + ".delay"
	if _, err := publishCh.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		_ = consumeCh.Close()
		_ = publishCh.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}
	delayQueueArgs := amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": queueName,
	}
	if _, err := publishCh.QueueDeclare(delayQueueName, true, false, false, false, delayQueueArgs); err != nil {
		_ = consumeCh.Close()
		_ = publishCh.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare delay queue: %w", err)
	}
	if _, err := consumeCh.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		_ = consumeCh.Close()
		_ = publishCh.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare consumer queue: %w", err)
	}

	return &RabbitMQTaskBus{
		cfg:            cfg,
		log:            log,
		queueName:      queueName,
		delayQueueName: delayQueueName,
		conn:           conn,
		publishCh:      publishCh,
		consumeCh:      consumeCh,
		pool:           NewWorkerPool(log, cfg.ConsumerConcurrency),
		handlers:       make(map[string]JobHandler),
	}, nil
}

func (b *RabbitMQTaskBus) Register(jobType string, handler JobHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[jobType] = handler
}

func (b *RabbitMQTaskBus) Dispatch(job AsyncJob) error {
	return b.dispatch(job, b.queueName, "")
}

func (b *RabbitMQTaskBus) DispatchDelayed(job AsyncJob, delay time.Duration) error {
	if delay <= 0 {
		return b.Dispatch(job)
	}
	return b.dispatch(job, b.delayQueueName, fmt.Sprintf("%d", delay.Milliseconds()))
}

func (b *RabbitMQTaskBus) dispatch(job AsyncJob, queueName string, expiration string) error {
	if b == nil {
		return errors.New("rabbitmq task bus is nil")
	}

	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}

	return b.publishCh.PublishWithContext(
		context.Background(),
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Type:         job.Type,
			Expiration:   expiration,
		},
	)
}

func (b *RabbitMQTaskBus) Start() error {
	if b == nil || !b.cfg.ConsumerEnabled {
		return nil
	}

	prefetch := b.cfg.PrefetchCount
	if prefetch <= 0 {
		prefetch = max(1, b.cfg.ConsumerConcurrency)
	}
	if err := b.consumeCh.Qos(prefetch, 0, false); err != nil {
		return fmt.Errorf("set rabbitmq qos: %w", err)
	}

	deliveries, err := b.consumeCh.Consume(
		b.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume queue: %w", err)
	}

	go func() {
		for delivery := range deliveries {
			d := delivery
			submitErr := b.pool.Submit("mq."+d.Type, func() {
				b.handleDelivery(d)
			})
			if submitErr != nil {
				b.log.Errorw("Failed to enqueue rabbitmq delivery into worker pool", "error", submitErr, "job_type", d.Type)
				_ = d.Nack(false, true)
			}
		}
	}()

	if b.log != nil {
		b.log.Infow("RabbitMQ task consumer started",
			"queue", b.queueName,
			"prefetch", prefetch,
			"concurrency", b.cfg.ConsumerConcurrency)
	}

	return nil
}

func (b *RabbitMQTaskBus) Stop(ctx context.Context) error {
	if b == nil {
		return nil
	}

	var errs []error
	if err := b.pool.Stop(ctx); err != nil && !errors.Is(err, context.Canceled) {
		errs = append(errs, err)
	}
	if b.consumeCh != nil {
		if err := b.consumeCh.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if b.publishCh != nil {
		if err := b.publishCh.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if b.conn != nil {
		if err := b.conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (b *RabbitMQTaskBus) handleDelivery(delivery amqp.Delivery) {
	var job AsyncJob
	if err := json.Unmarshal(delivery.Body, &job); err != nil {
		if b.log != nil {
			b.log.Errorw("Discarding invalid rabbitmq job payload", "error", err)
		}
		_ = delivery.Reject(false)
		return
	}

	b.mu.RLock()
	handler, ok := b.handlers[job.Type]
	b.mu.RUnlock()
	if !ok {
		if b.log != nil {
			b.log.Errorw("Discarding rabbitmq job without handler", "job_type", job.Type)
		}
		_ = delivery.Reject(false)
		return
	}

	if err := handler(context.Background(), job); err != nil {
		if b.log != nil {
			b.log.Errorw("Rabbitmq job handler failed", "job_type", job.Type, "error", err)
		}
		_ = delivery.Reject(false)
		return
	}

	_ = delivery.Ack(false)
}

func normalizeQueuePrefix(prefix string) string {
	if prefix == "" {
		return "drama"
	}
	return prefix
}
