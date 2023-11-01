package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	v3rpc "go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type WatchCallback func(*clientv3.Event) error

type Watcher interface {
	Watch(path string, prefix bool, fn WatchCallback) error
	Close()
}

type watcher struct {
	client *clientv3.Client
	list   sync.Map
	closed int32
}

func NewWatcher(client *clientv3.Client) Watcher {
	return &watcher{
		client: client,
	}
}

func (w *watcher) Watch(path string, prefix bool, fn WatchCallback) error {
	key := fmt.Sprintf("%s-%v", path, prefix)

	if _, ok := w.list.Load(key); ok {
		return fmt.Errorf("watcher already exists, path: %s, prefix: %v", path, prefix)
	}

	nw := &watch{
		client:   w.client,
		path:     path,
		prefix:   prefix,
		callback: fn,
		exitCh:   make(chan struct{}),
		respCh:   make(chan *clientv3.Event, 100),
	}

	if err := nw.get(); err != nil {
		return err
	}

	go nw.doCallback()
	go nw.watch()

	w.list.Store(key, nw)

	return nil
}

func (w *watcher) Close() {
	if atomic.CompareAndSwapInt32(&w.closed, 0, 1) {
		w.list.Range(func(_, v interface{}) bool {
			v.(*watch).close()
			return true
		})
	}
}

type watch struct {
	client   *clientv3.Client
	revision int64
	cancel   context.CancelFunc
	closed   int32
	path     string
	prefix   bool
	callback WatchCallback
	respCh   chan *clientv3.Event
	exitCh   chan struct{}
}

func (w *watch) get() error {
	var ops []clientv3.OpOption
	if w.prefix {
		ops = append(ops, clientv3.WithPrefix())
	}

	resp, err := w.client.Get(context.Background(), w.path, ops...)
	if err != nil {
		return err
	}

	if resp.Header.Revision > w.getRevision() {
		w.setRevision(resp.Header.Revision)
	}

	for _, kv := range resp.Kvs {
		if err := w.callback(&clientv3.Event{
			Type: mvccpb.PUT,
			Kv:   kv,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (w *watch) doCallback() {
	for {
		select {
		case <-w.exitCh:
			return
		case resp := <-w.respCh:
			if resp == nil {
				return
			}

			w.callback(resp)
		}
	}
}

func (w *watch) watch() {
	slog.Debug("tien client watch start", slog.String("path", w.path), slog.Bool("prefix", w.prefix))

	for {
		var ops []clientv3.OpOption
		if revision := w.getRevision(); revision == 0 {
			ops = append(ops, clientv3.WithCreatedNotify())
		} else {
			ops = append(ops, clientv3.WithRev(revision+1))
		}

		if w.prefix {
			ops = append(ops, clientv3.WithPrefix())
		}

		ctx, cancel := context.WithCancel(context.Background())
		w.cancel = cancel

		select {
		case <-w.exitCh:
			cancel()
			slog.Debug("tien client watch exit", slog.String("path", w.path), slog.Bool("prefix", w.prefix))
			return
		default:
		}

		rch := w.client.Watch(ctx, w.path, ops...)
		for resp := range rch {
			if resp.CompactRevision > w.getRevision() {
				w.setRevision(resp.CompactRevision)
			}

			if err := resp.Err(); err != nil {
				if errors.Is(err, v3rpc.ErrCompacted) {
					break
				}

				if strings.Contains(err.Error(), "etcdserver: mvcc: required revision has been compacted") {
					break
				}

				if clientv3.IsConnCanceled(err) {
					slog.Warn(
						"tien client connection is closing",
						slog.String("path", w.path),
						slog.Bool("prefix", w.prefix),
					)
					return
				}

				slog.Error(
					"tien client watch error",
					slog.String("path", w.path),
					slog.Bool("prefix", w.prefix),
					slog.String("error", err.Error()),
				)
				break
			}

			if resp.Header.Revision > w.getRevision() {
				w.setRevision(resp.Header.Revision)
			}

			if atomic.LoadInt32(&w.closed) == 1 {
				return
			}

			for _, event := range resp.Events {
				slog.Debug(
					"tien client receive event",
					slog.String("event", event.Type.String()),
					slog.String("key", string(event.Kv.Key)),
					slog.String("value", string(event.Kv.Value)),
				)

				w.respCh <- event
			}
		}

		slog.Debug(
			"tien client closed",
			slog.String("path", w.path),
			slog.Bool("prefix", w.prefix),
		)

		time.Sleep(time.Second)
	}
}

func (w *watch) getRevision() int64 {
	return atomic.LoadInt64(&w.revision)
}

func (w *watch) setRevision(rev int64) {
	atomic.StoreInt64(&w.revision, rev)
}

func (w *watch) close() {
	if atomic.CompareAndSwapInt32(&w.closed, 0, 1) {
		if w.cancel != nil {
			w.cancel()
		}

		close(w.exitCh)
		close(w.respCh)
	}
}
