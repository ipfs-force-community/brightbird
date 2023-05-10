package api

import (
	"context"
	"fmt"
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

	// swagger:route GET /plugin/deploy listDeployPlugins
	//
	// Lists all deploy plugin.
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
	//     Responses:
	//       200: []pluginOut
	//		 503: apiError
	group.GET("deploy/list", func(c *gin.Context) {
		output, err := service.DeployPlugins(c)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /plugin/exec listExecPlugin
	//
	// Lists all deploy plugin.
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
	//     Responses:
	//       200: []pluginOut
	//		 503: apiError
	group.GET("exec/list", func(c *gin.Context) {
		output, err := service.ExecPlugins(c)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /plugin/get getPlugin
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
	//         required: true
	//         type: string
	//       + name: version
	//         in: query
	//         description: version of plugin
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200: pluginOut
	//		 503: apiError
	group.GET("", func(c *gin.Context) {
		req := &models.GetPluginRequest{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err)
			return
		}

		output, err := service.GetPlugin(c, req.Name, req.Version)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route POST /plugin/upload Files uploadPluginFile
	//
	// Upload a file.
	//
	//
	//     Consumes:
	//     - multipart/form-data
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: https
	//
	//     Deprecated: false
	//
	//     Parameters:
	//       + name: file
	//         in: formData
	//         required: true
	//         type: file
	//
	//     Responses:
	//       200: StatusOK
	//       403: Error

	group.GET("upload", func(c *gin.Context) {
		form, err := c.MultipartForm()
		// The file cannot be received.
		if err != nil {
			c.Error(err)
			return
		}

		var pluginInfos []*models.PluginOut
		for _, files := range form.File {
			for _, fileHeader := range files {
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

				pluginInfos = append(pluginInfos, &models.PluginOut{
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
		}

		err = service.SavePlugins(c, pluginInfos)
		if err != nil {
			c.Error(err)
			return
		}
		c.Status(http.StatusOK)
	})

	// swagger:route POST /plugin/import importPlugin
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
		var pluginInfos []*models.PluginOut
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
			pluginInfos = append(pluginInfos, &models.PluginOut{
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
