package cacher

type CacheController interface {
	GetDataFromTable(out chan interface{}, key string)
	SetDataToTable(out chan interface{}, data []byte)
	ClearCache() interface{}
}
