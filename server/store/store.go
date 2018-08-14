package store

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/dave/jsgo/config"
	"github.com/dave/services"
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

type WasmDeploy struct {
	Time  time.Time
	Ip    string
	Files []WasmDeployFile
}

type WasmDeployFile struct {
	Type string
	Hash string
}

func StoreError(ctx context.Context, database services.Database, data Error) error {
	if _, err := database.Put(ctx, errorKey(), &data); err != nil {
		return err
	}
	return nil
}

func StoreShare(ctx context.Context, database services.Database, data ShareData) error {
	if _, err := database.Put(ctx, shareKey(), &data); err != nil {
		return err
	}
	return nil
}

func StoreDeploy(ctx context.Context, database services.Database, data DeployData) error {
	if _, err := database.Put(ctx, deployKey(), &data); err != nil {
		return err
	}
	return nil
}

func StoreCompile(ctx context.Context, database services.Database, path string, data CompileData) error {
	if _, err := database.Put(ctx, compileKey(), &data); err != nil {
		return err
	}
	if _, err := database.Put(ctx, packageKey(path), &data); err != nil {
		return err
	}
	return nil
}

func StoreWasmDeploy(ctx context.Context, database services.Database, data WasmDeploy) error {
	if _, err := database.Put(ctx, wasmDeployKey(), &data); err != nil {
		return err
	}
	return nil
}

func Package(ctx context.Context, database services.Database, path string) (bool, CompileData, error) {
	var data CompileData
	if err := database.Get(ctx, packageKey(path), &data); err != nil {
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

func wasmDeployKey() *datastore.Key {
	return datastore.IncompleteKey(config.WasmDeployKind, nil)
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
