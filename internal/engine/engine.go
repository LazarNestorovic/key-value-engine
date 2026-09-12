package engine

type DB interface {
	Put(key string, value []byte) error
	Get(key string) ([]byte, bool, error)
	Delete(key string) error
	Close() error
}
