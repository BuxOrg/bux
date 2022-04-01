package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/BuxOrg/bux/utils"
	"github.com/newrelic/go-agent/v3/integrations/nrmongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

const (
	logLine      = "MONGO %s %s: %+v\n"
	logErrorLine = "MONGO %s %s: %e: %+v\n"
)

// saveWithMongo will save a given struct to MongoDB
func (c *Client) saveWithMongo(
	ctx context.Context,
	model interface{},
	newRecord bool,
) (err error) {
	collectionName := utils.GetModelTableName(model)
	if collectionName == nil {
		return ErrUnknownCollection
	}

	// Set the collection
	collection := c.options.mongoDB.Collection(
		setPrefix(c.options.mongoDBConfig.TablePrefix, *collectionName),
	)

	// Create or update
	if newRecord {
		c.DebugLog(fmt.Sprintf(logLine, "insert", *collectionName, model))
		_, err = collection.InsertOne(ctx, model)
	} else {
		id := utils.GetModelStringAttribute(model, "ID")
		update := bson.M{"$set": model}
		unset := utils.GetModelUnset(model)
		if len(unset) > 0 {
			update = bson.M{"$set": model, "$unset": unset}
		}

		c.DebugLog(fmt.Sprintf(logLine, "update", *collectionName, model))

		_, err = collection.UpdateOne(
			ctx, bson.M{"_id": *id}, update,
		)
	}

	// Check for duplicate key (insert error, record exists)
	if mongo.IsDuplicateKeyError(err) {
		c.DebugLog(fmt.Sprintf(logErrorLine, "error", *collectionName, ErrDuplicateKey, model))
		return ErrDuplicateKey
	}

	if err != nil {
		c.DebugLog(fmt.Sprintf(logErrorLine, "error", *collectionName, err, model))
	}

	return
}

// incrementWithMongo will save a given struct to MongoDB
func (c *Client) incrementWithMongo(
	ctx context.Context,
	model interface{},
	fieldName string,
	increment int64,
) (newValue int64, err error) {
	collectionName := utils.GetModelTableName(model)
	if collectionName == nil {
		return newValue, ErrUnknownCollection
	}

	// Set the collection
	collection := c.options.mongoDB.Collection(
		setPrefix(c.options.mongoDBConfig.TablePrefix, *collectionName),
	)

	id := utils.GetModelStringAttribute(model, "ID")
	if id == nil {
		return newValue, errors.New("can only increment by id")
	}
	update := bson.M{"$inc": bson.M{fieldName: increment}}

	c.DebugLog(fmt.Sprintf(logLine, "increment", *collectionName, model))

	result := collection.FindOneAndUpdate(
		ctx, bson.M{"_id": *id}, update,
	)
	if result.Err() != nil {
		return newValue, result.Err()
	}
	var rawValue bson.Raw
	if rawValue, err = result.DecodeBytes(); err != nil {
		return
	}
	var newModel map[string]interface{}
	_ = bson.Unmarshal(rawValue, &newModel) // todo: cannot check error, breaks code atm

	newValue = newModel[fieldName].(int64) + increment

	if err != nil {
		c.DebugLog(fmt.Sprintf(logErrorLine, "error", *collectionName, err, model))
	}

	return
}

// getWithMongo will get given struct(s) from MongoDB
func (c *Client) getWithMongo(
	ctx context.Context,
	models interface{},
	conditions map[string]interface{},
	fieldResult interface{},
) error {
	queryConditions := getMongoQueryConditions(models, conditions)
	collectionName := utils.GetModelTableName(models)
	if collectionName == nil {
		return ErrUnknownCollection
	}

	// Set the collection
	collection := c.options.mongoDB.Collection(
		setPrefix(c.options.mongoDBConfig.TablePrefix, *collectionName),
	)

	var fields []string
	if fieldResult != nil {
		fields = getFieldNames(fieldResult)
	}

	if utils.IsModelSlice(models) {
		c.DebugLog(fmt.Sprintf(logLine, "findMany", *collectionName, queryConditions))

		var opts []*options.FindOptions
		if fields != nil {
			projection := bson.D{}
			for _, field := range fields {
				projection = append(projection, bson.E{Key: field, Value: 1})
			}
			opts = append(opts, options.Find().SetProjection(projection))
		}

		cursor, err := collection.Find(ctx, queryConditions, opts...)
		if err != nil {
			return err
		}
		if err = cursor.Err(); errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNoResults
		} else if err != nil {
			return cursor.Err()
		}

		if fieldResult != nil {
			if err = cursor.All(ctx, fieldResult); err != nil {
				return err
			}
		} else {
			if err = cursor.All(ctx, models); err != nil {
				return err
			}
		}
	} else {
		c.DebugLog(fmt.Sprintf(logLine, "find", *collectionName, queryConditions))

		var opts []*options.FindOneOptions
		if fields != nil {
			projection := bson.D{}
			for _, field := range fields {
				projection = append(projection, bson.E{Key: field, Value: 1})
			}
			opts = append(opts, options.FindOne().SetProjection(projection))
		}

		result := collection.FindOne(ctx, queryConditions, opts...)
		if err := result.Err(); errors.Is(err, mongo.ErrNoDocuments) {
			c.DebugLog(fmt.Sprintf(logLine, "result", *collectionName, "no result"))
			return ErrNoResults
		} else if err != nil {
			c.DebugLog(fmt.Sprintf(logLine, "result err", *collectionName, err))
			return result.Err()
		}

		if fieldResult != nil {
			if err := result.Decode(fieldResult); err != nil {
				c.DebugLog(fmt.Sprintf(logLine, "result err", *collectionName, err))
				return err
			}
		} else {
			if err := result.Decode(models); err != nil {
				c.DebugLog(fmt.Sprintf(logLine, "result err", *collectionName, err))
				return err
			}
		}
	}

	return nil
}

