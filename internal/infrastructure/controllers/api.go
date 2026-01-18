package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctx *AppContext) GetGameHandler(c *gin.Context) {
	c.Abort()
}

func (ctx *AppContext) PostGameHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (ctx *AppContext) PatchGameSettingsHandler(c *gin.Context) {

}
