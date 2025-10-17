package handler

import (
	"astron-xmod-shim/internal/core/orchestrator"
	dto "astron-xmod-shim/internal/dto/deploy"
	"astron-xmod-shim/pkg/log"
	"astron-xmod-shim/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DeleteServiceResponse Delete service response structure
type DeleteServiceResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ServiceID string `json:"serviceId"`
	} `json:"data"`
}

// GetServiceStatusResponse Get service status response structure
type GetServiceStatusResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ServiceID  string `json:"serviceId"`
		Status     string `json:"status"`   // Running/Blocked/Failed/Initializing/NotExists/Stopping
		Endpoint   string `json:"endpoint"` // openai like endpoint
		UpdateTime string `json:"updateTime"`
	} `json:"data"`
}

func DoDeploy(c *gin.Context) {
	var depSpec *dto.RequirementSpec
	if err := c.ShouldBindJSON(&depSpec); err != nil {
		log.Error("Failed to parse strategy request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"Code":    http.StatusBadRequest,
			"Message": "Invalid request parameters: " + err.Error(),
		})
		return
	}

	depSpec.ServiceId = utils.GenerateSimpleID()
	depSpec.GoalSetName = "opensource-llm-deploy"
	err := orchestrator.GlobalOrchestrator.Provision(depSpec)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "deploy submit failed: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "deploy submit success",
		"data":    map[string]string{"serviceId": depSpec.ServiceId},
	})
}

// GetServiceStatus Handle request to get model service status
func GetServiceStatus(c *gin.Context) {
	// Get serviceId from URL path
	serviceID := c.Param("serviceId")

	if serviceID == "" {
		log.Error("serviceId is required")
		response := GetServiceStatusResponse{
			Code:    1,
			Message: "serviceId is required",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	log.Info("Getting service status", "serviceID", serviceID)

	// Call orchestrator to get service status
	status, err := orchestrator.GlobalOrchestrator.GetServiceStatus(serviceID)
	if err != nil {
		log.Error("Get service status failed", "error", err)
		response := GetServiceStatusResponse{
			Code:    1,
			Message: "get service status failed",
			Data: struct {
				ServiceID  string `json:"serviceId"`
				Status     string `json:"status"`
				Endpoint   string `json:"endpoint"`
				UpdateTime string `json:"updateTime"`
			}{serviceID, "", "", ""},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Build OpenAI-style endpoint (should actually get from K8s service or configuration)

	// Get current time
	updateTime := time.Now().Format("2006-01-02 15:04:05")

	// Return successful response
	response := GetServiceStatusResponse{
		Code:    0,
		Message: "success",
		Data: struct {
			ServiceID  string `json:"serviceId"`
			Status     string `json:"status"`
			Endpoint   string `json:"endpoint"`
			UpdateTime string `json:"updateTime"`
		}{serviceID, string(status.Status), status.EndPoint, updateTime},
	}
	c.JSON(http.StatusOK, response)
}

// DeleteService Handle request to delete model service
func DeleteService(c *gin.Context) {
	// Get serviceId from URL path
	serviceID := c.Param("serviceId")

	if serviceID == "" {
		log.Error("serviceId is required")
		response := DeleteServiceResponse{
			Code:    1,
			Message: "serviceId is required",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	log.Info("Deleting service", "serviceID", serviceID)

	spec := &dto.RequirementSpec{ServiceId: serviceID, GoalSetName: "opensource-llm-delete", ResourceRequirements: &dto.ResourceRequirements{}}

	err := orchestrator.GlobalOrchestrator.Provision(spec)
	if err != nil {
		log.Error("Delete service failed", "error", err)
		response := DeleteServiceResponse{
			Code:    1,
			Message: "delete submit failed",
			Data: struct {
				ServiceID string `json:"serviceId"`
			}{serviceID},
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Return successful response
	response := DeleteServiceResponse{
		Code:    0,
		Message: "delete submit success",
		Data: struct {
			ServiceID string `json:"serviceId"`
		}{serviceID},
	}
	c.JSON(http.StatusOK, response)
}

// UpdateService Handle request to update model service
func UpdateService(c *gin.Context) {
	// Get serviceId from URL path
	serviceID := c.Param("serviceId")

	if serviceID == "" {
		log.Error("serviceId is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "serviceId is required",
		})
		return
	}

	var depSpec *dto.RequirementSpec
	if err := c.ShouldBindJSON(&depSpec); err != nil {
		log.Error("Failed to parse strategy request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request parameters: " + err.Error(),
		})
		return
	}

	// Use serviceId from URL instead of generating new
	depSpec.ServiceId = serviceID

	log.Info("Updating service", "serviceID", serviceID)
	depSpec.GoalSetName = "opensource-llm-deploy"
	// Reuse deployment logic for update
	err := orchestrator.GlobalOrchestrator.Provision(depSpec)
	if err != nil {
		log.Error("Update service failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "update submit failed: " + err.Error(),
			"data":    map[string]string{"serviceId": serviceID},
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "update submit success",
		"data":    map[string]string{"serviceId": serviceID},
	})
}
