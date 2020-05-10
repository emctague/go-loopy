package systems

import (
	"fmt"
	"github.com/emctague/go-loopy/ecs"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"os"
)

// HUDLine is a component which provides a line of text on-screen above all other content.
type HUDLine struct {
	Prompt   string  // The contents of the HUD line.
	Centered bool    // Whether or not the line is centered horizontally on the transform position.
	FontSize float64 // The font size as a multiplier.
	width    float64 // The width of the HUDLine with its current prompt value. This is used for centering.
}

// Renderable is a component which defines a colored circle to be drawn on-screen by the renderer.
type Renderable struct {
	Sprite *pixel.Sprite
}

// ChangeHUDPromptEvent represents a request to change the prompt string of a HUDLine.
type ChangeHUDPromptEvent struct {
	ID     uint64
	Prompt string
}

type eDebugRenderable struct {
	*Renderable
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

			win.Clear(color.RGBA{R: 0, G: 0, B: 0, A: 255})

			// Draw all debug circles
			for _, renderable := range debugRenderables {
				renderable.Sprite.Draw(win, pixel.IM.Rotated(pixel.V(0, 0), renderable.Rotation).Moved(pixel.V(renderable.X, renderable.Y)))
			}

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

// From pixelGL tutorials
func LoadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}
