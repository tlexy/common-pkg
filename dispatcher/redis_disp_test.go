package dispatcher

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestRedsiConnection(t *testing.T) {
	addr := "localhost:6379"
	passwd := ""
	db := 0
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       db,
	})
	_, err := rdb.Ping(context.TODO()).Result()
	if err != nil {
		t.Fatalf("ping redis failed, err: %v", err)
	}
}

func TestDispatch(t *testing.T) {
	addr := "localhost:6379"
	passwd := ""
	db := 0
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       db,
	})
	dispatcher := NewRedisDispatcher(rdb, "test_list", "test_set")
	err := dispatcher.Dispatch(context.TODO(), "test_type", "test_id2")
	if err != nil {
		t.Fatalf("dispatch failed, err: %v", err)
	}
	exist, err := dispatcher.ExistTask(context.TODO(), "test_type", "test_id")
	if err != nil {
		t.Fatalf("exist task failed, err: %v", err)
	}
	if !exist {
		t.Fatalf("exist task failed, expect: true, actual: %v", exist)
	}
}

func TestGetTaskId(t *testing.T) {
	addr := "localhost:6379"
	passwd := ""
	db := 0
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       db,
	})
	dispatcher := NewRedisDispatcher(rdb, "test_list", "test_set")
	idStr, err := dispatcher.GetTask(context.TODO(), "test_type")
	if err != nil {
		t.Fatalf("get task id failed, err: %v", err)
	}
	if len(idStr) != 0 {
		t.Fatalf("get task id failed, expect: 0, actual: %s", idStr)
	}
}
