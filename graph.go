package wfe

type Graph interface {
	Commit(response *Response) error
}

type GraphBackend interface {
	Graph(id string, request Request) (Graph, error)
}

type noopGrapher struct{}
type noopGraph struct{}

func (g *noopGrapher) Graph(id string, request Request) (Graph, error) {
	return (*noopGraph)(nil), nil
}

func (g *noopGraph) Commit(response *Response) error {
	return nil
}
