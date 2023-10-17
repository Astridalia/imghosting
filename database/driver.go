package database

type Object struct {
	ID         string                 `json:"id"`
	Collection string                 `json:"collection"`
	Data       map[string]interface{} `json:"data"`
}

type Impl interface {
	FindOne(object Object) (interface{}, error)
	Upsert(object Object, update Object) interface{}
	DeleteOne(object Object) interface{}
	Disconnect() error
}
