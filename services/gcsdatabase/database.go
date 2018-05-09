package gcsdatabase

import (
	"cloud.google.com/go/datastore"
)

func New(client *datastore.Client) *Database {
	return &Database{
		Client: client,
	}
}

type Database struct {
	*datastore.Client
}

/*
func (d *Database) Get(ctx context.Context, key *datastore.Key, dst interface{}) (err error) {
	return d.client.Get(ctx, key, dst)
}

func (d *Database) Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	return d.client.Put(ctx, key, src)
}

func (d *Database) GetMulti(ctx context.Context, keys []*datastore.Key, dst interface{}) (err error) {
	return d.client.GetMulti(ctx, keys, dst)
}

func (d *Database) PutMulti(ctx context.Context, keys []*datastore.Key, src interface{}) (_ []*datastore.Key, err error) {
	return d.client.PutMulti(ctx, keys, src)
}
*/
