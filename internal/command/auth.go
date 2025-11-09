package command

import (
	"assistant-sf/internal/config"
	"context"
	"fmt"
)

func AuthRun(ctx context.Context) error {
	cnf := config.MustLoad(ctx)

	fmt.Printf("%+v\n", cnf)
	return nil
}
