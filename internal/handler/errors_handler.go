package handler

import (
	"errors"
	"net/http"

	"github.com/GuilhAndrad/bookclub/pkg/apperr"
	"github.com/gin-gonic/gin"
)

// respondError converte um erro de domínio no status HTTP correspondente.
// Centralizado aqui para que todos os handlers usem a mesma conversão.
func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, apperr.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, apperr.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, apperr.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, apperr.ErrPendingApproval):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, apperr.ErrAccountRejected):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro interno"})
	}
}