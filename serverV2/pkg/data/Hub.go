package data

type Hub struct {
	rooms map[string]*Room
}

// TODO:
// <- userJoin
// <- userLeft

func NewHub() *Hub {
	return &Hub{}
}