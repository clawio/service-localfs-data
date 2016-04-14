package datacontroller

import (
	"io"

	"github.com/clawio/entities"
)

// DataController is an interface to upload and download blobs.
type DataController interface {
	UploadBLOB(user entities.User, pathSpec string, r io.Reader, clientChecksum string) error
	DownloadBLOB(user entities.User, pathSpec string) (io.Reader, error)
}
