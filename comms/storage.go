package comms

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

// GzrMetadataStore is the interface that should be implemented for any
// backend that needs to handle image data storage
type GzrMetadataStore interface {
	// Store stores image metadata with a name
	Store(string, ImageMetadata) error
	// List lists all of the images under a name
	List(string) (*ImageList, error)
	// Cleanup allows the storage backend to clean up any connections, etc
	Cleanup()
	// Delete deletes all images under a nmae, returns number of deleted entries
	Delete(string) (int, error)
	// Get gets a single image with a version
	Get(string) (*Image, error)
	// GetLatest gets the most recent image from a name
	GetLatest(string) (*Image, error)
	// StartTransaction starts a new transaction within the GzrMetadataStore
	StartTransaction() error
	// CommitTransaction commits the active transaction
	CommitTransaction() error
}

// StorageTransaction is an interface to manage transactions around storage
// concerns
type StorageTransaction interface {
	// Commit persists the operations from the transaction
	Commit() error
}

// Image is a struct unifying an image name with its metadata
type Image struct {
	// Name is the image's full name
	Name string `json:"name"`
	// Meta is the metadata related to the image
	Meta ImageMetadata `json:"metadata"`
}

// ImageMetadata is a struct containing necessary metadata about a particular image
type ImageMetadata struct {
	// GitCommit is the commit related to the built image
	GitCommit string `json:"git-commit"`
	// GitTag is the tag related to the built image if it exists
	GitTag []string `json:"git-tag"`
	// GitAnnotation is the annotation related to the built image if it exists
	GitAnnotation []string `json:"git-annotation"`
	// GitOrigin is the remote origin for the git repo associated with the image
	GitOrigin string `json:"git-origin"`
	// CreatedAt is the time the metadata was stored, with day granularity
	CreatedAt string `json:"created-at"`
}

// ImageList is a collection of Images
type ImageList struct {
	Images []*Image `json:"images"`
}

// SerializeForCLI takes an io.Writer and writes templatized data to it representing an ImageList
func (l *ImageList) SerializeForCLI(wr io.Writer) error {
	return l.cliTemplate().Execute(wr, l)
}

func (l *ImageList) cliTemplate() *template.Template {
	t := template.New("Images")
	t, _ = t.Parse(`Images {{range .Images}}
- name: {{.Name}}
  -- git-commit: {{.Meta.GitCommit}}
  -- git-tag: [{{ range $index, $element := .Meta.GitTag}}{{if $index}}, {{end}}{{$element}}{{end}}]
  -- git-annotation: [{{ range $index, $element := .Meta.GitAnnotation}}{{if $index}}, {{end}}{{$element}}{{end}}]
  -- git-origin: {{.Meta.GitOrigin}}
  -- created-at: {{.Meta.CreatedAt}}
{{end}}
`)
	return t
}

// SerializeForWire returns a JSON representation of the ImageList
func (imageList *ImageList) SerializeForWire() ([]byte, error) {
	data, err := json.Marshal(imageList)
	return data, errors.Wrap(err, "Failed to transform image list to JSON")
}

// SerializeForWire returns a JSON representation of the Image
func (image *Image) SerializeForWire() ([]byte, error) {
	data, err := json.Marshal(image)
	return data, errors.Wrap(err, "Failed to transform image to JSON")
}

// SerializeForCLI takes an io.Writer and writes templated data to it representing an Image
func (l *Image) SerializeForCLI(wr io.Writer) error {
	return l.cliTemplate().Execute(wr, l)
}

func (l *Image) cliTemplate() *template.Template {
	t := template.New("Image")
	t, _ = t.Parse(`{{.Name}}`)
	return t
}

// CreateMeta takes a ReadWriter and returns an instance of ImageMetadata
// after parsing
func CreateMeta(reader io.ReadWriter) (ImageMetadata, error) {
	var meta ImageMetadata
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return ImageMetadata{}, errors.Wrapf(err, "Could not read metadata file")
	}

	err = json.Unmarshal(b, &meta)
	if err != nil {
		return ImageMetadata{}, errors.Wrapf(err, "Could not read metadata file: \nCheck the data types in your image metadata JSON.")
	}
	return meta, nil
}

// createKey creates the key used to tag data in stores
func createKey(imageName string) (string, error) {
	splitName := strings.Split(imageName, ":")
	if len(splitName) != 2 {
		return "", errors.New("IMAGE_NAME must be formatted as NAME:VERSION and must contain only the seperating colon")
	}
	name := fmt.Sprintf("%s:%s", splitName[0], splitName[1])
	fmt.Println(name)
	return name, nil
}
