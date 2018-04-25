package cachefileserver

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"path/filepath"
	"sync"
)

func New(maxTotal, maxItem uint64) *Fileserver {
	return &Fileserver{
		maxTotal: maxTotal,
		maxItem:  maxItem,
	}
}

type Fileserver struct {
	maxTotal uint64           // the maximum size in bytes
	maxItem  uint64           // the maximum size of a single item in bytes
	keys     map[string]*item // items ordered by key
	ids      map[uint64]*item // items ordered by id
	total    uint64           // the current total size in bytes
	oldest   uint64           // the next id to be evicted
	newest   uint64           // the next id to be added
	m        sync.Mutex
}

type item struct {
	key  string
	id   uint64
	data []byte
}

func (f *Fileserver) Write(ctx context.Context, bucket, name string, reader io.Reader, overwrite bool, contentType, cacheControl string) (saved bool, err error) {

	key := filepath.Join(bucket, name)

	if f.keys == nil {
		f.keys = map[string]*item{}
	}
	if f.ids == nil {
		f.ids = map[uint64]*item{}
	}

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return false, err
	}

	var length = uint64(len(b))

	if length > f.maxItem {
		// double check the actual size of the bytes
		return false, nil
	}

	f.m.Lock()
	defer f.m.Unlock()

	f.newest++
	newid := f.newest

	if i, ok := f.keys[key]; ok {
		// if the item already exist, update the id so it's not evicted

		// total after updating this item will be total + new length - old length.
		for f.total+length-uint64(len(i.data)) > f.maxTotal {
			f.evictOldest()
		}

		delete(f.ids, i.id)
		i.id = newid
		i.data = b
		f.ids[newid] = i

	} else {

		// total after adding a new item is total + length
		for f.total+length > f.maxTotal {
			f.evictOldest()
		}

		i := &item{key: key, id: newid, data: b}
		f.keys[key] = i
		f.ids[newid] = i
		f.total += length
	}

	return true, nil
}

// this should only run when the mutex is locked
func (f *Fileserver) evictOldest() {
	for {
		if i, ok := f.ids[f.oldest]; ok {
			f.total -= uint64(len(i.data))
			delete(f.ids, f.oldest)
			delete(f.keys, i.key)
			f.oldest++
			break
		}
		f.oldest++
	}
}

func (f *Fileserver) Read(ctx context.Context, bucket, name string, writer io.Writer) (found bool, err error) {
	f.m.Lock()
	defer f.m.Unlock()
	key := filepath.Join(bucket, name)
	i, ok := f.keys[key]
	if !ok {
		return false, nil
	}
	// if the item already exist, update the id so it's not evicted
	f.newest++
	newid := f.newest
	delete(f.ids, i.id)
	i.id = newid
	f.ids[newid] = i
	if _, err := io.Copy(writer, bytes.NewBuffer(i.data)); err != nil {
		return false, err
	}
	return true, nil
}
