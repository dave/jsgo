package cachepersister

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
)

type Persister struct {
	MaxTotal uint64           // the maximum size in bytes
	MaxItem  uint64           // the maximum size of a single item in bytes
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

func (p *Persister) Save(ctx context.Context, url string, size uint64, reader io.Reader) error {

	fmt.Println("caching", url)

	if p.keys == nil {
		p.keys = map[string]*item{}
	}
	if p.ids == nil {
		p.ids = map[uint64]*item{}
	}

	if size > p.MaxItem {
		// don't cache anything over MaxItem
		fmt.Println(url, "skipped. too big", size, p.MaxItem)
		return nil
	}

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	var length = uint64(len(b))

	if length > p.MaxItem {
		fmt.Println(url, "skipped. too big", length, p.MaxItem)
		// double check the actual size of the bytes
		return nil
	}

	p.m.Lock()
	defer p.m.Unlock()

	p.newest++
	newid := p.newest

	if i, ok := p.keys[url]; ok {
		// if the item already exist, update the id so it's not evicted

		// total after updating this item will be total + new length - old length.
		for p.total+length-uint64(len(i.data)) > p.MaxTotal {
			p.evictOldest()
		}

		fmt.Println("cache updated", url)

		delete(p.ids, i.id)
		i.id = newid
		i.data = b
		p.ids[newid] = i

	} else {

		// total after adding a new item is total + length
		for p.total+length > p.MaxTotal {
			p.evictOldest()
		}

		fmt.Println("cache added", url)

		i := &item{key: url, id: newid, data: b}
		p.keys[url] = i
		p.ids[newid] = i
		p.total += length
	}

	return nil
}

// this should only run when the mutex is locked
func (p *Persister) evictOldest() {
	for {
		if i, ok := p.ids[p.oldest]; ok {
			p.total -= uint64(len(i.data))
			delete(p.ids, p.oldest)
			delete(p.keys, i.key)
			p.oldest++
			break
		}
		p.oldest++
	}
}

func (p *Persister) Load(ctx context.Context, url string, writer io.Writer) (found bool, err error) {
	fmt.Println("loading", url)
	p.m.Lock()
	defer p.m.Unlock()
	if i, ok := p.keys[url]; ok {
		// if the item already exist, update the id so it's not evicted
		p.newest++
		newid := p.newest
		delete(p.ids, i.id)
		i.id = newid
		p.ids[newid] = i
		if _, err := io.Copy(writer, bytes.NewBuffer(i.data)); err != nil {
			return false, err
		}
		fmt.Println("Cache hit", url)
		return true, nil
	}
	fmt.Println("can't find", url)
	return false, nil
}
