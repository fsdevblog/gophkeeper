package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/fsdevblog/gophkeeper/internal/transport/http/dto"

	"github.com/fsdevblog/gophkeeper/internal/pag"
	"github.com/fsdevblog/gophkeeper/internal/transport/http/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EntriesHandler struct {
	entryProvider EntryProvider
}

func NewEntriesHandler(provider EntryProvider) *EntriesHandler {
	return &EntriesHandler{
		entryProvider: provider,
	}
}

// GetAll GET /api/entries.
func (e *EntriesHandler) GetAll(c *gin.Context) {
	currentUserID, _ := c.Get(middlewares.CurrentUserIDKey)
	currentUserUUID, _ := currentUserID.(uuid.UUID)
	pagination := pag.New(func(options *pag.Options) {
		cpInt, _ := strconv.ParseInt(c.Query("page"), 10, 32)
		ppInt, _ := strconv.ParseInt(c.Query("per_page"), 10, 32)
		options.CurrentPage = int32(cpInt)
		options.PerPage = int32(ppInt)
	})

	ctx, cancel := context.WithTimeout(c, DefaultServiceTimeout)
	defer cancel()

	entries, totalRecords, err := e.entryProvider.GetUserEntries(ctx, currentUserUUID, pagination)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err).
			SetType(gin.ErrorTypePrivate)
		return
	}
	var response = make([]dto.EntryResponseItem, len(entries))
	for i, entry := range entries {
		response[i] = dto.EntryResponseItem{
			ID:        entry.ID,
			Title:     entry.Title,
			EntryType: entry.EntryType,
		}
	}

	c.JSON(http.StatusOK, gin.H{"entries": response, "meta": dto.PaginationMeta{
		TotalRecords: totalRecords,
		CurrentPage:  pagination.Page(),
		PerPage:      pagination.PerPage(),
	}})
}

// GetEntryFields GET /api/entries/:entryID/fields.
func (e *EntriesHandler) GetEntryFields(c *gin.Context) {
	curUserNoTyped, _ := c.Get(middlewares.CurrentUserIDKey)
	currentUserUUID, _ := curUserNoTyped.(uuid.UUID)

	entryID, errParse := uuid.Parse(c.Param("entryID"))
	if errParse != nil {
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("param \"entryID\" is invalid")).
			SetType(gin.ErrorTypePublic)
		return
	}
	ctx, cancel := context.WithTimeout(c, DefaultServiceTimeout)
	defer cancel()

	dbFields, err := e.entryProvider.GetSafeEntryFields(ctx, currentUserUUID, entryID)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err).
			SetType(gin.ErrorTypePrivate)
		return
	}

	var response = make([]dto.EntryFieldResponseItem, len(dbFields))

	for i, field := range dbFields {
		response[i] = dto.EntryFieldResponseItem{
			ID:        field.ID,
			EntryID:   field.EntryID,
			Key:       field.Key,
			Value:     field.Value,
			IsPrivate: field.IsPrivate,
		}
	}

	c.JSON(http.StatusOK, response)
}
