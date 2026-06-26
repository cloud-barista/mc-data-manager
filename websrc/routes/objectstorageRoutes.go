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

package routes

import (
	"github.com/cloud-barista/mc-data-manager/websrc/controllers"
	"github.com/labstack/echo/v4"
)

func ObjectStorageRoutes(g *echo.Group) {
	g.PUT("/buckets", controllers.ObjectstorageCreateBucketHandler)
	g.DELETE("/buckets", controllers.ObjectstorageDeleteBucketHandler)
	g.POST("/buckets", controllers.ObjectstorageBucketsHandler)
	g.POST("/buckets/objects", controllers.ObjectstorageObjectListHandler)
	g.DELETE("/buckets/object", controllers.ObjectstorageDeleteObjectHandler)
}
