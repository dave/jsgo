package localdatabase

import (
	"context"
	"errors"
	"math/rand"
	"reflect"

	"time"

	"fmt"
	"net/url"
	"path/filepath"

	"os"

	"encoding/json"

	"cloud.google.com/go/datastore"
)

func New(dir string) *Database {
	return &Database{
		dir: dir,
	}
}

type Database struct {
	dir string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (d *Database) fpath(key *datastore.Key) string {
	kind := url.PathEscape(key.Kind)
	var name string
	if key.ID > 0 {
		name = fmt.Sprint(key.ID)
	} else {
		name = url.PathEscape(key.Name)
	}
	return filepath.Join(d.dir, "datastore", kind, name+".json")
}

func (d *Database) Get(ctx context.Context, key *datastore.Key, dst interface{}) (err error) {
	if key.Incomplete() {
		return datastore.ErrInvalidKey
	}
	f, err := os.Open(d.fpath(key))
	if err != nil {
		if os.IsNotExist(err) {
			return datastore.ErrNoSuchEntity
		}
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(dst)
}

func (d *Database) Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	if key.Incomplete() {
		key.ID = rand.Int63()
	}
	fpath := d.fpath(key)
	dir, _ := filepath.Split(fpath)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}
	f, err := os.Create(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(src); err != nil {
		return nil, err
	}
	return key, nil
}

func (d *Database) GetMulti(ctx context.Context, keys []*datastore.Key, dst interface{}) (err error) {
	v := reflect.ValueOf(dst)
	multiArgType, _ := checkMultiArg(v)

	// Sanity checks
	if multiArgType == multiArgTypeInvalid {
		return errors.New("datastore: dst has invalid type")
	}
	if len(keys) != v.Len() {
		return errors.New("datastore: keys and dst slices have different length")
	}
	if len(keys) == 0 {
		return nil
	}

	multiErr, any := make(datastore.MultiError, len(keys)), false
	for i, k := range keys {
		if k.Incomplete() {
			multiErr[i] = datastore.ErrInvalidKey
			any = true
		}
	}
	if any {
		return multiErr
	}

	for i, k := range keys {
		elem := v.Index(i)
		if multiArgType == multiArgTypePropertyLoadSaver || multiArgType == multiArgTypeStruct {
			elem = elem.Addr()
		}
		if multiArgType == multiArgTypeStructPtr && elem.IsNil() {
			elem.Set(reflect.New(elem.Type().Elem()))
		}
		if err := d.Get(ctx, k, elem.Interface()); err != nil {
			if err == datastore.ErrNoSuchEntity {
				fmt.Println("FOO")
				multiErr[i] = datastore.ErrNoSuchEntity
				any = true
			} else {
				fmt.Println("BAR")
				multiErr[i] = err
				any = true
			}
		}
	}

	if any {
		return multiErr
	}
	return nil
}

func (d *Database) PutMulti(ctx context.Context, keys []*datastore.Key, src interface{}) (_ []*datastore.Key, err error) {
	v := reflect.ValueOf(src)
	multiArgType, _ := checkMultiArg(v)
	if multiArgType == multiArgTypeInvalid {
		return nil, errors.New("datastore: src has invalid type")
	}
	if len(keys) != v.Len() {
		return nil, errors.New("datastore: key and src slices have different length")
	}
	if len(keys) == 0 {
		return nil, nil
	}
	multiErr := make(datastore.MultiError, len(keys))
	hasErr := false
	for i, k := range keys {
		elem := v.Index(i)
		// Two cases where we need to take the address:
		// 1) multiArgTypePropertyLoadSaver => &elem implements PLS
		// 2) multiArgTypeStruct => saveEntity needs *struct
		if multiArgType == multiArgTypePropertyLoadSaver || multiArgType == multiArgTypeStruct {
			elem = elem.Addr()
		}
		if _, err := d.Put(ctx, k, elem.Interface()); err != nil {
			multiErr[i] = err
			hasErr = true
		}
	}
	if hasErr {
		return nil, multiErr
	}
	return keys, nil
}

// checkMultiArg checks that v has type []S, []*S, []I, or []P, for some struct
// type S, for some interface type I, or some non-interface non-pointer type P
// such that P or *P implements PropertyLoadSaver.
//
// It returns what category the slice's elements are, and the reflect.Type
// that represents S, I or P.
//
// As a special case, PropertyList is an invalid type for v.
//
// TODO(djd): multiArg is very confusing. Fold this logic into the
// relevant Put/Get methods to make the logic less opaque.
func checkMultiArg(v reflect.Value) (m multiArgType, elemType reflect.Type) {
	if v.Kind() != reflect.Slice {
		return multiArgTypeInvalid, nil
	}
	if v.Type() == typeOfPropertyList {
		return multiArgTypeInvalid, nil
	}
	elemType = v.Type().Elem()
	if reflect.PtrTo(elemType).Implements(typeOfPropertyLoadSaver) {
		return multiArgTypePropertyLoadSaver, elemType
	}
	switch elemType.Kind() {
	case reflect.Struct:
		return multiArgTypeStruct, elemType
	case reflect.Interface:
		return multiArgTypeInterface, elemType
	case reflect.Ptr:
		elemType = elemType.Elem()
		if elemType.Kind() == reflect.Struct {
			return multiArgTypeStructPtr, elemType
		}
	}
	return multiArgTypeInvalid, nil
}

type multiArgType int

const (
	multiArgTypeInvalid multiArgType = iota
	multiArgTypePropertyLoadSaver
	multiArgTypeStruct
	multiArgTypeStructPtr
	multiArgTypeInterface
)

var (
	typeOfPropertyLoadSaver = reflect.TypeOf((*datastore.PropertyLoadSaver)(nil)).Elem()
	typeOfPropertyList      = reflect.TypeOf(datastore.PropertyList(nil))
)
