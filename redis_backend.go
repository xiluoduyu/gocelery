// Copyright (c) 2019 Sick Yoon
// This file is part of gocelery which is released under MIT license.
// See file LICENSE for full license details.

package gocelery

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

const defaultTTL = time.Hour * 24

// RedisCeleryBackend is celery backend for redis
type RedisCeleryBackend struct {
	*redis.Pool
	TTL time.Duration
}

// NewRedisBackend creates new RedisCeleryBackend with given redis pool.
// RedisCeleryBackend can be initialized manually as well.
func NewRedisBackend(conn *redis.Pool) *RedisCeleryBackend {
	return &RedisCeleryBackend{
		Pool: conn,
		TTL:  defaultTTL,
	}
}

// NewRedisCeleryBackend creates new RedisCeleryBackend
//
// Deprecated: NewRedisCeleryBackend exists for historical compatibility
// and should not be used. Pool should be initialized outside of gocelery package.
func NewRedisCeleryBackend(uri string) *RedisCeleryBackend {
	return &RedisCeleryBackend{
		Pool: NewRedisPool(uri),
		TTL:  defaultTTL,
	}
}

// GetResult queries redis backend to get asynchronous result
func (cb *RedisCeleryBackend) GetResult(taskID string) (*ResultMessage, error) {
	conn := cb.Get()
	defer conn.Close()
	val, err := conn.Do("GET", fmt.Sprintf("celery-task-meta-%s", taskID))
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, fmt.Errorf("result not available")
	}
	var resultMessage ResultMessage
	err = json.Unmarshal(val.([]byte), &resultMessage)
	if err != nil {
		return nil, err
	}
	return &resultMessage, nil
}

// SetResult pushes result back into redis backend
func (cb *RedisCeleryBackend) SetResult(taskID string, result *ResultMessage) error {
	resBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}
	conn := cb.Get()
	defer conn.Close()

	ttl := defaultTTL
	if cb.TTL > 0 {
		ttl = cb.TTL
	}
	_, err = conn.Do("SETEX", fmt.Sprintf("celery-task-meta-%s", taskID), int(ttl.Seconds()), resBytes)
	return err
}
