package usecasex

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

type FetchManyFn func(ctx context.Context, ids []string) (map[string]any, error)

func EnrichBatch(ctx context.Context, timeout time.Duration, log zerolog.Logger, tasks map[string]func(context.Context) error) {
	eg := new(errgroup.Group)
	for name, fn := range tasks {
		name, fn := name, fn
		eg.Go(func() error {
			cctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			if err := fn(cctx); err != nil {
				log.Warn().Str("task", name).Err(err).Msg("enrich_failed")
			}
			return nil
		})
	}
	_ = eg.Wait()
}
