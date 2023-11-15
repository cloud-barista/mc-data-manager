package osc

import (
	"io"

	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/sirupsen/logrus"
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

	logger  *logrus.Logger
	threads int
}

type Result struct {
	name string
	err  error
}

func (osc *OSController) CreateBucket() error {
	err := osc.osfs.CreateBucket()
	if err != nil {
		return err
	}
	return nil
}

func (osc *OSController) DeleteBucket() error {
	err := osc.osfs.DeleteBucket()
	if err != nil {
		return err
	}
	return nil
}

func (osc *OSController) ObjectList() ([]*utils.Object, error) {
	objList, err := osc.osfs.ObjectList()
	if err != nil {
		return objList, err
	}
	return objList, nil
}

type Option func(*OSController)

func WithThreads(count int) Option {
	return func(o *OSController) {
		if count >= 1 {
			o.threads = count
		}
	}
}

func WithLogger(logger *logrus.Logger) Option {
	return func(o *OSController) {
		o.logger = logger
	}
}

func New(osfs OSFS, opts ...Option) (*OSController, error) {
	osc := &OSController{
		osfs:    osfs,
		threads: 10,
		logger:  nil,
	}

	for _, opt := range opts {
		opt(osc)
	}

	return osc, nil
}

func (osc *OSController) logWrite(logLevel, msg string, err error) {
	if osc.logger != nil {
		switch logLevel {
		case "Info":
			osc.logger.Info(msg)
		case "Error":
			osc.logger.Errorf("%s : %v", msg, err)
		}
	}
}
