package main

import (
	"C"
	"context"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"bytes"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/hunjixin/brightbird/mongo_plugin/log"
	"github.com/vmihailenco/msgpack/v5"

	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PluginID = "mongo"

//export FLBPluginRegister
func FLBPluginRegister(ctxPointer unsafe.Pointer) int {
	logger, err := log.New(log.OutputPlugin, PluginID)
	if err != nil {
		fmt.Printf("error initializing logger: %s\n", err)

		return output.FLB_ERROR
	}

	logger.Info("Registering plugin", nil)

	result := output.FLBPluginRegister(ctxPointer, PluginID, "Go mongo go")

	switch result {
	case output.FLB_OK:
		Set(ctxPointer, &Value{
			Logger: logger,
		})
	default:
		// nothing to do
	}

	return result
}

// (fluentbit will call this)
// ctx (context) pointer to fluentbit context (state/ c code)
//
//export FLBPluginInit
func FLBPluginInit(ctxPointer unsafe.Pointer) int {
	value, err := Get(ctxPointer)
	if err != nil {
		logger, err := log.New(log.OutputPlugin, PluginID)
		if err != nil {
			fmt.Printf("error initializing logger: %s\n", err)

			return output.FLB_ERROR
		}

		logger.Info("New logger initialized", nil)

		value.Logger = logger
	}

	value.Logger.Info("Initializing plugin", nil)

	cfg := GetConfig(ctxPointer)
	value.Config = cfg

	value.Logger.Info("Connecting to mongodb", map[string]interface{}{
		"user": cfg.Database,
	})

	client, err := mongoDriver.Connect(context.Background(), options.Client().ApplyURI(cfg.URL))
	if err != nil {
		value.Logger.Error("error connect mongo", map[string]interface{}{
			"error": err,
		})

		return output.FLB_ERROR
	}
	db := client.Database(cfg.Database)
	db.Collection("logs").Indexes().CreateMany(context.Background(), []mongoDriver.IndexModel{{Keys: "kubernetes.labels.testid"}, {Keys: "kubernetes.pod_name"}, {Keys: "time"}})
	value.Db = db
	Set(ctxPointer, value)

	msgpack.RegisterExt(0, &EventTime{})
	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	panic(errors.New("not supported call"))
}

//export FLBPluginFlushCtx
func FLBPluginFlushCtx(ctxPointer, data unsafe.Pointer, length C.int, tag *C.char) (result int) {
	value, err := Get(ctxPointer)
	if err != nil {
		fmt.Printf("error getting value: %s\n", err)

		return output.FLB_ERROR
	}

	logger := value.Logger
	ctx := log.WithLogger(context.TODO(), logger)
	msgPacks := GetBytes(data, int(length)) // Create Fluent Bit decoder
	if err := ProcessAll(ctx, msgPacks, C.GoString(tag), value.Db); err != nil {
		logger.Error("Failed to process logs", map[string]interface{}{
			"error": err,
		})

		if errors.Is(err, &ErrRetry{}) {
			return output.FLB_RETRY
		}

		return output.FLB_ERROR
	}

	// Return options:
	//
	// output.FLB_OK    = data have been processed.
	// output.FLB_ERROR = unrecoverable error, do not try this again.
	// output.FLB_RETRY = retry to flush later.
	return output.FLB_OK
}

func ProcessAll(ctx context.Context, data []byte, tag string, db *mongo.Database) error {
	// For log purpose
	startTime := time.Now()
	total := 0
	logger, err := log.GetLogger(ctx)
	if err != nil {
		return fmt.Errorf("get logger: %w", err)
	}
	logger.Info("ProcessAll", map[string]interface{}{})

	dec := msgpack.NewDecoder(bytes.NewReader(data))
	dec.SetMapDecoder(func(dec *msgpack.Decoder) (interface{}, error) {
		return dec.DecodeMap()
	})
	// Iterate Records
	logCollector := db.Collection("logs")
	var documents []interface{}
	for {
		// Extract Record
		record, err := msgPackToMap(dec)
		if err != nil {
			logger.Debug("Record parser errr", map[string]interface{}{
				"count":    total,
				"duration": time.Since(startTime),
				"err":      err,
			})
			break
		}

		total++
		record["tags"] = tag
		documents = append(documents, record)
	}
	_, err = logCollector.InsertMany(ctx, documents)
	if err != nil {
		logger.Error("Failed to save log", map[string]interface{}{
			"error": err,
		})
	}
	return nil
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}
