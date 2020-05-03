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

type HUDLine struct {
	Prompt   string
	Centered bool
	FontSize float64
	Width    float64
}

type DebugCircle struct {
	Color  color.Color
	Radius float64
}

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

			for _, e := range debugRenderables {
				imd.Color = e.Color
				imd.Push(pixel.V(e.X, e.Y))
				imd.Circle(e.Radius, 0)

			}

			imd.Draw(win)

			for _, hudLine := range hudLines {
				txt.Clear()

				if hudLine.Centered {
					txt.Dot.X -= hudLine.Width / 2
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
				line.Width = 0
				break
			}

			txt.Clear()
			line.Width = txt.BoundsOf(line.Prompt).W()

		}

		ev.Wg.Done()
	}

}
