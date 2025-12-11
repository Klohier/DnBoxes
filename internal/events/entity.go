package events

type Entity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (e *Entity) GetID() string {
	return e.ID
}

func (e *Entity) GetName() string {
	return e.Name
}

func (e *Entity) setID(id string) {
	e.ID = id
}

func (e *Entity) setName(name string) {
	e.Name = name
}
