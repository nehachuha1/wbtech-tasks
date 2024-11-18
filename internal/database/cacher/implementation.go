package cacher

type CacheController interface {
	GetData(key string) interface{}
	SetData(key string, value interface{})
}
