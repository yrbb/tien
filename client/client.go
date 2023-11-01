package client

import (
	"bytes"
	"context"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Callback func(*Response) error

type Config = clientv3.Config

type Client interface {
	Client() *clientv3.Client
	Watch(env, project, key string, prefix bool, callback Callback) error
	Get(env, project, key string, prefix bool) ([]*Response, error)
	Close() error
}

func NewClient(config Config) (Client, error) {
	cli, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	if config.Logger == nil {
		config.Logger = NewLogger()
	}

	return &client{
		client:  cli,
		watcher: NewWatcher(cli),
	}, nil
}

type client struct {
	client  *clientv3.Client
	watcher Watcher
}

func (c *client) Client() *clientv3.Client {
	return c.client
}

func (c *client) Watch(env, project, key string, prefix bool, fn Callback) error {
	if env == "" || project == "" {
		return fmt.Errorf("env or project is empty")
	}

	if key == "" {
		prefix = true
	}

	pBts := []byte(fmt.Sprintf("/%s/%s/", env, project))
	path := fmt.Sprintf("/%s/%s/%s", env, project, key)

	return c.watcher.Watch(path, prefix, func(e *clientv3.Event) error {
		keyName := bytes.TrimPrefix(e.Kv.Key, pBts)

		return fn(&Response{
			Event:   e.Type.String(),
			Project: project,
			Key:     keyName,
			Value:   e.Kv.Value,
		})
	})
}

func (c *client) Get(env, project, key string, prefix bool) ([]*Response, error) {
	if env == "" || project == "" {
		return nil, fmt.Errorf("env or project is empty")
	}

	if key == "" {
		prefix = true
	}

	var ops []clientv3.OpOption
	if prefix {
		ops = append(ops, clientv3.WithPrefix())
	}

	pBts := []byte(fmt.Sprintf("/%s/%s/", env, project))
	path := fmt.Sprintf("/%s/%s/%s", env, project, key)

	resp, err := c.client.Get(context.Background(), path, ops...)
	if err != nil {
		return nil, err
	}

	res := make([]*Response, resp.Count)
	for i, kv := range resp.Kvs {
		keyName := bytes.TrimPrefix(kv.Key, pBts)

		res[i] = &Response{
			Event:   clientv3.EventTypePut.String(),
			Project: project,
			Key:     keyName,
			Value:   kv.Value,
		}
	}

	return res, nil
}

func (c *client) Close() error {
	c.watcher.Close()
	return c.client.Close()
}
