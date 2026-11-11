package action

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/responder"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/messaging"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/query"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/feed"
)

const (
	feedCacheControl = "public, max-age=300"
	maxVisitBody     = 1024
)

type feedPayload struct {
	Items []feed.Item `json:"items"`
}

func HandleListRepositories(queryBus messaging.QueryBus) gin.HandlerFunc {
	return func(c *gin.Context) {
		respondWithFeed(c, queryBus, &query.ListRepositories{}, "repositories")
	}
}

func HandleListPackages(queryBus messaging.QueryBus) gin.HandlerFunc {
	return func(c *gin.Context) {
		respondWithFeed(c, queryBus, &query.ListPackages{}, "packages")
	}
}

func respondWithFeed(c *gin.Context, queryBus messaging.QueryBus, q messaging.Query, name string) {
	result, err := queryBus.Fetch(c.Request.Context(), q)
	if err != nil {
		log.Printf("feed %s unavailable: %v", name, err)
		responder.BadGatewayResponse(c, "feed unavailable")
		return
	}

	items, ok := result.([]feed.Item)
	if !ok {
		log.Printf("feed %s returned unexpected result type %T", name, result)
		responder.BadGatewayResponse(c, "feed unavailable")
		return
	}

	c.Header("Cache-Control", feedCacheControl)
	responder.OKResponse(c, feedPayload{Items: items})
}

type visitRequest struct {
	Language  string `json:"language"`
	UserAgent string `json:"userAgent"`
}

func HandleRecordVisit(commandBus messaging.CommandBus) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body visitRequest
		limited := http.MaxBytesReader(c.Writer, c.Request.Body, maxVisitBody)
		if err := decodeOptionalJSON(limited, &body); err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				responder.PayloadTooLargeResponse(c, "request body too large")
				return
			}
			responder.BadRequestResponse(c, errors.New("invalid request body"))
			return
		}

		if body.Language == "" {
			body.Language = c.GetHeader("Accept-Language")
		}
		if body.UserAgent == "" {
			body.UserAgent = c.GetHeader("User-Agent")
		}

		err := commandBus.Execute(c.Request.Context(), &command.RecordVisit{
			Language:  body.Language,
			UserAgent: body.UserAgent,
			ClientIP:  c.ClientIP(),
		})
		if err != nil {
			log.Printf("record visit failed: %v", err)
		}

		responder.NoContentResponse(c)
	}
}

func decodeOptionalJSON(r io.Reader, out any) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return nil
	}

	return json.Unmarshal(raw, out)
}
