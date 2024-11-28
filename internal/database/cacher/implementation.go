package cacher

// Интерфейс для работы с хранилищем кэша
type CacheController interface {
	GetDataFromTable(out chan interface{}, key string)
	SetDataToTable(out chan interface{}, data []byte)
	ClearCache() interface{}
	LoadOrdersToCache(data []byte)
}
