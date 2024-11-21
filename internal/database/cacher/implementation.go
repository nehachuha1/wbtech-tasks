package cacher

type CacheController interface {
	GetDataFromTable(key string) interface{}
	SetDataToTable(key string, value []byte)
}
