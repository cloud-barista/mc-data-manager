package osc

import (
	"io"

	"github.com/cloud-barista/cm-data-mold/pkg/utils"
)

type OSFS interface {
	CreateBucket() error
	DeleteBucket() error
	ObjectList() ([]*utils.Object, error)

	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
}

type OSController struct {
	osfs OSFS

	threads int
}

func (osc *OSController) CreateBucket() error {
	return osc.osfs.CreateBucket()
}

func (osc *OSController) DeleteBucket() error {
	return osc.osfs.DeleteBucket()
}

func (osc *OSController) ObjectList() ([]*utils.Object, error) {
	return osc.osfs.ObjectList()
}

type Option func(*OSController)

func WithThreads(count int) Option {
	return func(o *OSController) {
		if count >= 1 {
			o.threads = count
		}
	}
}

func New(osfs OSFS, opts ...Option) (*OSController, error) {
	osc := &OSController{
		osfs:    osfs,
		threads: 10,
	}

	for _, opt := range opts {
		opt(osc)
	}

	return osc, nil
}
