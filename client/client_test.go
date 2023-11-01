package client

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestClient_Watch(t *testing.T) {
	testCases := []struct {
		name     string
		env      string
		project  string
		key      string
		prefix   bool
		callback Callback
		err      error
	}{
		{
			name:    "EmptyEnv",
			env:     "",
			project: "foo",
			key:     "bar",
			prefix:  false,
			err:     fmt.Errorf("env or project is empty"),
		},
		{
			name:    "EmptyProject",
			env:     "test",
			project: "",
			key:     "bar",
			prefix:  false,
			err:     fmt.Errorf("env or project is empty"),
		},
		{
			name:    "EmptyKey",
			env:     "test",
			project: "foo",
			key:     "",
			prefix:  true,
			err:     nil,
			callback: func(r *Response) error {
				t.Log(string(r.Key), string(r.Value))
				return nil
			},
		},
		{
			name:    "ValidInput",
			env:     "test",
			project: "foo",
			key:     "bar",
			prefix:  false,
			err:     nil,
			callback: func(r *Response) error {
				t.Log(string(r.Key), string(r.Value))
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewClient(Config{
				Endpoints: []string{"127.0.0.1:2379"},
			})
			if err != nil {
				t.Fatal(err)
			}

			err = c.Watch(tc.env, tc.project, tc.key, tc.prefix, tc.callback)
			assert.Equal(t, tc.err, err)

			c.Close()
		})
	}
}

func TestClient_Get(t *testing.T) {
	c, err := NewClient(Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	testCases := []struct {
		name     string
		env      string
		project  string
		key      string
		prefix   bool
		response []*Response
		err      error
	}{
		{
			name:    "EmptyEnv",
			env:     "",
			project: "foo",
			key:     "bar",
			prefix:  false,
			err:     fmt.Errorf("env or project is empty"),
		},
		{
			name:    "EmptyProject",
			env:     "test",
			project: "",
			key:     "bar",
			prefix:  false,
			err:     fmt.Errorf("env or project is empty"),
		},
		{
			name:    "EmptyKey",
			env:     "test",
			project: "foo",
			key:     "",
			prefix:  true,
			response: []*Response{
				{
					Event:   clientv3.EventTypePut.String(),
					Project: "foo",
					Key:     []byte("bar"),
					Value:   []byte("bar"),
				},
				{
					Event:   clientv3.EventTypePut.String(),
					Project: "foo",
					Key:     []byte("bar1"),
					Value:   []byte("bar1"),
				},
			},
			err: nil,
		},
		{
			name:    "ValidInput",
			env:     "test",
			project: "foo",
			key:     "bar",
			prefix:  false,
			response: []*Response{
				{
					Event:   clientv3.EventTypePut.String(),
					Project: "foo",
					Key:     []byte("bar"),
					Value:   []byte("bar"),
				},
			},
			err: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := c.Get(tc.env, tc.project, tc.key, tc.prefix)
			assert.Equal(t, tc.err, err)

			if err == nil {
				assert.Equal(t, len(response), len(tc.response))
				for i := 0; i < len(response); i++ {
					assert.Equal(t, response[i].Key, tc.response[i].Key)
					assert.Equal(t, response[i].Value, tc.response[i].Value)
				}
			}
		})
	}
}
