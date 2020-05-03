package systems

import (
	"github.com/emctague/go-loopy/ecs"
	"log"
)

// Wallet is a component which stores the monetary balance of an entity.
type Wallet struct {
	Balance int
}

// BalanceChangeEvent represents a change in the balance of an entity's wallet.
type BalanceChangeEvent struct {
	ID     uint64
	Change int
}

// BalanceSystem handles wallets and balance change events, keeping track of in-game currency.
func BalanceSystem(e *ecs.ECS) {
	wallets := make(map[uint64]struct {
		*Wallet
	})

	events := e.Subscribe()

	go func() {

		for ev := range events {
			switch event := ev.Event.(type) {

			case ecs.EntityAddedEvent:
				ecs.UnpackEntity(event, &wallets)

			case ecs.EntityRemovedEvent:
				ecs.RemoveEntity(event.ID, &wallets)

			case BalanceChangeEvent:
				wallet, ok := wallets[event.ID]
				if !ok {
					log.Fatal("Trying to change balance of nonexistent wallet")
				}

				wallet.Balance += event.Change
			}

			ev.Done()
		}
	}()
}
