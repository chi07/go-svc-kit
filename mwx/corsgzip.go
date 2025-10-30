// pkg/mwx/corsgzip.go
package mw

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CORS(originsCSV string) fiber.Handler {
	o := strings.TrimSpace(originsCSV)
	if o == "" {
		o = "*"
	}
	return cors.New(cors.Config{
		AllowOrigins:     o,
		AllowMethods:     "GET,POST,PATCH,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Authorization,Content-Type,Accept,Origin,X-Requested-With,X-API-Key",
		ExposeHeaders:    "Content-Length,Content-Encoding,ETag",
		AllowCredentials: true,
		MaxAge:           int((12 * time.Hour).Seconds()),
	})
}
func Gzip() fiber.Handler {
	return compress.New(compress.Config{Level: compress.LevelBestSpeed})
}
