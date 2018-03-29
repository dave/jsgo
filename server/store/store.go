package store

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/dave/jsgo/config"
)

type Error struct {
	Time  time.Time
	Error string
	Ip    string
}

type ShareData struct {
	Time  time.Time
	Hash  string
	Files int
	Ip    string
}

type CompileData struct {
	Path string
	Time time.Time
	Min  CompileContents
	Max  CompileContents
	Ip   string

	Success bool
	Error   string
}

type DeployData struct {
	Time     time.Time
	Contents DeployContents
	Minify   bool
	Ip       string
}

type CompileContents struct {
	Main     string
	Packages []CompilePackage
}

type DeployContents struct {
	Index    string
	Main     string
	Packages []CompilePackage
}

type CompilePackage struct {
	Path     string
	Hash     string
	Standard bool
}

func StoreError(ctx context.Context, data Error) error {
	client, err := datastore.NewClient(ctx, config.ProjectID)
	if err != nil {
		return err
	}
	defer client.Close()
	if _, err := client.Put(ctx, errorKey(), &data); err != nil {
		return err
	}
	return nil
}

func StoreShare(ctx context.Context, data ShareData) error {
	client, err := datastore.NewClient(ctx, config.ProjectID)
	if err != nil {
		return err
	}
	defer client.Close()
	if _, err := client.Put(ctx, shareKey(), &data); err != nil {
		return err
	}
	return nil
}

func StoreDeploy(ctx context.Context, data DeployData) error {
	client, err := datastore.NewClient(ctx, config.ProjectID)
	if err != nil {
		return err
	}
	defer client.Close()
	if _, err := client.Put(ctx, deployKey(), &data); err != nil {
		return err
	}
	return nil
}

func StoreCompile(ctx context.Context, path string, data CompileData) error {
	client, err := datastore.NewClient(ctx, config.ProjectID)
	if err != nil {
		return err
	}
	defer client.Close()
	if _, err := client.Put(ctx, compileKey(), &data); err != nil {
		return err
	}
	if _, err := client.Put(ctx, packageKey(path), &data); err != nil {
		return err
	}
	return nil
}

func Package(ctx context.Context, path string) (bool, CompileData, error) {
	client, err := datastore.NewClient(ctx, config.ProjectID)
	if err != nil {
		return false, CompileData{}, err
	}
	defer client.Close()
	var data CompileData

	if err := client.Get(ctx, packageKey(path), &data); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return false, CompileData{}, nil
		}
		return false, CompileData{}, err
	}
	return true, data, nil
}

func errorKey() *datastore.Key {
	return datastore.IncompleteKey(config.ErrorKind, nil)
}

func compileKey() *datastore.Key {
	return datastore.IncompleteKey(config.CompileKind, nil)
}

func deployKey() *datastore.Key {
	return datastore.IncompleteKey(config.DeployKind, nil)
}

func shareKey() *datastore.Key {
	return datastore.IncompleteKey(config.ShareKind, nil)
}

func packageKey(path string) *datastore.Key {
	return datastore.NameKey(config.PackageKind, path, nil)
}
