package ports

type MatchingEngine interface {
	SubmitOrder(orderId int) error
}

type MatchineEngineImpl struct{}

func (m *MatchineEngineImpl) SubmitOrder(orderId int) error {
	return nil
}

var _ MatchingEngine = (*MatchineEngineImpl)(nil) // Ensure interface is implemented at compile time
