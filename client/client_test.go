package client_test

import (
	"context"
	"testing"

	"github.com/lestrrat-go/gnap"
	"github.com/lestrrat-go/gnap/client"
)

func TestGrantRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cl := client.New()
	cl.NewGrantRequest().Interact(
		gnap.NewInteraction(gnap.StartRedirect).
			AddFinish(gnap.NewInteractionFinish(gnap.FinishRedirect, `1234567890`, `https://localhost:8080/finish`)),
	).Do(ctx)
}
