package datacontroller

import (
	"github.com/clawio/service-auth/server/spec"
	"io"
)

// DataController is an interface to upload and download blobs.
type DataController interface {
	UploadBLOB(user *spec.Identity, pathSpec string, r io.Reader, clientChecksum string) error
	DownloadBLOB(user *spec.Identity, pathSpec string) (io.Reader, error)
}
