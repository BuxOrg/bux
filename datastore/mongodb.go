package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

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

// CreateInBatchesMongo insert multiple models vai bulk.Write
func (c *Client) CreateInBatchesMongo(
	ctx context.Context,
	models interface{},
	batchSize int,
) error {

	collectionName := utils.GetModelTableName(models)
	if collectionName == nil {
		return ErrUnknownCollection
	}

	mongoModels := make([]mongo.WriteModel, 0)
	collection := c.GetMongoCollection(*collectionName)
	bulkOptions := options.BulkWrite().SetOrdered(true)
	count := 0

	switch reflect.TypeOf(models).Kind() { //nolint:exhaustive // we only get slices
	case reflect.Slice:
		s := reflect.ValueOf(models)
		for i := 0; i < s.Len(); i++ {
			m := mongo.NewInsertOneModel()
			m.SetDocument(s.Index(i).Interface())
			mongoModels = append(mongoModels, m)
			count++

			if count%batchSize == 0 {
				_, err := collection.BulkWrite(ctx, mongoModels, bulkOptions)
				if err != nil {
					return err
				}
				// reset the bulk
				mongoModels = make([]mongo.WriteModel, 0)
			}
		}
	}

	if count%batchSize != 0 {
		_, err := collection.BulkWrite(ctx, mongoModels, bulkOptions)
		if err != nil {
			return err
		}
	}

	return nil
}

// getWithMongo will get given struct(s) from MongoDB
func (c *Client) getWithMongo(
	ctx context.Context,
	models interface{},
	conditions map[string]interface{},
	fieldResult interface{},
	queryParams *QueryParams,
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

		if queryParams.Page > 0 {
			opts = append(opts, options.Find().SetLimit(int64(queryParams.PageSize)).SetSkip(int64(queryParams.PageSize*(queryParams.Page-1))))
		}

		if queryParams.OrderByField == "id" {
			queryParams.OrderByField = "_id" // use Mongo _id instead of default id field
		}
		if queryParams.OrderByField != "" {
			sortOrder := 1
			if queryParams.SortDirection == "desc" {
				sortOrder = -1
			}
			opts = append(opts, options.Find().SetSort(bson.D{{Key: queryParams.OrderByField, Value: sortOrder}}))
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

// countWithMongo will get a count of all models matching the conditions
func (c *Client) countWithMongo(
	ctx context.Context,
	models interface{},
	conditions map[string]interface{},
) (int64, error) {
	queryConditions := getMongoQueryConditions(models, conditions)
	collectionName := utils.GetModelTableName(models)
	if collectionName == nil {
		return 0, ErrUnknownCollection
	}

	// Set the collection
	collection := c.options.mongoDB.Collection(
		setPrefix(c.options.mongoDBConfig.TablePrefix, *collectionName),
	)

	c.DebugLog(fmt.Sprintf(logLine, "count", *collectionName, queryConditions))

	count, err := collection.CountDocuments(ctx, queryConditions)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// aggregateWithMongo will get a count of all models aggregate by aggregateColumn matching the conditions
func (c *Client) aggregateWithMongo(
	ctx context.Context,
	models interface{},
	conditions map[string]interface{},
	aggregateColumn string,
	timeout time.Duration,
) (map[string]interface{}, error) {
	queryConditions := getMongoQueryConditions(models, conditions)
	collectionName := utils.GetModelTableName(models)
	if collectionName == nil {
		return nil, ErrUnknownCollection
	}

	// Set the collection
	collection := c.options.mongoDB.Collection(
		setPrefix(c.options.mongoDBConfig.TablePrefix, *collectionName),
	)

	c.DebugLog(fmt.Sprintf(logLine, "count", *collectionName, queryConditions))

	var matchStage bson.D
	data, err := bson.Marshal(queryConditions)
	if err != nil {
		return nil, err
	}
	err = bson.Unmarshal(data, &matchStage)
	if err != nil {
		return nil, err
	}

	aggregateOn := bson.E{
		Key:   "_id",
		Value: "$" + aggregateColumn,
	} // default
	if aggregateColumn == "created_at" {
		aggregateOn = bson.E{
			Key: "_id",
			Value: bson.D{{
				Key: "$dateToString",
				Value: bson.D{
					{Key: "format", Value: "%Y%m%d"},
					{Key: "date", Value: "$created_at"},
				}},
			},
		}
	}

	groupStage := bson.D{{Key: "$group", Value: bson.D{
		aggregateOn, {
			Key: "count",
			Value: bson.D{{
				Key:   "$sum",
				Value: 1,
			}},
		},
	}}}

	pipeline := mongo.Pipeline{bson.D{{Key: "$match", Value: matchStage}}, groupStage}

	// anonymous struct for unmarshalling result bson
	var results []struct {
		ID    string `bson:"_id"`
		Count int64  `bson:"count"`
	}

	var aggregateCursor *mongo.Cursor
	aggregateCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	aggregateCursor, err = collection.Aggregate(aggregateCtx, pipeline)
	if err != nil {
		return nil, err
	}

	err = aggregateCursor.All(ctx, &results)
	if err != nil {
		return nil, err
	}

	aggregateResult := make(map[string]interface{})
	for _, result := range results {
		aggregateResult[result.ID] = result.Count
	}

	return aggregateResult, nil
}

// GetMongoCollection will get the mongo collection for the given tableName
func (c *Client) GetMongoCollection(
	collectionName string,
) *mongo.Collection {
	return c.options.mongoDB.Collection(
		setPrefix(c.options.mongoDBConfig.TablePrefix, collectionName),
	)
}

// GetMongoCollectionByTableName will get the mongo collection for the given tableName
func (c *Client) GetMongoCollectionByTableName(
	tableName string,
) *mongo.Collection {
	return c.options.mongoDB.Collection(tableName)
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

func processMongoConditions(conditions *map[string]interface{}) *map[string]interface{} {

	// transform the id field to mongo _id field
	_, ok := (*conditions)["id"]
	if ok {
		(*conditions)["_id"] = (*conditions)["id"]
		delete(*conditions, "id")
	}

	// transform the map of metadata to key / value query
	_, ok = (*conditions)["metadata"]
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

	for key, condition := range *conditions {
		if key == "$and" || key == "$or" {
			var slice []map[string]interface{}
			a, _ := json.Marshal(condition) // nolint: errchkjson // this check might break the current code
			_ = json.Unmarshal(a, &slice)
			var newConditions []map[string]interface{}
			for _, c := range slice {
				newConditions = append(newConditions, *processMongoConditions(&c)) // nolint: scopelint,gosec // ignore for now
			}
			(*conditions)[key] = newConditions
		}
	}

	return conditions
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

	return map[string][]mongo.IndexModel{
		"block_headers": {
			mongo.IndexModel{Keys: bsonx.Doc{{
				Key:   "height",
				Value: bsonx.Int32(1),
			}}},
			mongo.IndexModel{Keys: bsonx.Doc{{
				Key:   "synced",
				Value: bsonx.Int32(1),
			}}},
		},
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
