package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
	err = bson.Unmarshal(rawValue, &newModel)
	if err != nil {
		return
	}

	newValue = newModel[fieldName].(int64) + increment

	if err != nil {
		c.DebugLog(fmt.Sprintf(logErrorLine, "error", *collectionName, err, model))
	}

	return
}

// getWithMongo will get a given struct from MongoDB
func (c *Client) getWithMongo(
	ctx context.Context,
	model interface{},
	conditions map[string]interface{},
) error {
	queryConditions := getMongoQueryConditions(model, conditions)
	collectionName := utils.GetModelTableName(model)
	if collectionName == nil {
		return ErrUnknownCollection
	}

	// Set the collection
	collection := c.options.mongoDB.Collection(
		setPrefix(c.options.mongoDBConfig.TablePrefix, *collectionName),
	)

	if utils.IsModelSlice(model) {
		c.DebugLog(fmt.Sprintf(logLine, "findMany", *collectionName, queryConditions))

		cursor, err := collection.Find(ctx, queryConditions)
		if err != nil {
			return err
		}
		if err = cursor.Err(); errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNoResults
		} else if err != nil {
			return cursor.Err()
		}

		if err = cursor.All(ctx, model); err != nil {
			return err
		}
	} else {
		c.DebugLog(fmt.Sprintf(logLine, "find", *collectionName, queryConditions))

		result := collection.FindOne(ctx, queryConditions)
		if err := result.Err(); errors.Is(err, mongo.ErrNoDocuments) {
			c.DebugLog(fmt.Sprintf(logLine, "result", *collectionName, "no result"))
			return ErrNoResults
		} else if err != nil {
			c.DebugLog(fmt.Sprintf(logLine, "result err", *collectionName, err))
			return result.Err()
		}

		if err := result.Decode(model); err != nil {
			c.DebugLog(fmt.Sprintf(logLine, "result err", *collectionName, err))
			return err
		}
	}

	return nil
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
				a, err := json.Marshal((*conditions)["$and"])
				if err != nil {
					// todo: How to handle the error?
					continue
				}
				if err = json.Unmarshal(a, &and); err != nil {
					// todo: How to handle the error?
					continue
				}
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
				a, err := json.Marshal((*conditions)["$or"])
				if err != nil {
					// todo: How to handle the error?
					continue
				}
				if err = json.Unmarshal(a, &or); err != nil {
					// todo: How to handle the error?
					continue
				}
			}
			for _, c := range or {
				processMongoConditions(&c) // nolint: scopelint,gosec // ignore for now
			}
		}
	}
}

func processXpubMetadataConditions(conditions *map[string]interface{}) {
	// marshal / unmarshal into standard map[string]interface{}
	m, err := json.Marshal((*conditions)["xpub_metadata"])
	if err != nil {
		// todo: How to handle the error?
		return
	}
	var r map[string]interface{}
	if err = json.Unmarshal(m, &r); err != nil {
		// todo: How to handle the error?
		return
	}

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
	m, err := json.Marshal((*conditions)["xpub_output_value"])
	if err != nil {
		// todo: How to handle the error?
		return
	}
	var r map[string]interface{}
	if err = json.Unmarshal(m, &r); err != nil {
		// todo: How to handle the error?
		return
	}

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
	m, err := json.Marshal((*conditions)["metadata"])
	if err != nil {
		// todo: How to handle the error?
		return
	}
	var r map[string]interface{}
	if err = json.Unmarshal(m, &r); err != nil {
		// todo: How to handle the error?
		return
	}

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
