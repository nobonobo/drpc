package services

// Service ...
type Service interface {
	Name() string
	Rcvr() interface{}
}
