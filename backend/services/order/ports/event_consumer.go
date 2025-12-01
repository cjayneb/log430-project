package ports

type EventConsumer interface {
	Start(topic string)
}
