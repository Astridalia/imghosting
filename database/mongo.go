package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	client   *mongo.Client
	database *mongo.Database
	db       *Impl
}

func (m *Mongo) Disconnect() error {
	return m.Disconnect()
}

func (m *Mongo) FindOne(object Object) interface{} {
	return m.database.Collection(object.Collection).FindOne(context.Background(), bson.M{"_id": object.ID})
}

func (m *Mongo) Upsert(object Object, update Object) interface{} {
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	result := m.database.Collection(object.Collection).FindOneAndUpdate(context.Background(), bson.M{"_id": object.ID}, update, opts)
	return result
}

func (m *Mongo) DeleteOne(object Object) interface{} {
	return m.database.Collection(object.Collection).FindOneAndDelete(context.Background(), bson.M{"_id": object.ID})
}
