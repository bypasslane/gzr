package comms

import (
	"bytes"
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/bradfitz/slice"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	ImageBucket = "images"
)

// BoltStorage implements GzrMetadataStore and has an un-exported bolt.db pointer
type BoltStorage struct {
	db        *bolt.DB
	activeTxn *bolt.Tx
}

// NewBoltStorage initializes a BoltDB connection, makes sure the correct buckets exist,
// and returns a BoltStorage pointer with the established connection
func NewBoltStorage() (GzrMetadataStore, error) {
	dbPath := viper.GetString("datastore.db_path")
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open connection to bolt database")
	}
	store := &BoltStorage{db: db}
	txn, err := store.db.Begin(true)
	if err != nil {
		store.Cleanup()
		return nil, errors.Wrap(err, "Failed to start transaction in bolt database")
	}
	_, err = txn.CreateBucketIfNotExists([]byte(ImageBucket))
	if err != nil {
		store.Cleanup()
		return nil, errors.Wrapf(err, "Failed to create bucket %q", ImageBucket)
	}
	err = txn.Commit()
	if err != nil {
		store.Cleanup()
		return nil, errors.Wrap(err, "Failed to commit changes to bolt database")
	}
	return store, nil
}

// List queries the Bolt store for all images stored under a particular name
func (store *BoltStorage) List(imageName string) (*ImageList, error) {
	var images []*Image
	err := store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ImageBucket))
		c := b.Cursor()
		prefix := []byte(imageName)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			img := store.extractImage(v, k)
			images = append(images, img)
		}
		return nil
	})
	if err != nil {
		return &ImageList{}, errors.Wrap(err, "Failed to retrieve image list from bolt database")
	}
	return &ImageList{Images: images}, nil
}

// Store stores the metadata about an image associated with its name
func (store *BoltStorage) Store(imageName string, meta ImageMetadata) error {
	b := store.activeTxn.Bucket([]byte(ImageBucket))
	data, err := json.Marshal(meta)
	if err != nil {
		return errors.Wrapf(err, "Failed to convert metadata into json for image %q", imageName)
	}
	key, err := createKey(imageName)
	if err != nil {
		return errors.Wrapf(err, "Failed to create db key %q in bolt db", imageName)
	}
	err = b.Put([]byte(key), data)
	if err != nil {
		return errors.Wrapf(err, "Failed to store metadata for key %q in bolt db", imageName)
	}
	return nil
}

// Cleanup closes the Bolt connection
func (store *BoltStorage) Cleanup() {
	store.db.Close()
}

// Delete deletes all information related to IMAGE_NAME:VERSION
func (store *BoltStorage) Delete(imageName string) (int, error) {
	b := store.activeTxn.Bucket([]byte(ImageBucket))
	c := b.Cursor()
	prefix := []byte(imageName)
	deleted := 0
	for key, _ := c.Seek(prefix); key != nil && bytes.HasPrefix(key, prefix); key, _ = c.Next() {
		err := b.Delete(key)
		if err != nil {
			return 0, errors.Wrapf(err, "Failed to delete key %q from bolt db", key)
		}
		deleted += 1
	}
	return deleted, nil
}

// Get returns a single image based on a name
func (store *BoltStorage) Get(imageName string) (*Image, error) {
	var image *Image
	err := store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ImageBucket))
		c := b.Cursor()
		prefix := []byte(imageName)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			image = store.extractImage(v, k)
		}
		return nil
	})
	if err != nil {
		return &Image{}, errors.Wrapf(err, "Failed to get images for %q from bolt db", imageName)
	}
	return image, nil
}

// GetLatest returns the latest image from a name
func (store *BoltStorage) GetLatest(imageName string) (*Image, error) {
	images, err := store.List(imageName)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to get images for %q", imageName)
	}
	slice.Sort(images.Images, func(i, j int) bool {
		return images.Images[j].Meta.CreatedAt < images.Images[i].Meta.CreatedAt
	})
	return images.Images[0], nil
}

// StartTransaction starts a new Bolt transaction and adds it to the Storage
func (store *BoltStorage) StartTransaction() error {
	bTxn, err := store.db.Begin(true)
	if err != nil {
		return errors.Wrap(err, "Failed to create transaction")
	}
	store.activeTxn = bTxn
	return nil
}

// CommitTransaction commits the active transaction
func (store *BoltStorage) CommitTransaction() error {
	return store.activeTxn.Commit()
}

// extractImage transforms raw []byte of metadata and key into a full Image
func (store *BoltStorage) extractImage(data []byte, key []byte) *Image {
	var meta ImageMetadata
	json.Unmarshal(data, &meta)
	return &Image{Name: string(key), Meta: meta}
}
