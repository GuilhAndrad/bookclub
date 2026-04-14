package pagination

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultPage é a página inicial quando nenhuma é informada.
	DefaultPage = 1

	// DefaultLimit é o tamanho de página padrão.
	DefaultLimit = 20

	// MaxLimit é o tamanho máximo permitido para evitar queries excessivas.
	MaxLimit = 100
)

// Params contém os parâmetros de paginação extraídos de uma requisição HTTP.
type Params struct {
	// Page é o número da página solicitada (base 1).
	Page int

	// Limit é o número de itens por página.
	Limit int
}

// Offset retorna o deslocamento SQL correspondente a esses parâmetros.
func (p Params) Offset() int {
	return (p.Page - 1) * p.Limit
}

// Page representa uma página de resultados de qualquer tipo T.
type Page[T any] struct {
	// Items contém os registros da página atual.
	Items []T `json:"items"`

	// Total é o número total de registros sem paginação.
	Total int64 `json:"total"`

	// Page é a página atual.
	Page int `json:"page"`

	// Limit é o tamanho da página.
	Limit int `json:"limit"`

	// TotalPages é o número total de páginas calculado a partir de Total e Limit.
	TotalPages int `json:"total_pages"`

	// HasNext indica se existe uma próxima página.
	HasNext bool `json:"has_next"`

	// HasPrev indica se existe uma página anterior.
	HasPrev bool `json:"has_prev"`
}

// New constrói uma Page[T] a partir dos itens, total e parâmetros de paginação.
func New[T any](items []T, total int64, p Params) Page[T] {
	totalPages := int(math.Ceil(float64(total) / float64(p.Limit)))
	if totalPages == 0 {
		totalPages = 1
	}
	return Page[T]{
		Items:      items,
		Total:      total,
		Page:       p.Page,
		Limit:      p.Limit,
		TotalPages: totalPages,
		HasNext:    p.Page < totalPages,
		HasPrev:    p.Page > 1,
	}
}

// FromRequest extrai e valida os parâmetros de paginação de uma requisição Gin.
// Retorna DefaultPage e DefaultLimit quando os parâmetros estiverem ausentes.
// Retorna 400 Bad Request e interrompe a cadeia se os valores forem inválidos.
func FromRequest(c *gin.Context) (Params, bool) {
	p := Params{Page: DefaultPage, Limit: DefaultLimit}

	if pageStr := c.Query("page"); pageStr != "" {
		v, err := strconv.Atoi(pageStr)
		if err != nil || v < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "page deve ser um inteiro maior que 0"})
			return p, false
		}
		p.Page = v
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		v, err := strconv.Atoi(limitStr)
		if err != nil || v < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit deve ser um inteiro maior que 0"})
			return p, false
		}
		if v > MaxLimit {
			v = MaxLimit
		}
		p.Limit = v
	}

	return p, true
}
