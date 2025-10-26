package ports

type MatchingEngine interface {
	SubmitOrder(orderId int) error
}