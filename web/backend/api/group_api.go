package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hunjixin/brightbird/repo"
	"github.com/hunjixin/brightbird/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// updateGroupRequest
// swagger:model updateGroupRequest
type UpdateGroupRequest struct {
	Name        string `json:"name"`
	IsShow      bool   `json:"isShow"`
	Description string `json:"description"`
}

// GroupResp
// swagger:model groupResp
type GroupResp struct {
	*types.Group
	TestFlowCount int `json:"testFlowCount"`
}

// ListGroupResp
// swagger:model listGroupResp
type ListGroupResp []GroupResp

func RegisterGroupRouter(ctx context.Context, v1group *V1RouterGroup, groupSvc repo.IGroupRepo, testFlowSvc repo.ITestFlowRepo) {
	group := v1group.Group("/group")

	// swagger:route GET /group listGroup
	//
	// Lists all group.
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
	//       200: listGroupResp
	group.GET("", func(c *gin.Context) {
		groups, err := groupSvc.List(ctx)
		if err != nil {
			c.Error(err)
			return
		}
		groupOutList := make([]GroupResp, len(groups))
		for i, group := range groups {
			count, err := testFlowSvc.CountByGroup(ctx, group.ID)
			if err != nil {
				c.Error(err)
				return
			}

			groupOutList[i] = GroupResp{
				Group:         group,
				TestFlowCount: int(count),
			}
		}
		c.JSON(http.StatusOK, groupOutList)
	})

	// swagger:route Get /group/{id} getTestFlow
	//
	// Get specific group by id.
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
	//         in: path
	//         description: name of test flow
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200: groupResp
	group.GET(":id", func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err)
			return
		}

		group, err := groupSvc.Get(ctx, id)
		if err != nil {
			c.Error(err)
			return
		}

		count, err := testFlowSvc.CountByGroup(ctx, group.ID)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, GroupResp{
			Group:         group,
			TestFlowCount: int(count),
		})
	})

	// swagger:route POST /group saveCases
	//
	// Save group
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
	//       + name: group
	//         in: body
	//         description: group json
	//         required: true
	//         type: group
	//         allowEmpty:  false
	//
	//     Responses:
	//       200:
	group.POST("", func(c *gin.Context) {
		testFlow := types.Group{}
		err := c.ShouldBindJSON(&testFlow)
		if err != nil {
			c.Error(err)
			return
		}

		id, err := groupSvc.Save(ctx, testFlow)
		if err != nil {
			c.Error(err)
			return
		}

		c.String(http.StatusOK, id.String())
	})

	// swagger:route POST /group/{id} updateGroup
	//
	// Update group name/show/description
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
	//         in: path
	//         description: id of  group
	//         required: true
	//         type: string
	//       + name: group
	//         in: body
	//         description: update group request json
	//         required: true
	//         type: updateGroupRequest
	//         allowEmpty:  false
	//
	//     Responses:
	//       200:
	group.POST("/:id", func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err)
			return
		}

		req := UpdateGroupRequest{}
		err = c.ShouldBindJSON(&req)
		if err != nil {
			c.Error(err)
			return
		}

		group, err := groupSvc.Get(c, id)
		if err != nil {
			c.Error(err)
			return
		}

		group.Name = req.Name
		group.Description = req.Description
		group.IsShow = req.IsShow
		group.ModifiedTime = time.Now().Unix()

		_, err = groupSvc.Save(ctx, *group)
		if err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusOK)
	})
	// swagger:route DELETE /group/{id} deleteGroup
	//
	// Delete group by id
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
	//         in: path
	//         description: id of  group
	//         required: true
	//         type: string
	//
	//     Responses:
	//       200:
	group.DELETE("/:id", func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.Error(err)
			return
		}
		err = groupSvc.Delete(c, id)
		if err != nil {
			c.Error(err)
			return
		}

		c.Status(http.StatusOK)
	})
}
