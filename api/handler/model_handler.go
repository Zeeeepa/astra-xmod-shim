package handler

import (
	"astron-xmod-shim/internal/config"
	"astron-xmod-shim/pkg/log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// ModelInfo Model information structure
type ModelInfo struct {
	ModelName string `json:"modelName"`
	ModelPath string `json:"modelPath"`
}

// ModelListResponse Model list response structure
type ModelListResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    []ModelInfo `json:"data"`
}

func ListModel(c *gin.Context) {
	// Get global configuration
	conf := config.Get()

	// Get model root directory from configuration
	modelsRootDir := conf.ModelManage.ModelRoot
	if modelsRootDir == "" {
		modelsRootDir = "/models"
		log.Info("Using default model root directory: %s", modelsRootDir)
	}

	log.Info("Starting to get model list from root directory %s", modelsRootDir)

	// Check if directory exists and read contents
	entries, err := os.ReadDir(modelsRootDir)
	if err != nil {
		log.Error("Failed to read model directory: %v", err)
		c.JSON(http.StatusInternalServerError, ModelListResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to read model directory",
			Data:    []ModelInfo{},
		})
		return
	}

	// Collect model information (consider only directories)
	models := make([]ModelInfo, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			models = append(models, ModelInfo{
				ModelName: entry.Name(),
				ModelPath: filepath.Join(modelsRootDir, entry.Name()),
			})
		}
	}

	log.Info("Successfully retrieved %d models", len(models))

	// Return model list
	c.JSON(http.StatusOK, ModelListResponse{
		Code:    0,
		Message: "success",
		Data:    models,
	})
}