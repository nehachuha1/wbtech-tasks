package database

import "math/rand"

// Генерация внутренних айди, когда мы декомпозируем входящий JSON на несколько сущностей
func GenerateNewID() string {
	b := make([]rune, 16)
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
