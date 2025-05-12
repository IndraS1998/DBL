package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func AppendRedisPayload(p Payload) error {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	wrapped, err := WrapPayload(p)
	if err != nil {
		return err
	}
	data, _ := json.Marshal(wrapped)
	if err := rdb.RPush(ctx, "payloads", data).Err(); err != nil {
		return err
	}
	fmt.Printf("stored payload \n")
	return nil
}

func RetreivePayloads() ([]Payload, error) {
	var payloads []Payload
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Retrieve payloads from Redis
	results, err := rdb.LRange(ctx, "payloads", 0, -1).Result()
	if err != nil {
		return nil, err
	}

	fmt.Println("\nRetrieved from Redis:")
	for _, entry := range results {
		payload, err := UnwrapPayload(entry)
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, payload)
	}
	return payloads, nil
}

func ClearPayloads() error {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if err := rdb.Del(ctx, "payloads").Err(); err != nil {
		return err
	}
	fmt.Println("cleared all payloads from Redis")
	return nil
}
