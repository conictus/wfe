package wfe

type Result interface {
	ID() string
	Get() Result
}
