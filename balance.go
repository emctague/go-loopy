package main

import (
	"log"
)

type Wallet struct {
	Balance int
}

type BalanceChangeEvent struct {
	ID     uint64
	Change int
}

func BalanceSystem(e *ECS) {
	wallets := make(map[uint64]struct {
		*Wallet
	})

	events := e.Subscribe()

	go func() {

		for ev := range events {
			switch event := ev.Event.(type) {

			case EntityAddedEvent:
				UnpackEntity(event, &wallets)

			case EntityRemovedEvent:
				RemoveEntity(event.ID, &wallets)

			case BalanceChangeEvent:
				wallet, ok := wallets[event.ID]
				if !ok {
					log.Fatal("Trying to change balance of nonexistent wallet")
				}

				wallet.Balance += event.Change
			}

			ev.Wg.Done()
		}
	}()
}
