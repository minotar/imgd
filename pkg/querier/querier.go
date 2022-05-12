package querier

type Querier struct {
}

func NewQuerier(storage interface{}) *Querier {
	querier := &Querier{}
	return querier
}
