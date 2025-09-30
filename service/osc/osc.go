/*
Copyright 2023 The Cloud-Barista Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package osc

import (
	"io"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/filtering"
	"github.com/rs/zerolog"
)

type OSFS interface {
	CreateBucket() error
	DeleteBucket() error
	ObjectList() ([]*models.Object, error)

	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
}

type OSController struct {
	osfs OSFS

	logger  *zerolog.Logger
	threads int
}

type FilterableOSFS interface {
	ObjectListWithFilter(*filtering.ObjectFilter) ([]*models.Object, error)
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

func (osc *OSController) ObjectList() ([]*models.Object, error) {
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

func WithLogger(logger *zerolog.Logger) Option {
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
			osc.logger.Info().Msg(msg)
		case "Error":
			osc.logger.Error().Msgf("%s : %v", msg, err)
		}
	}
}

func (o *OSController) ObjectListWithFilter(flt *filtering.ObjectFilter) ([]*models.Object, error) {
	if f, ok := o.osfs.(FilterableOSFS); ok {
		return f.ObjectListWithFilter(flt)
	}
	objs, err := o.osfs.ObjectList()
	if err != nil {
		return nil, err
	}
	out := make([]*models.Object, 0, len(objs))
	for _, m := range objs {
		c := filtering.Candidate{Key: m.Key, Size: m.Size, LastModified: m.LastModified}
		if filtering.MatchCandidate(flt, c) {
			out = append(out, m)
		}
	}
	return out, nil
}
