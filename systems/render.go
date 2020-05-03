package systems

import (
	"fmt"
	"github.com/emctague/go-loopy/ecs"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
	"image/color"
	"log"
)

// HUDLine is a component which provides a line of text on-screen above all other content.
type HUDLine struct {
	Prompt   string  // The contents of the HUD line.
	Centered bool    // Whether or not the line is centered horizontally on the transform position.
	FontSize float64 // The font size as a multiplier.
	width    float64 // The width of the HUDLine with its current prompt value. This is used for centering.
}

// DebugCircle is a component which defines a colored circle to be drawn on-screen by the renderer.
type DebugCircle struct {
	Color  color.Color
	Radius float64
}

// ChangeHUDPromptEvent represents a request to change the prompt string of a HUDLine.
type ChangeHUDPromptEvent struct {
	ID     uint64
	Prompt string
}

type eDebugRenderable struct {
	*DebugCircle
	*Transform
}

type eHudText struct {
	*HUDLine
	*Transform
}

// RenderSystem is a system which draws to the screen.
// Unlike other systems, the RenderSystem does not run itself in a goroutine - because PixelGL requires rendering to
// occur in the main thread, RenderSystem takes it over and runs the passed whenReady function in a goroutine, where
// the user of the RenderSystem may continue setup.
func RenderSystem(e *ecs.ECS, win *pixelgl.Window, whenReady func()) {
	imd := imdraw.New(nil)

	debugRenderables := make(map[uint64]eDebugRenderable)
	hudLines := make(map[uint64]eHudText)

	events := e.Subscribe()

	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pixel.V(0, 0), atlas)

	go whenReady()

	for ev := range events {
		switch event := ev.Event.(type) {
		case ecs.EntityAddedEvent:
			ecs.UnpackEntity(event, &debugRenderables)
			ecs.UnpackEntity(event, &hudLines)

		case ecs.EntityRemovedEvent:
			ecs.RemoveEntity(event.ID, &debugRenderables)
			ecs.RemoveEntity(event.ID, &hudLines)

		case ecs.UpdateBeginEvent:

			if win.Closed() {
				e.Stop()
			}

		case ecs.UpdateEndEvent:

			win.Clear(color.RGBA{A: 255})

			imd.Clear()

			// Draw all debug circles
			for _, renderable := range debugRenderables {
				imd.Color = renderable.Color
				imd.Push(pixel.V(renderable.X, renderable.Y))
				imd.Circle(renderable.Radius, 0)
			}

			imd.Draw(win)

			// Draw all HUD lines
			for _, hudLine := range hudLines {
				txt.Clear()

				if hudLine.Centered {
					txt.Dot.X -= hudLine.width / 2
				}

				_, _ = fmt.Fprintln(txt, hudLine.Prompt)

				win.SetMatrix(pixel.IM.Moved(pixel.V(hudLine.X, hudLine.Y)))
				txt.Draw(win, pixel.IM.Scaled(txt.Orig, hudLine.FontSize))
			}

			win.SetMatrix(pixel.IM)
			win.Update()

		case ChangeHUDPromptEvent:
			line, ok := hudLines[event.ID]
			if !ok {
				log.Fatal("Cannot change prompt on an entity with no HUDLine component")
			}

			if event.Prompt == line.Prompt {
				break
			}

			line.HUDLine.Prompt = event.Prompt

			if event.Prompt == "" {
				line.width = 0
				break
			}

			txt.Clear()
			line.width = txt.BoundsOf(line.Prompt).W()

		}

		ev.Wg.Done()
	}

}
