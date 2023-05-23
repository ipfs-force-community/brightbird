package api

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hunjixin/brightbird/env/plugin"
	"github.com/hunjixin/brightbird/models"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"github.com/hunjixin/brightbird/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterDeployRouter(ctx context.Context, pluginStore types.PluginStore, v1group *V1RouterGroup, service repo.IPluginService) {
	group := v1group.Group("/plugin")

	// swagger:route GET /plugin/mainfest plugin getPluginMainfest
	//
	// Get plugin mainfest.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//     - application/text
	//
	//     Schemes: http, https
	//
	//     Deprecated: false
	//
	//     Parameters:
	//       + name: name
	//         in: query
	//         description: name of plugin
	//         required: false
	//         type: string
	//       + name: pluginType
	//         in: query
	//         description: pluginType of plugin
	//         required: false
	//         type: string
	//
	//     Responses:
	//       200: []pluginInfo
	//		 503: apiError
	group.GET("mainfest", func(c *gin.Context) {
		req := &repo.ListMainFestParams{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err)
			return
		}

		output, err := service.PluginSummary(c, req)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /plugin plugin getPlugin
	//
	// Get plugin by name and version.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//     - application/text
	//
	//     Schemes: http, https
	//
	//     Deprecated: false
	//
	//     Parameters:
	//       + name: name
	//         in: query
	//         description: name of plugin
	//         required: false
	//         type: string
	//       + name: version
	//         in: query
	//         description: version of plugin
	//         required: false
	//         type: string
	//       + name: pluginType
	//         in: query
	//         description: pluginType of plugin
	//         required: false
	//         type: string
	//
	//     Responses:
	//       200: []pluginDetail
	//		 503: apiError
	group.GET("", func(c *gin.Context) {
		req := &repo.ListPluginParams{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err)
			return
		}

		if req.Name == nil && req.Version == nil && req.PluginType == nil {
			c.Error(errors.New("no params"))
			return
		}

		output, err := service.ListPlugin(c, req)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route DELETE /plugin plugin deletePlugin
	//
	// Delete plugin by id
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//     - application/text
	//
	//     Schemes: http, https
	//
	//     Deprecated: false
	//
	//     Parameters:
	//       + name: id
	//         in: query
	//         description: id of plugin
	//         required: false
	//         type: string
	//
	//     Responses:
	//       200:
	//		 503: apiError
	group.DELETE("", func(c *gin.Context) {
		idStr := c.Query("id")

		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			c.Error(err)
			return
		}

		err = service.DeletePlugin(c, id)
		if err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusOK)
	})

	// UploadPluginFilesParams contains the uploaded file data
	// swagger:parameters uploadPluginFilesParams
	type UploadPluginFilesParams struct {
		// Plugin file.
		//
		// in: formData
		//
		// swagger:file
		PluginFiles []*multipart.FileHeader `json:"plugins" form:"plugins"`
	}

	// uploadPlugin swagger:route POST /plugin/upload plugin uploadPluginFilesParams
	//
	// Upload plugin files
	//
	// Responses:
	//	    200:
	//	    403: apiError
	group.POST("upload", func(c *gin.Context) {
		params := &UploadPluginFilesParams{}
		err := c.ShouldBind(params)
		// The file cannot be received.
		if err != nil {
			c.Error(err)
			return
		}

		var pluginInfos []*models.PluginDetail
		for _, fileHeader := range params.PluginFiles {
			// The file is received, so let's save it
			tmpPath := path.Join(os.TempDir(), uuid.NewString())
			if err := c.SaveUploadedFile(fileHeader, tmpPath); err != nil {
				c.Error(err)
				return
			}

			pluginInfo, err := plugin.GetPluginInfo(tmpPath)
			if err != nil {
				c.Error(err)
				return
			}

			// copy plugin to plugin store
			fname := fmt.Sprintf("%s_%s_%s", pluginInfo.PluginType, pluginInfo.Name, pluginInfo.Version)
			err = utils.CopyFile(tmpPath, filepath.Join(string(pluginStore), fname))
			if err != nil {
				c.Error(err)
				return
			}

			pluginInfos = append(pluginInfos, &models.PluginDetail{
				ID: primitive.NewObjectID(),
				BaseTime: models.BaseTime{
					CreateTime:   time.Now().Unix(),
					ModifiedTime: time.Now().Unix(),
				},
				PluginInfo: *pluginInfo,
				Path:       fname,
				Instance: types.DependencyProperty{
					Name:        plugin.InstancePropertyName,
					Value:       "default",
					Type:        pluginInfo.PluginType,
					SockPath:    "",
					Description: "named a plugin instance",
					Require:     true,
				},
			})
		}

		err = service.SavePlugins(c, pluginInfos)
		if err != nil {
			c.Error(err)
			return
		}
		c.Status(http.StatusOK)
	})

	// swagger:route POST /plugin/import plugin importPlugin
	//
	// import plugin mainfest.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//     - application/text
	//
	//     Schemes: http, https
	//
	//     Deprecated: false
	//
	//     Parameters:
	//       + name: path
	//         in: query
	//         description: directory of plugins
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200:
	//		 503: apiError
	group.POST("import", func(c *gin.Context) {
		path := c.Query("path")
		filePaths := []string{}
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				filePaths = append(filePaths, path)
			}
			return nil
		})
		if err != nil {
			c.Error(err)
			return
		}
		var pluginInfos []*models.PluginDetail
		for _, pluginPath := range filePaths {
			pluginInfo, err := plugin.GetPluginInfo(pluginPath)
			if err != nil {
				c.Error(err)
				return
			}

			//copy plugin to plugin store
			fname := fmt.Sprintf("%s_%s_%s", pluginInfo.PluginType, pluginInfo.Name, pluginInfo.Version)
			err = utils.CopyFile(pluginPath, filepath.Join(string(pluginStore), fname))
			if err != nil {
				fmt.Println(err)
				c.Error(err)
				return
			}
			pluginInfos = append(pluginInfos, &models.PluginDetail{
				ID: primitive.NewObjectID(),
				BaseTime: models.BaseTime{
					CreateTime:   time.Now().Unix(),
					ModifiedTime: time.Now().Unix(),
				},
				PluginInfo: *pluginInfo,
				Path:       fname,
				Instance: types.DependencyProperty{
					Name:        plugin.InstancePropertyName,
					Value:       "default",
					Type:        pluginInfo.PluginType,
					SockPath:    "",
					Description: "named a plugin instance",
					Require:     true,
				},
			})
		}

		err = service.SavePlugins(c, pluginInfos)
		if err != nil {
			c.Error(err)
			return
		}
		c.Status(http.StatusOK)
	})

}
