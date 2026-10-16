package core

// so for now, we should be able to insert, get and delete from our internal "store"
type Store interface {
	Insert(id string, path []Link) error
	Delete(id string) error
	Get(id string) ([]Link, error)
}
