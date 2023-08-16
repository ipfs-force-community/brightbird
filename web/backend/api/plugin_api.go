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
	"github.com/ipfs-force-community/brightbird/env/plugin"
	"github.com/ipfs-force-community/brightbird/models"
	"github.com/ipfs-force-community/brightbird/repo"
	"github.com/ipfs-force-community/brightbird/types"
	"github.com/ipfs-force-community/brightbird/utils"
	"github.com/mholt/archiver/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterDeployRouter(ctx context.Context, pluginStore types.PluginStore, v1group *V1RouterGroup, service repo.IPluginService) {
	group := v1group.Group("/plugin")

	// swagger:route GET /plugin plugin getPluginParams
	//
	// Get plugin by name and version.
	//
	//     Consumes:
	//
	//     Produces:
	//     - application/json
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
			c.Error(err) //nolint
			return
		}

		if req.Name == nil && req.PluginType == nil {
			c.Error(errors.New("no params")) //nolint
			return
		}

		output, err := service.GetPluginDetail(c, req)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route GET /plugin/def plugin getPluginDef
	//
	// Get plugin by name and version.
	//
	//     Consumes:
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//     Deprecated: false
	//
	//     Responses:
	//       200: pluginDetail
	//		 503: apiError
	group.GET("def", func(c *gin.Context) {
		req := &repo.GetPluginDefParams{}
		err := c.ShouldBindQuery(req)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		pluginDef, err := service.GetPlugin(c, req.Name, req.Version)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.JSON(http.StatusOK, pluginDef)
	})

	// swagger:route GET /plugin/list plugin listPluginParams
	//
	// List plugin by name and version.
	//
	//     Consumes:
	//
	//     Produces:
	//     - application/json
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
			c.Error(err) //nolint
			return
		}

		if req.Name == nil && req.PluginType == nil {
			c.Error(errors.New("no params")) //nolint
			return
		}

		output, err := service.ListPlugin(c, req)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.JSON(http.StatusOK, output)
	})

	// swagger:route POST /plugin/label plugin addLabelParams
	//
	// Add label in plugin.
	//
	//     Consumes:
	//
	//     Produces:
	//     - application/json
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
			c.Error(err) //nolint
			return
		}

		if req.Name == nil && req.Label == nil {
			c.Error(errors.New("no params")) //nolint
			return
		}

		err = service.AddLabel(c, *req.Name, *req.Label)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.Status(http.StatusOK)
	})

	// swagger:route DELETE /plugin/label plugin deleteLabelParams
	//
	// Delete label in plugin.
	//
	//     Consumes:
	//
	//     Produces:
	//     - application/json
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
			c.Error(err) //nolint
			return
		}

		if req.Name == nil && req.Label == nil {
			c.Error(errors.New("no params")) //nolint
			return
		}

		err = service.DeleteLabel(c, *req.Name, *req.Label)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.Status(http.StatusOK)
	})

	// swagger:route DELETE /plugin plugin deletePluginReq
	//
	// Delete plugin by id and specific version
	//
	//     Consumes:
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//     Deprecated: false
	//
	//     Responses:
	//       200:
	//		 503: apiError
	group.DELETE("", func(c *gin.Context) {
		req := &models.DeletePluginByVersionReq{}
		err := c.ShouldBind(&req)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		params := &repo.DeletePluginParams{
			Version: req.Version,
		}
		params.ID, err = primitive.ObjectIDFromHex(req.ID)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		pluginDetail, err := service.GetPluginDetail(ctx, &repo.GetPluginParams{
			ID: &req.ID,
		})
		if err != nil {
			c.Error(err) //nolint
			return
		}

		plugin, err := service.GetPlugin(ctx, pluginDetail.Name, req.Version)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		err = service.DeletePluginByVersion(c, params)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		err = os.Remove(path.Join(string(pluginStore), plugin.Path))
		if err != nil {
			c.Error(err) //nolint
			return
		}

		c.Status(http.StatusOK)
	})

	// swagger:route DELETE /plugin/all plugin deletePluginReq
	//
	// Delete plugin all version
	//
	//     Consumes:
	//
	//     Produces:
	//     - application/json
	//
	//     Schemes: http, https
	//
	//     Deprecated: false
	//
	//     Responses:
	//       200:
	//		 503: apiError
	group.DELETE("all", func(c *gin.Context) {
		req := &models.DeletePluginReq{}
		err := c.ShouldBind(&req)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		pluginDetail, err := service.GetPluginDetail(ctx, &repo.GetPluginParams{
			ID: &req.ID,
		})
		if err != nil {
			c.Error(err) //nolint
			return
		}

		pluginId, err := primitive.ObjectIDFromHex(req.ID)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		err = service.DeletePlugin(c, pluginId)
		if err != nil {
			c.Error(err) //nolint
			return
		}

		for _, plugin := range pluginDetail.PluginDefs {
			err = os.Remove(path.Join(string(pluginStore), plugin.Path))
			if err != nil {
				c.Error(err) //nolint
				return
			}
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
	//     Consumes:
	//
	//     Produces:
	//     - application/json
	//
	// Responses:
	//	    200:
	//	    503: apiError
	group.POST("upload", func(c *gin.Context) {
		params := &UploadPluginFilesParams{}
		err := c.ShouldBind(params)
		// The file cannot be received.
		if err != nil {
			c.Error(err) //nolint
			return
		}

		for _, fileHeader := range params.PluginFiles {
			// The file is received, so let's save it
			tmpPath := path.Join(os.TempDir(), uuid.NewString())
			if err := c.SaveUploadedFile(fileHeader, tmpPath); err != nil {
				c.Error(err) //nolint
				return
			}

			err = os.Chmod(tmpPath, 0750)
			if err != nil {
				c.Error(err) //nolint
				return
			}

			pluginInfo, err := plugin.GetPluginInfo(tmpPath)
			if err != nil {
				c.Error(err) //nolint
				return
			}

			if err != nil && !errors.Is(err, repo.ErrPluginNotFound) {
				c.Error(err) //nolint
				return
			}

			// copy plugin to plugin store
			fname := fmt.Sprintf("%s_%s_%s", pluginInfo.PluginType, pluginInfo.Name, pluginInfo.Version)
			err = utils.CopyFile(tmpPath, filepath.Join(string(pluginStore), fname))
			if err != nil {
				c.Error(err) //nolint
				return
			}

			plugin := &models.PluginDef{
				PluginInfo: *pluginInfo,
				Path:       fname,
			}

			err = service.SavePlugins(c, plugin)
			if err != nil {
				c.Error(err) //nolint
				return
			}
		}

		c.Status(http.StatusOK)
	})

	// swagger:route POST /plugin/import plugin importPlugin
	//
	// import plugin.
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
			c.Error(err) //nolint
			return
		}

		for _, pluginPath := range filePaths {
			pluginInfo, err := plugin.GetPluginInfo(pluginPath)
			if err != nil {
				c.Error(err) //nolint
				return
			}

			if err != nil && !errors.Is(err, repo.ErrPluginNotFound) {
				c.Error(err) //nolint
				return
			}

			//copy plugin to plugin store
			fname := fmt.Sprintf("%s_%s_%s", pluginInfo.PluginType, pluginInfo.Name, pluginInfo.Version)
			err = utils.CopyFile(pluginPath, filepath.Join(string(pluginStore), fname))
			if err != nil {
				c.Error(err) //nolint
				return
			}

			plugin := &models.PluginDef{
				PluginInfo: *pluginInfo,
				Path:       fname,
			}

			err = service.SavePlugins(c, plugin)
			if err != nil {
				c.Error(err) //nolint
				return
			}
		}
		c.Status(http.StatusOK)
	})

	// swagger:route GET /download download downloadParams
	//
	// Download a directory as a zip file.
	//
	//     Consumes:
	//
	//     Produces:
	//     - application/zip
	//
	//     Schemes: http, https
	//
	//     Deprecated: false
	//
	//     Responses:
	//		 500: apiError
	group.GET("/download", func(c *gin.Context) {
		targetPath := "/root/brightbird/web/ui/public" // 需要打包的目标文件夹路径
		zipPath := "public.zip"                        // 打包后的 zip 文件名

		// Check if the zip file already exists
		if _, err := os.Stat(zipPath); err == nil {
			// If the zip file exists, remove it
			err = os.Remove(zipPath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else if !os.IsNotExist(err) {
			// If there was an error other than the file not existing, return the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err := archiver.Archive([]string{targetPath}, zipPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Disposition", `attachment; filename="`+filepath.Base(zipPath)+`"`)
		c.Header("Content-Type", "application/zip")
		c.File(zipPath)
	})
}
