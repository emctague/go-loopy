package systems

import (
	"github.com/emctague/go-loopy/ecs"
	"github.com/faiface/pixel/pixelgl"
	"math"
	"strconv"
)

// menuChoice represents one choice in an interactive menu.
// It has a label to describe what the interactor is selecting / saying, and a function which performs some actions and
// then returns another menu, or nil to exit the menus.
// The Action function itself may also be nil to just exit when selected.
type menuChoice struct {
	Label  string
	Action func(ecs.EventContainer) *InteractionMenu
}

// InteractionMenu represents a menu with a prompt and several selectable options.
type InteractionMenu struct {
	Prompt  string
	Choices []menuChoice
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

// IMenu is Shorthand to construct a menu screen.
// The first argument is the prompt for the menu, followed by pairs of choice labels and their action functions
// (see menuChoice.)
func IMenu(message string, args ...interface{}) func(ecs.EventContainer) *InteractionMenu {
	return func(ev ecs.EventContainer) *InteractionMenu {
		menu := &InteractionMenu{message, []menuChoice{}}

		for i := 0; i < len(args); i += 2 {
			if args[i+1] == nil {
				menu.Choices = append(menu.Choices, menuChoice{args[i].(string), nil})
			} else {
				menu.Choices = append(menu.Choices, menuChoice{args[i].(string), args[i+1].(func(ecs.EventContainer) *InteractionMenu)})
			}
		}

		return menu
	}
}

type eInteractor struct {
	*Transform
	*Interactor
}
type eInteractive struct {
	*Transform
	*Interactive
}

// InteractiveSystem handles interactive in-game menus.
func InteractiveSystem(e *ecs.ECS, win *pixelgl.Window) {

	primaryLabel := &HUDLine{"", true, 2, 0}
	var ePrimaryLabel uint64

	secondaryLabel := &HUDLine{"", true, 1.5, 0}
	tSecondaryLabel := &Transform{0, 0, 0}
	var eSecondaryLabel uint64

	interactors := make(map[uint64]eInteractor)
	interactives := make(map[uint64]eInteractive)

	events := e.Subscribe()

	go func() {
		for ev := range events {
			switch event := ev.Event.(type) {
			case ecs.SetupEvent:
				eSecondaryLabel = e.AddEntity(secondaryLabel, tSecondaryLabel)
				ePrimaryLabel = e.AddEntity(primaryLabel, &Transform{0, 20, eSecondaryLabel})

			case ecs.EntityAddedEvent:
				ecs.UnpackEntity(event, &interactors)
				ecs.UnpackEntity(event, &interactives)

			case ecs.EntityRemovedEvent:
				ecs.RemoveEntity(event.ID, &interactors)
				ecs.RemoveEntity(event.ID, &interactives)

			case ecs.UpdateBeginEvent:

				for _, interactor := range interactors {

					if interactor.InMenu {

						secondaryLabel.Centered = false
						ev.Next <- ChangeHUDPromptEvent{ePrimaryLabel, interactor.Menu.Prompt}

						choiceList := ""
						for i, choice := range interactor.Menu.Choices {
							choiceList += strconv.Itoa(i+1) + ") " + choice.Label + "\n"

							if win.JustPressed(pixelgl.Key1+pixelgl.Button(i)) || len(interactor.Menu.Choices) == 1 && win.JustPressed(pixelgl.KeySpace) {
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

						ev.Next <- ChangeHUDPromptEvent{eSecondaryLabel, choiceList}

					} else {

						secondaryLabel.Centered = true
						niid, nearestInteractive := findNearestInteractive(interactor, &interactives)

						if niid == 0 {
							ev.Next <- ChangeHUDPromptEvent{ePrimaryLabel, ""}
							ev.Next <- ChangeHUDPromptEvent{eSecondaryLabel, ""}
						} else {
							ev.Next <- TransformEvent{eSecondaryLabel, nearestInteractive.X, nearestInteractive.Y + 40, true}
							ev.Next <- ChangeHUDPromptEvent{ePrimaryLabel, nearestInteractive.Name}
							ev.Next <- ChangeHUDPromptEvent{eSecondaryLabel, nearestInteractive.Prompt}

							if win.JustPressed(pixelgl.KeySpace) {
								interactor.InMenu = true
								interactor.NearbyInteractive = 0
								interactor.Menu = nearestInteractive.Menu(ev)
							}
						}

					}
				}
			}

			ev.Done()
		}
	}()
}

func findNearestInteractive(interactor eInteractor, interactives *map[uint64]eInteractive) (uint64, eInteractive) {
	// Identify nearest interactive
	for iid, interactive := range *interactives {
		if math.Hypot(interactive.X-interactor.X, interactive.Y-interactor.Y) < 100 {
			return iid, interactive
		}
	}

	return 0, eInteractive{}
}
