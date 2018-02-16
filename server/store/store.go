package store

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/dave/jsgo/config"
)

type CompileData struct {
	Path    string
	Time    time.Time
	Success bool
	Error   string
	Min     CompileContents
	Max     CompileContents
	Ip      string
}

type CompileContents struct {
	Main     string
	Prelude  string
	Packages []CompilePackage
}

type CompilePackage struct {
	Path     string
	Standard bool
	Hash     string
}

func Save(ctx context.Context, path string, data CompileData) error {
	client, err := datastore.NewClient(ctx, config.ProjectID)
	if err != nil {
		return err
	}
	defer client.Close()
	if _, err := client.Put(ctx, compileKey(), &data); err != nil {
		return err
	}
	if data.Success {
		if _, err := client.Put(ctx, packageKey(path), &data); err != nil {
			return err
		}
	}
	return nil
}

func Lookup(ctx context.Context, path string) (bool, CompileData, error) {
	client, err := datastore.NewClient(ctx, config.ProjectID)
	if err != nil {
		return false, CompileData{}, err
	}
	defer client.Close()
	var data CompileData

	/*q := datastore.NewQuery("Compile").Filter("Success =", true).Filter("Path =", path).Order("-Time").Limit(1)

	if _, err := client.Run(ctx, q).Next(&data); err != nil {
		if err == iterator.Done {
			return false, CompileData{}, nil
		}
		return false, CompileData{}, err
	}*/

	if err := client.Get(ctx, packageKey(path), &data); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return false, CompileData{}, nil
		}
		return false, CompileData{}, err
	}
	return true, data, nil
}

func compileKey() *datastore.Key {
	return datastore.IncompleteKey("Compile", nil)
}

func packageKey(path string) *datastore.Key {
	return datastore.NameKey("Package", path, nil)
}
