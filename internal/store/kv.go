package store

import (
	"bytes"
	"encoding/json"
	"fmt"

	"go.etcd.io/bbolt"
)

func putJSON(tx *bbolt.Tx, bucket []byte, key string, value any) error {
	b := tx.Bucket(bucket)
	if b == nil {
		return fmt.Errorf("bucket %s missing", bucket)
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return b.Put([]byte(key), payload)
}

func getJSON(tx *bbolt.Tx, bucket []byte, key string, target any) error {
	b := tx.Bucket(bucket)
	if b == nil {
		return fmt.Errorf("bucket %s missing", bucket)
	}
	payload := b.Get([]byte(key))
	if payload == nil {
		return fmt.Errorf("%s/%s not found", bucket, key)
	}
	return json.Unmarshal(payload, target)
}

func deleteKey(tx *bbolt.Tx, bucket []byte, key string) error {
	b := tx.Bucket(bucket)
	if b == nil {
		return fmt.Errorf("bucket %s missing", bucket)
	}
	return b.Delete([]byte(key))
}

func scanJSON[T any](tx *bbolt.Tx, bucket []byte, fn func(string, T) bool) error {
	b := tx.Bucket(bucket)
	if b == nil {
		return fmt.Errorf("bucket %s missing", bucket)
	}
	return b.ForEach(func(k, v []byte) error {
		var item T
		if err := json.Unmarshal(v, &item); err != nil {
			return err
		}
		if !fn(string(bytes.Clone(k)), item) {
			return errStopScan
		}
		return nil
	})
}

var errStopScan = fmt.Errorf("stop scan")
