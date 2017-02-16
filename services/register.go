package services

// Registory ...
type Registory struct {
	list []string
	m    map[string]Service
}

// NewRegistory ...
func NewRegistory() *Registory {
	return &Registory{
		m: make(map[string]Service),
	}
}

// Register ...
func (r *Registory) Register(s Service) error {
	r.list = append(r.list, s.Name())
	r.m[s.Name()] = s
	return nil
}

// RegisterName ...
func (r *Registory) RegisterName(name string, s Service) error {
	r.list = append(r.list, name)
	r.m[name] = s
	return nil
}

// List ...
func (r *Registory) List() []string {
	return r.list
}
