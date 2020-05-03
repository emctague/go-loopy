package main

import (
	"github.com/emctague/go-loopy/ecs"
	"github.com/emctague/go-loopy/systems"
	"github.com/emctague/go-loopy/utils"
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

			e.AddEntity(&systems.Transform{X: 200, Y: 200}, &systems.Interactive{
				Prompt: "[space] Talk", Name: "Alice",
				Menu: utils.MakeDialogScript(func(prompt utils.PromptTool, ev **ecs.EventContainer) {
					switch prompt("Hi, what's your name?", "Ethan", "Alice") {
					case 0:
						prompt("I'm not sure I believe you!", "...ok?")

					case 1:
						switch prompt("Hey, that's *my* name!", "Well it's mine too!", "uh... nice to know") {
						case 0:
							prompt("Fineeee, we can share...", "...Bye!")
						case 1:
							prompt("Yeah, isn't it?", "...I am so confused...")
						}
					}
				}),
			}, &systems.DebugCircle{Color: colornames.Aliceblue, Radius: 10})

			e.AddEntity(&systems.Transform{X: 500, Y: 300}, &systems.DebugCircle{Color: colornames.Goldenrod, Radius: 10},
				&systems.Interactive{
					Prompt: "[space] talk", Name: "Rod",
					Menu: utils.MakeDialogScript(func(prompt utils.PromptTool, ev **ecs.EventContainer) {
						for {
							switch prompt("What would you like?", "One million dollars!", "Food.", "For you to go away, weirdo...") {
							case 0:
								(*ev).Next <- systems.BalanceChangeEvent{ID: player, Change: 50}
								prompt("Here's 50, stop complaining.", "...Fine")

							case 1:
								if pWallet.Balance >= 20 {
									(*ev).Next <- systems.BalanceChangeEvent{ID: player, Change: -20}
									prompt("Here you go! Your balance is now $"+strconv.Itoa(pWallet.Balance-20), "Wow, thanks!")
								} else {
									prompt("You're just too damn broke.", "...Oh")
								}

							case 2:
								return
							}
						}
					}),
				})

			e.Run()
		})
	})
}