func getFieldNames(fieldResult interface{}) []string {
	if fieldResult == nil {
		return []string{}
	}

	fields := make([]string, 0)

	model := reflect.ValueOf(fieldResult)
	if model.Kind() == reflect.Ptr {
		model = model.Elem()
	}
	if model.Kind() == reflect.Slice {
		elemType := model.Type().Elem()
		fmt.Println(elemType.Kind())
		if elemType.Kind() == reflect.Ptr {
			model = reflect.New(elemType.Elem())
		} else {
			model = reflect.New(elemType)
		}
	}
	if model.Kind() == reflect.Ptr {
		model = model.Elem()
	}

	for i := 0; i < model.Type().NumField(); i++ {
		field := model.Type().Field(i)
		fields = append(fields, field.Tag.Get("bson"))
	}

	return fields
}

// setPrefix will automatically append the table prefix if found
func setPrefix(prefix, collection string) string {
	if len(prefix) > 0 {
		return prefix + "_" + collection
	}
	return collection
}

// getMongoQueryConditions will build the Mongo query conditions
// this functions tries to mimic the way gorm generates a where clause (naively)
func getMongoQueryConditions(
	model interface{},
	conditions map[string]interface{},
) map[string]interface{} {
	if conditions == nil {
		conditions = map[string]interface{}{}
	} else {
		// check for id field
		_, ok := conditions["id"]
		if ok {
			conditions["_id"] = conditions["id"]
			delete(conditions, "id")
		}

		processMongoConditions(&conditions)
	}

	// add model ID to the query conditions, if set on the model
	id := utils.GetModelStringAttribute(model, "ID")
	if id != nil && *id != "" {
		conditions["_id"] = *id
	}

	return conditions
}

func processMongoConditions(conditions *map[string]interface{}) {

	// transform the map of metadata to key / value query
	_, ok := (*conditions)["metadata"]
	if ok {
		processMetadataConditions(conditions)
	}

	// transform the map of xpub_metadata to key / value query
	_, ok = (*conditions)["xpub_metadata"]
	if ok {
		processXpubMetadataConditions(conditions)
	}

	// transform the map of xpub_metadata to key / value query
	_, ok = (*conditions)["xpub_output_value"]
	if ok {
		processXpubOutputValueConditions(conditions)
	}

	for key := range *conditions {
		if key == "$and" {
			var and []map[string]interface{}
			if _, ok = (*conditions)["$and"].([]map[string]interface{}); ok {
				and = (*conditions)["$and"].([]map[string]interface{})
			} else {
				a, _ := json.Marshal((*conditions)["$and"]) // nolint: errchkjson // this check might break the current code
				_ = json.Unmarshal(a, &and)
			}
			for _, c := range and {
				processMongoConditions(&c) // nolint: scopelint,gosec // ignore for now
			}
		}
		if key == "$or" {
			var or []map[string]interface{}
			if _, ok = (*conditions)["$or"].([]map[string]interface{}); ok {
				or = (*conditions)["$or"].([]map[string]interface{})
			} else {
				a, _ := json.Marshal((*conditions)["$or"]) // nolint: errchkjson // this check might break the current code
				_ = json.Unmarshal(a, &or)
			}
			for _, c := range or {
				processMongoConditions(&c) // nolint: scopelint,gosec // ignore for now
			}
		}
	}
}

func processXpubMetadataConditions(conditions *map[string]interface{}) {
	// marshal / unmarshal into standard map[string]interface{}
	m, _ := json.Marshal((*conditions)["xpub_metadata"]) // nolint: errchkjson // this check might break the current code
	var r map[string]interface{}
	_ = json.Unmarshal(m, &r)

	for xPub, xr := range r {
		xPubMetadata := make([]map[string]interface{}, 0)
		for key, value := range xr.(map[string]interface{}) {
			xPubMetadata = append(xPubMetadata, map[string]interface{}{
				"xpub_metadata.x": xPub,
				"xpub_metadata.k": key,
				"xpub_metadata.v": value,
			})
		}
		if len(xPubMetadata) > 0 {
			_, ok := (*conditions)["$and"]
			if ok {
				and := (*conditions)["$and"].([]map[string]interface{})
				and = append(and, xPubMetadata...)
				(*conditions)["$and"] = and
			} else {
				(*conditions)["$and"] = xPubMetadata
			}
		}
	}
	delete(*conditions, "xpub_metadata")
}

