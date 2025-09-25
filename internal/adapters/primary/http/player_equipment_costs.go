package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ab-dauletkhan/coc/internal/application/usecases"
)

type PlayerEquipmentCostsHandler struct {
	uc *usecases.PlayerEquipmentCostsUseCase
}

func NewPlayerEquipmentCostsHandler(uc *usecases.PlayerEquipmentCostsUseCase) *PlayerEquipmentCostsHandler {
	return &PlayerEquipmentCostsHandler{uc: uc}
}

func (h *PlayerEquipmentCostsHandler) Register(r *gin.Engine) {
	r.GET("/v1/players/:tag/hero-equipments/costs", h.get)
}

func (h *PlayerEquipmentCostsHandler) get(c *gin.Context) {
	tag := c.Param("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing tag"})
		return
	}
	nTag := normalizePlayerTag(tag)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Second)
	defer cancel()

	res, status, err := h.uc.Execute(ctx, nTag)
	if err != nil {
		if status == 0 {
			status = http.StatusBadGateway
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	if status >= 400 {
		c.Status(status)
		return
	}
	c.JSON(http.StatusOK, res)
}
