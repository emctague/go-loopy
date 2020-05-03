package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"log"
	"strconv"
)

func main() {
	pixelgl.Run(func() {
		cfg := pixelgl.WindowConfig{Title: "PixelGL!!!", Bounds: pixel.R(0, 0, 1024, 768), VSync: true}

		win, err := pixelgl.NewWindow(cfg)
		if err != nil {
			log.Fatal(err)
		}

		e := NewECS()

		// Add all systems
		TransformSystem(&e)
		PhysicsSystem(&e, win)
		PlayerSystem(&e, win)
		ParticleSystem(&e)
		BalanceSystem(&e)
		InteractiveSystem(&e, win)

		// The render system needs to run on the main thread, so we let it transfer our setup to a goroutine.
		RenderSystem(&e, win, func() {

			// Create player entity. The wallet is stored separately so that it can be interacted with from the NPC
			// scripts provided below.
			pWallet := &Wallet{100}
			player := e.AddEntity(&Transform{20, 20, 0},
				pWallet,
				&Physics{0, 0},
				&Player{}, &Interactor{},
				&DebugCircle{colornames.Green, 10})

			// ESPECIALLY UGLY CODE AHEAD
			// YOU HAVE BEEN WARNED

			// Add an interactve entity with an ugly hardcoded dialog script
			e.AddEntity(&Transform{200, 200, 0}, &Interactive{
				"[space] talk", "Alice",
				IMenu("Hi, what's your name?",
					"Ethan", IMenu("I'm not sure I believe you...", "*leave*", nil),
					"Alice", IMenu("Hey, that's *my* name!", "*leave*", nil),
					"None of your business!", func(ev EventContainer) *InteractionMenu {
						ev.Next <- BalanceChangeEvent{player, -pWallet.Balance}
						return IMenu("OK, jerk.",
							"...", IMenu("Just for that...",
								"...", IMenu("I'm stealing your wallet.", "*leave*", nil)))(ev)
					}),
			}, &DebugCircle{colornames.Aliceblue, 10})

			// Another ugly interactive entity script
			var watchuWant func(EventContainer) *InteractionMenu
			watchuWant = IMenu("What would you like??",
				"Cash", func(ev EventContainer) *InteractionMenu {
					ev.Next <- BalanceChangeEvent{player, 200}
					return IMenu("...fine. Here's $200. Your balance is now $"+strconv.Itoa(pWallet.Balance+200),
						"Wow, awesome!", watchuWant)(ev)
				},
				"Food", func(ev EventContainer) *InteractionMenu {
					if pWallet.Balance >= 20 {
						ev.Next <- BalanceChangeEvent{player, -20}
						return IMenu("Here you go! Your balance is now $"+strconv.Itoa(pWallet.Balance-20), "Wow, thanks!", watchuWant)(ev)
					} else {
						return IMenu("Man, you're broke! You only have $"+strconv.Itoa(pWallet.Balance), "...ok?", watchuWant)(ev)
					}
				},
				"*leave*", nil,
			)

			e.AddEntity(&Transform{500, 300, 0}, &DebugCircle{colornames.Goldenrod, 10},
				&Interactive{"[space] talk", "Rod", watchuWant})

			e.Run()
		})
	})
}