func processXpubOutputValueConditions(conditions *map[string]interface{}) {
	m, _ := json.Marshal((*conditions)["xpub_output_value"]) // nolint: errchkjson // this check might break the current code
	var r map[string]interface{}
	_ = json.Unmarshal(m, &r)

	xPubOutputValue := make([]map[string]interface{}, 0)
	for xPub, value := range r {
		outputKey := "xpub_output_value." + xPub
		xPubOutputValue = append(xPubOutputValue, map[string]interface{}{
			outputKey: value,
		})
	}
	if len(xPubOutputValue) > 0 {
		_, ok := (*conditions)["$and"]
		if ok {
			and := (*conditions)["$and"].([]map[string]interface{})
			and = append(and, xPubOutputValue...)
			(*conditions)["$and"] = and
		} else {
			(*conditions)["$and"] = xPubOutputValue
		}
	}

	delete(*conditions, "xpub_output_value")
}

func processMetadataConditions(conditions *map[string]interface{}) {
	// marshal / unmarshal into standard map[string]interface{}
	m, _ := json.Marshal((*conditions)["metadata"]) // nolint: errchkjson // this check might break the current code
	var r map[string]interface{}
	_ = json.Unmarshal(m, &r)

	metadata := make([]map[string]interface{}, 0)
	for key, value := range r {
		metadata = append(metadata, map[string]interface{}{
			"metadata.k": key,
			"metadata.v": value,
		})
	}
	if len(metadata) > 0 {
		_, ok := (*conditions)["$and"]
		if ok {
			and := (*conditions)["$and"].([]map[string]interface{})
			and = append(and, metadata...)
			(*conditions)["$and"] = and
		} else {
			(*conditions)["$and"] = metadata
		}
	}
	delete(*conditions, "metadata")
}

// openMongoDatabase will open a new database or use an existing connection
func openMongoDatabase(ctx context.Context, config *MongoDBConfig) (*mongo.Database, error) {

	// Use an existing connection
	if config.ExistingConnection != nil {
		return config.ExistingConnection, nil
	}

	// Create the new client
	nrMon := nrmongo.NewCommandMonitor(nil)
	client, err := mongo.Connect(
		ctx,
		options.Client().SetMonitor(nrMon),
		options.Client().ApplyURI(config.URI),
	)
	if err != nil {
		return nil, err
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	// Return the client
	return client.Database(
		config.DatabaseName,
	), nil
}

// getMongoIndexes will get indexes from mongo
func getMongoIndexes() map[string][]mongo.IndexModel {

	// todo: move these to bux out of this package
	return map[string][]mongo.IndexModel{
		"destinations": {
			mongo.IndexModel{Keys: bsonx.Doc{{
				Key:   "address",
				Value: bsonx.Int32(1),
			}}},
		},
		"draft_transactions": {
			mongo.IndexModel{Keys: bsonx.Doc{{
				Key:   "status",
				Value: bsonx.Int32(1),
			}}},
			mongo.IndexModel{Keys: bsonx.Doc{{
				Key:   "xpub_id",
				Value: bsonx.Int32(1),
			}, {
				Key:   "status",
				Value: bsonx.Int32(1),
			}}},
		},
		"transactions": {
			mongo.IndexModel{Keys: bsonx.Doc{{
				Key:   "xpub_metadata.x",
				Value: bsonx.Int32(1),
			}, {
				Key:   "xpub_metadata.k",
				Value: bsonx.Int32(1),
			}, {
				Key:   "xpub_metadata.v",
				Value: bsonx.Int32(1),
			}}},
			mongo.IndexModel{Keys: bsonx.Doc{{
				Key:   "xpub_in_ids",
				Value: bsonx.Int32(1),
			}}},
			mongo.IndexModel{Keys: bsonx.Doc{{
				Key:   "xpub_out_ids",
				Value: bsonx.Int32(1),
			}}},
		},
		"utxos": {
			mongo.IndexModel{Keys: bsonx.Doc{{
				Key:   "transaction_id",
				Value: bsonx.Int32(1),
			}, {
				Key:   "output_index",
				Value: bsonx.Int32(1),
			}}},
			mongo.IndexModel{Keys: bsonx.Doc{{
				Key:   "xpub_id",
				Value: bsonx.Int32(1),
			}, {
				Key:   "type",
				Value: bsonx.Int32(1),
			}, {
				Key:   "draft_id",
				Value: bsonx.Int32(1),
			}, {
				Key:   "spending_tx_id",
				Value: bsonx.Int32(1),
			}}},
		},
	}
}
