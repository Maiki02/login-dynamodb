package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

// GetDBClient devuelve la instancia del cliente de MongoDB.
// Esta es la función que usarán las otras capas para obtener la conexión.
func GetDBClient() *mongo.Client {
	return Client
}

func ConnectMongoDB() {
	uri := os.Getenv("MONGO_URI")
	clientOptions := options.Client().ApplyURI(uri)

	var err error
	Client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Verifica la conexión
	err = Client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	log.Println("Connected to MongoDB!")
}

func DisconnectMongoDB() {
	if err := Client.Disconnect(context.TODO()); err != nil {
		log.Fatal(err)
	}
	log.Println("Disconnected from MongoDB!")
}

func InsertDocument(nameDB string, collectionName string, document interface{}) (*mongo.InsertOneResult, error) {
	collection := Client.Database(nameDB).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func UpdateDocumentByCode(nameDB, collectionName string, code int32, updates map[string]interface{}) error {
	collection := Client.Database(nameDB).Collection(collectionName)
	filter := bson.M{"code": code}
	update := bson.M{"$set": updates}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

func UpdateDocumentByID(nameDB, collectionName string, id string, updates map[string]interface{}) error {
	collection := Client.Database(nameDB).Collection(collectionName)

	// Convertir id de string a ObjectID
	objectID, err2 := primitive.ObjectIDFromHex(id)
	if err2 != nil {
		return err2
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": updates}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

// Actualiza el documento y lo retorna
func UpdateAndFindDocumentByID(nameDB, collectionName string, id string, updates map[string]interface{}) *mongo.SingleResult {
	collection := Client.Database(nameDB).Collection(collectionName)

	objectID, _ := primitive.ObjectIDFromHex(id) // El error se maneja en el .Decode()

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": updates}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	return collection.FindOneAndUpdate(context.Background(), filter, update, opts)
}

func FindDocuments(nameDB, collectionName string, filter bson.M) (*mongo.Cursor, error) {
	collection := Client.Database(nameDB).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	return cursor, nil
}

func GetLastCode(nameDB, collectionName string) (int32, error) {
	collection := Client.Database(nameDB).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	sort := bson.M{"code": -1}
	opts := options.FindOne().SetSort(sort)

	var result bson.M
	err := collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		return 0, err
	}

	code := result["code"].(int32)
	return code, nil
}

// DeleteDocumentByID elimina un documento basado en su _id.
func DeleteDocumentByID(nameDB, collectionName string, id string) error {
	collection := Client.Database(nameDB).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Convertir el ID de string a ObjectID de MongoDB
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// Retorna un error si el string del ID no es válido
		return err
	}

	filter := bson.M{"_id": objectID}

	// Ejecuta la operación de borrado en la base de datos
	_, err = collection.DeleteOne(ctx, filter)
	return err
}
