package nats_infra

import (
	"context"
	"log/slog"
	"time"

	core "goodin/internal"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var _ (NatsInfrastructure) = (*NatsInfrastructureImpl)(nil)

func NewNatsInfrastructureImpl(core core.App) *NatsInfrastructureImpl {

	return &NatsInfrastructureImpl{
		app: core,
	}
}

type NatsInfrastructureImpl struct {
	app       core.App
	natsCore  *nats.Conn
	jetstream jetstream.JetStream
}

func (n *NatsInfrastructureImpl) CreateConsumer(ctx context.Context, stream string, cfg jetstream.ConsumerConfig) (jetstream.Consumer, error) {
	return n.jetstream.CreateConsumer(ctx, stream, cfg)
}

// DeleteConsumer implements NatsInfrastructure.
func (n *NatsInfrastructureImpl) DeleteConsumer(ctx context.Context, stream string, consumer string) error {
	return n.jetstream.DeleteConsumer(ctx, stream, consumer)
}

// OrderedConsumer implements NatsInfrastructure.
func (n *NatsInfrastructureImpl) OrderedConsumer(ctx context.Context, stream string, cfg jetstream.OrderedConsumerConfig) (jetstream.Consumer, error) {
	return n.jetstream.OrderedConsumer(ctx, stream, cfg)
}

// Publish implements NatsInfrastructure.
func (n *NatsInfrastructureImpl) Publish(ctx context.Context, subject string, payload []byte, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
	return n.jetstream.Publish(ctx, subject, payload, opts...)
}

// PublishAsync implements NatsInfrastructure.
func (n *NatsInfrastructureImpl) PublishAsync(subject string, payload []byte, opts ...jetstream.PublishOpt) (jetstream.PubAckFuture, error) {
	return n.jetstream.PublishAsync(subject, payload, opts...)
}

// PublishAsyncComplete implements NatsInfrastructure.
func (n *NatsInfrastructureImpl) PublishAsyncComplete() <-chan struct{} {
	return n.jetstream.PublishAsyncComplete()
}

// PublishAsyncPending implements NatsInfrastructure.
func (n *NatsInfrastructureImpl) PublishAsyncPending() int {
	return n.jetstream.PublishAsyncPending()
}

// PublishMsg implements NatsInfrastructure.
func (n *NatsInfrastructureImpl) PublishMsg(ctx context.Context, msg *nats.Msg, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error) {
	return n.jetstream.PublishMsg(ctx, msg, opts...)
}

// PublishMsgAsync implements NatsInfrastructure.
func (n *NatsInfrastructureImpl) PublishMsgAsync(msg *nats.Msg, opts ...jetstream.PublishOpt) (jetstream.PubAckFuture, error) {
	return n.jetstream.PublishMsgAsync(msg, opts...)
}

// CreateOrUpdateConsumer implements NatsInfrastructure.
func (n *NatsInfrastructureImpl) CreateOrUpdateConsumer(ctx context.Context, stream string, cfg jetstream.ConsumerConfig) (jetstream.Consumer, error) {
	return n.jetstream.CreateOrUpdateConsumer(ctx, stream, cfg)
}

// CreateOrUpdateKeyValue implements NatsInfrastructure.
func (n *NatsInfrastructureImpl) CreateOrUpdateKeyValue(ctx context.Context, cfg jetstream.KeyValueConfig) (jetstream.KeyValue, error) {
	return n.jetstream.CreateOrUpdateKeyValue(ctx, cfg)
}

// CreateOrUpdateObjectStore implements NatsInfrastructure.
func (n *NatsInfrastructureImpl) CreateOrUpdateObjectStore(ctx context.Context, cfg jetstream.ObjectStoreConfig) (jetstream.ObjectStore, error) {
	return n.jetstream.CreateOrUpdateObjectStore(ctx, cfg)
}

// CreateOrUpdateStream implements NatsInfrastructure.
func (n *NatsInfrastructureImpl) CreateOrUpdateStream(ctx context.Context, cfg jetstream.StreamConfig) (jetstream.Stream, error) {
	return n.jetstream.CreateOrUpdateStream(ctx, cfg)
}

func (n *NatsInfrastructureImpl) Request(subj string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	return n.natsCore.Request(subj, data, timeout)
}

func (n *NatsInfrastructureImpl) Subscribe(subj string, cb nats.MsgHandler) (*nats.Subscription, error) {
	return n.natsCore.Subscribe(subj, cb)
}

// Close implements driver.IDriver.
func (n *NatsInfrastructureImpl) Close() error {
	if n.natsCore == nil {
		return nil
	}

	if n.jetstream == nil {
		return nil
	}

	if err := n.natsCore.Drain(); err != nil {
		n.app.Logger().Error("cannot drain nats connection", slog.Any("msg", err.Error()))
		return err
	}
	n.natsCore.Close()

	n.app.Logger().Info("Core Nats & Jetstream connection closed")

	return nil
}

// Init implements driver.IDriver.
func (n *NatsInfrastructureImpl) Init() error {

	n.app.Logger().Debug(nats.DefaultURL)

	nc, err := nats.Connect(
		nats.DefaultURL,
		nats.NoReconnect(),
	)

	if err != nil {
		n.app.Logger().Error("Cannot connect to nats core", slog.Any("msg", err.Error()))
		return err
	}

	if nc.IsConnected() {
		nc.Flush()
		n.app.Logger().Info("Nats core is all set and connected!")
		js, err := jetstream.New(nc)
		if err != nil {
			n.app.Logger().Error("can't connect to jetstrea", slog.Any("msg", err.Error()))
			return err
		}
		n.app.Logger().Info("Jetstream connected")
		n.jetstream = js
	}

	n.natsCore = nc

	return nil
}
