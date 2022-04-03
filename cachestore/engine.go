package cachestore

// Engine is the different engines that are supported (key->value)
type Engine string

// Supported engines
const (
	Empty     Engine = "empty"
	FreeCache Engine = "freecache"
	Redis     Engine = "redis"
)

// String is the string version of engine
func (e Engine) String() string {
	return string(e)
}

// IsEmpty will return true if the engine is not set
func (e Engine) IsEmpty() bool {
	return e == Empty
}
