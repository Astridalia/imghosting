package database

type Middleware struct {
	next Impl
}

func (m *Middleware) FindOne(object Object) (interface{}, error) {
	result, err := m.next.FindOne(object)
	return result, err
}

func (m *Middleware) Upsert(object Object, update Object) interface{} {
	result := m.next.Upsert(object, update)
	return result
}
func (m *Middleware) DeleteOne(object Object) interface{} {
	result := m.next.DeleteOne(object)
	return result
}

func (m *Middleware) Disconnect() error {
	err := m.next.Disconnect()
	return err
}
