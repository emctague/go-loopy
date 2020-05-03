package main

import (
	"github.com/emctague/go-loopy/ecs"
	"github.com/emctague/go-loopy/systems"
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

		e := ecs.NewECS()

		// Add all systems
		systems.TransformSystem(&e)
		systems.PhysicsSystem(&e, win)
		systems.PlayerSystem(&e, win)
		systems.ParticleSystem(&e)
		systems.BalanceSystem(&e)
		systems.InteractiveSystem(&e, win)

		// The render system needs to run on the main thread, so we let it transfer our setup to a goroutine.
		systems.RenderSystem(&e, win, func() {

			// Create player entity. The wallet is stored separately so that it can be interacted with from the NPC
			// scripts provided below.
			pWallet := &systems.Wallet{Balance: 100}
			player := e.AddEntity(&systems.Transform{X: 20, Y: 20},
				pWallet,
				&systems.Physics{},
				&systems.Player{}, &systems.Interactor{},
				&systems.DebugCircle{Color: colornames.Green, Radius: 10})

			// ESPECIALLY UGLY CODE AHEAD
			// YOU HAVE BEEN WARNED

			// Add an interactve entity with an ugly hardcoded dialog script
			e.AddEntity(&systems.Transform{200, 200, 0}, &systems.Interactive{
				Prompt: "[space] talk", Name: "Alice",
				Menu: systems.IMenu("Hi, what's your name?",
					"Ethan", systems.IMenu("I'm not sure I believe you...", "*leave*", nil),
					"Alice", systems.IMenu("Hey, that's *my* name!", "*leave*", nil),
					"None of your business!", func(ev ecs.EventContainer) *systems.InteractionMenu {
						ev.Next <- systems.BalanceChangeEvent{ID: player, Change: -pWallet.Balance}
						return systems.IMenu("OK, jerk.",
							"...", systems.IMenu("Just for that...",
								"...", systems.IMenu("I'm stealing your wallet.", "*leave*", nil)))(ev)
					}),
			}, &systems.DebugCircle{Color: colornames.Aliceblue, Radius: 10})

			// Another ugly interactive entity script
			var watchuWant func(ecs.EventContainer) *systems.InteractionMenu
			watchuWant = systems.IMenu("What would you like??",
				"Cash", func(ev ecs.EventContainer) *systems.InteractionMenu {
					ev.Next <- systems.BalanceChangeEvent{ID: player, Change: 200}
					return systems.IMenu("...fine. Here's $200. Your balance is now $"+strconv.Itoa(pWallet.Balance+200),
						"Wow, awesome!", watchuWant)(ev)
				},
				"Food", func(ev ecs.EventContainer) *systems.InteractionMenu {
					if pWallet.Balance >= 20 {
						ev.Next <- systems.BalanceChangeEvent{ID: player, Change: -20}
						return systems.IMenu("Here you go! Your balance is now $"+strconv.Itoa(pWallet.Balance-20), "Wow, thanks!", watchuWant)(ev)
					}

					return systems.IMenu("Man, you're broke! You only have $"+strconv.Itoa(pWallet.Balance), "...ok?", watchuWant)(ev)
				},
				"*leave*", nil,
			)

			e.AddEntity(&systems.Transform{X: 500, Y: 300}, &systems.DebugCircle{Color: colornames.Goldenrod, Radius: 10},
				&systems.Interactive{Prompt: "[space] talk", Name: "Rod", Menu: watchuWant})

			e.Run()
		})
	})
}
