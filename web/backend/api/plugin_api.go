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

	// swagger:route GET /plugin plugin getPluginParams
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
	//     Responses:
	//       200: pluginDetail
	//		 503: apiError
	group.GET("", func(c *gin.Context) {
		req := &repo.GetPluginParams{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err)
			return
		}

		if req.Name == nil && req.PluginType == nil {
			c.Error(errors.New("no params"))
			return
		}

		output, err := service.GetPluginDetail(c, req)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /plugin/list plugin listPluginParams
	//
	// List plugin by name and version.
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
	//       200: []pluginDetail
	//		 503: apiError
	group.GET("list", func(c *gin.Context) {
		req := &repo.ListPluginParams{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err)
			return
		}

		if req.Name == nil && req.PluginType == nil {
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

	// swagger:route GET /plugin/label plugin addLabelParams
	//
	// Add label in plugin.
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
	//       200:
	//		 503: apiError
	group.POST("label", func(c *gin.Context) {
		req := &repo.AddLabelParams{}
		err := c.ShouldBindJSON(req)
		if err != nil {
			c.Error(err)
			return
		}

		if req.Name == nil && req.Label == nil {
			c.Error(errors.New("no params"))
			return
		}

		err = service.AddLabel(c, *req.Name, *req.Label)
		if err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusOK)
	})

	// swagger:route GET /plugin/label plugin deleteLabelParams
	//
	// Delete label in plugin.
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
	//       200:
	//		 503: apiError
	group.DELETE("label", func(c *gin.Context) {
		req := &repo.DeleteLabelParams{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err)
			return
		}

		if req.Name == nil && req.Label == nil {
			c.Error(errors.New("no params"))
			return
		}

		err = service.DeleteLabel(c, *req.Name, *req.Label)
		if err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusOK)
	})

	// swagger:route DELETE /plugin plugin deletePlugin
	//
	// Delete plugin by id and specific version
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
	//       + name: version
	//         in: query
	//         description: version of plugin
	//         required: false
	//         type: string
	//
	//     Responses:
	//       200:
	//		 503: apiError
	group.DELETE("", func(c *gin.Context) {
		req := &models.DeletePluginReq{}
		err := c.ShouldBind(&req)
		if err != nil {
			c.Error(err)
			return
		}

		params := &repo.DeletePluginParams{
			Version: req.Version,
		}
		params.Id, err = primitive.ObjectIDFromHex(req.Id)
		if err != nil {
			c.Error(err)
			return
		}

		pluginDetail, err := service.GetPluginDetail(ctx, &repo.GetPluginParams{
			Id: &req.Id,
		})
		if err != nil {
			c.Error(err)
			return
		}

		plugin, err := service.GetPlugin(ctx, pluginDetail.Name, req.Version)
		if err != nil {
			c.Error(err)
			return
		}

		err = service.DeletePluginByVersion(c, params)
		if err != nil {
			c.Error(err)
			return
		}

		err = os.Remove(plugin.Path)
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

			plugin := &models.Plugin{
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
			}

			err = service.SavePlugins(c, plugin)
			if err != nil {
				c.Error(err)
				return
			}
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

			plugin := &models.Plugin{
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
			}

			err = service.SavePlugins(c, plugin)
			if err != nil {
				c.Error(err)
				return
			}
		}
		c.Status(http.StatusOK)
	})

}
