package comms

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bradfitz/slice"
	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// EtcdStorage implements GzrMetadataStore and has exported
// Etcd clients and KV accessors
type EtcdStorage struct {
	Client    *clientv3.Client
	KV        clientv3.KV
	activeTxn clientv3.Txn
}

// NewEtcdStorage initializes and returns a pointer to an EtcdStorage
// with a connected Client and KV
func NewEtcdStorage() (GzrMetadataStore, error) {
	newEtcd := &EtcdStorage{}
	cxnString := fmt.Sprintf("%s:%s", viper.GetString("datastore.host"), viper.GetString("datastore.port"))
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{cxnString},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create etcd client")
	}
	newEtcd.Client = cli

	kv := clientv3.NewKV(cli)
	newEtcd.KV = kv
	return newEtcd, nil
}

// List queries the etcd store for all images stored under a particular name
func (store *EtcdStorage) List(imageName string) (*ImageList, error) {
	resp, err := store.KV.Get(context.Background(), fmt.Sprintf("%s:", imageName), clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to retrieve images from etcd for %q", imageName)
	}
	return store.extractImages(resp)
}

// Store stores the metadata about an image associated with its name in etcd
func (store *EtcdStorage) Store(imageName string, meta ImageMetadata) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return errors.Wrap(err, "Failed to convert image metadata into json")
	}

	key, err := createKey(imageName)
	if err != nil {
		return errors.Wrapf(err, "Failed to create key %q in etcd", imageName)
	}
	store.activeTxn = store.activeTxn.Then(clientv3.OpPut(key, string(data)))
	if err != nil {
		return errors.Wrapf(err, "Failed to store image metadata for image %q in etcd", imageName)
	}
	return nil
}

// Cleanup closes the etcd client connection
func (store *EtcdStorage) Cleanup() {
	store.Client.Close()
}

// Delete deletes all information related to IMAGE_NAME:VERSION
func (store *EtcdStorage) Delete(imageName string) (int, error) {
	resp, err := store.KV.Delete(context.Background(), imageName, clientv3.WithPrefix())
	return int(resp.Deleted), errors.Wrapf(err, "Failed to delete image %q", imageName)
}

// Get returns a single image based on a name
func (store *EtcdStorage) Get(imageName string) (*Image, error) {
	resp, err := store.KV.Get(context.Background(), fmt.Sprintf("%s", imageName), clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get image %q from etcd", imageName)
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	if len(resp.Kvs) > 1 {
		return nil, errors.Errorf("Found multiple images for %q", imageName)
	}
	return store.extractImage(resp.Kvs[0].Value, resp.Kvs[0].Key), nil
}

// GetLatest returns the latest image from a name
func (store *EtcdStorage) GetLatest(imageName string) (*Image, error) {
	images, err := store.List(imageName)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get images for %q", imageName)
	}
	slice.Sort(images.Images, func(i, j int) bool {
		return images.Images[j].Meta.CreatedAt < images.Images[i].Meta.CreatedAt
	})
	return images.Images[0], nil
}

// StartTransaction sets a new transaction on the EtcdStorage
func (store *EtcdStorage) StartTransaction() error {
	eTxn := store.KV.Txn(context.Background())
	store.activeTxn = eTxn
	return nil
}

// CommitTransaction commits the active transaction
func (store *EtcdStorage) CommitTransaction() error {
	_, err := store.activeTxn.Commit()
	// TODO: Check response for errors
	return errors.Wrap(err, "Failed to commit transaction to etcd")
}

// extractImage transforms raw []byte of metadata and key into a full Image
func (store *EtcdStorage) extractImage(data []byte, key []byte) *Image {
	var meta ImageMetadata
	json.Unmarshal(data, &meta)
	return &Image{Name: string(key), Meta: meta}
}

// extractImages transforms an etcd response into an []Image
func (store *EtcdStorage) extractImages(resp *clientv3.GetResponse) (*ImageList, error) {
	if len(resp.Kvs) < 1 {
		return nil, errors.Errorf("No results found")
	}
	var images []*Image
	for _, kv := range resp.Kvs {
		images = append(images, store.extractImage(kv.Value, kv.Key))
	}
	return &ImageList{Images: images}, nil
}
