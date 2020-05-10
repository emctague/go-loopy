package systems

import (
	"github.com/emctague/go-loopy/ecs"
	"github.com/faiface/pixel/pixelgl"
	"math"
	"strconv"
)

// MenuChoice represents one choice in an interactive menu.
// It has a label to describe what the interactor is selecting / saying, and a function which performs some actions and
// then returns another menu, or nil to exit the menus.
// The Action function itself may also be nil to just exit when selected.
type MenuChoice struct {
	Label  string
	Action func(ecs.EventContainer) *InteractionMenu
}

// InteractionMenu represents a menu with a prompt and several selectable options.
type InteractionMenu struct {
	Prompt  string
	Choices []MenuChoice
}

// Interactive is a component placed upon entities that can be interacted with by an interactor, resulting in some menu
// appearing.
type Interactive struct {
	Prompt string                                    // The prompt line describes the action and trigger, e.g. "[space] Talk"
	Name   string                                    // The in-world name of the entity, e.g. "Jeff".
	Menu   func(ecs.EventContainer) *InteractionMenu // A function that performs some action and opens a menu.
}

// Interactor is a component placed upon entities that can interact with others, interrupting its flow with a menu.
type Interactor struct {
	InMenu            bool             // True if a menu is currently active.
	Menu              *InteractionMenu // A pointer to the currently active menu.
	NearbyInteractive uint64           // The ID of a nearby interactive, or 0 if nothing is in range
}

type eInteractor struct {
	*Transform
	*Interactor
}
type eInteractive struct {
	*Transform
	*Interactive
}

type interactiveContext struct {
	primaryLabel  *HUDLine
	ePrimaryLabel uint64

	secondaryLabel  *HUDLine
	tSecondaryLabel *Transform
	eSecondaryLabel uint64

	interactors  map[uint64]eInteractor
	interactives map[uint64]eInteractive

	events chan ecs.EventContainer

	e   *ecs.ECS
	win *pixelgl.Window
}

// InteractiveSystem handles interactive in-game menus.
func InteractiveSystem(e *ecs.ECS, win *pixelgl.Window) {

	var ctx = interactiveContext{
		primaryLabel: &HUDLine{Centered: true, FontSize: 2},

		secondaryLabel:  &HUDLine{Centered: true, FontSize: 1.5},
		tSecondaryLabel: &Transform{},

		interactors:  make(map[uint64]eInteractor),
		interactives: make(map[uint64]eInteractive),

		events: e.Subscribe(),

		e:   e,
		win: win,
	}

	go func() {
		for ev := range ctx.events {
			switch event := ev.Event.(type) {
			case ecs.SetupEvent:
				ctx.eSecondaryLabel = e.AddEntity(ctx.secondaryLabel, ctx.tSecondaryLabel)
				ctx.ePrimaryLabel = e.AddEntity(ctx.primaryLabel, &Transform{0, 20, 0, 0, 0, ctx.eSecondaryLabel})

			case ecs.EntityAddedEvent:
				ecs.UnpackEntity(event, &ctx.interactors)
				ecs.UnpackEntity(event, &ctx.interactives)

			case ecs.EntityRemovedEvent:
				ecs.RemoveEntity(event.ID, &ctx.interactors)
				ecs.RemoveEntity(event.ID, &ctx.interactives)

			case ecs.UpdateBeginEvent:

				for _, interactor := range ctx.interactors {

					// Deal with the interactor differently if it's already in a menu.
					if interactor.InMenu {
						ctx.handleInteractorInMenu(ev, interactor)
					} else {
						ctx.handleInteractorInGame(ev, interactor)
					}
				}
			}

			ev.Done()
		}
	}()
}

// handleInteractorInMenu handles user input during an interaction with an interactive.
func (ctx *interactiveContext) handleInteractorInMenu(ev ecs.EventContainer, interactor eInteractor) {
	ctx.secondaryLabel.Centered = false
	ev.Next <- ChangeHUDPromptEvent{ctx.ePrimaryLabel, interactor.Menu.Prompt}

	choiceList := ""
	for i, choice := range interactor.Menu.Choices {
		choiceList += strconv.Itoa(i+1) + ") " + choice.Label + "\n"

		if ctx.win.JustPressed(pixelgl.Key1+pixelgl.Button(i)) ||
			(len(interactor.Menu.Choices) == 1 && ctx.win.JustPressed(pixelgl.KeySpace)) {

			if choice.Action == nil {
				interactor.Menu = nil
			} else {
				interactor.Menu = choice.Action(ev)
			}

			if interactor.Menu == nil {
				interactor.InMenu = false
			}

			break
		}
	}

	ev.Next <- ChangeHUDPromptEvent{ctx.eSecondaryLabel, choiceList}
}

// handleInteractorInGame handles button prompt HUDs for interactives during gameplay.
func (ctx *interactiveContext) handleInteractorInGame(ev ecs.EventContainer, interactor eInteractor) {
	ctx.secondaryLabel.Centered = true
	niid, nearestInteractive := ctx.findNearestInteractive(interactor)

	if niid == 0 {
		ev.Next <- ChangeHUDPromptEvent{ctx.ePrimaryLabel, ""}
		ev.Next <- ChangeHUDPromptEvent{ctx.eSecondaryLabel, ""}
	} else {
		ev.Next <- TransformEvent{ctx.eSecondaryLabel, nearestInteractive.X, nearestInteractive.Y + 40, true}
		ev.Next <- ChangeHUDPromptEvent{ctx.ePrimaryLabel, nearestInteractive.Name}
		ev.Next <- ChangeHUDPromptEvent{ctx.eSecondaryLabel, nearestInteractive.Prompt}

		if ctx.win.JustPressed(pixelgl.KeySpace) {
			interactor.InMenu = true
			interactor.NearbyInteractive = 0
			interactor.Menu = nearestInteractive.Menu(ev)
		}
	}
}

// findNearestInteractive locates an arbitrary interactive component within interaction range of the given interactor.
// It will return 0 if no such components are found.
func (ctx *interactiveContext) findNearestInteractive(interactor eInteractor) (uint64, eInteractive) {
	// Identify nearest interactive
	for iid, interactive := range ctx.interactives {
		if math.Hypot(interactive.X-interactor.X, interactive.Y-interactor.Y) < 100 {
			return iid, interactive
		}
	}

	return 0, eInteractive{}
}
