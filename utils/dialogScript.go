package utils

import (
	"github.com/emctague/go-loopy/ecs"
	"github.com/emctague/go-loopy/systems"
)

// dialogScriptPrompt represents an on-screen prompt and its possible choices
type dialogScriptPrompt struct {
	Prompt  string
	Choices []string
}

// PromptTool is a function which produces a new dialog prompt with the given message and possible choices.
type PromptTool func(message string, choices ...string) int

// dialogScript is any function that handles back-and-forth dialog.
// The provided 'prompt' method should be used to prompt for responses.
type dialogScript func(prompt PromptTool, ev **ecs.EventContainer)

// MakeDialogScript generates the appropriate dialog handling functions given a dialogScript (handler function.)
// This simplifies the process of writing an interactive menu significantly, to feel more like writing a blocking
// command-line menu.
func MakeDialogScript(script dialogScript) func(ev ecs.EventContainer) *systems.InteractionMenu {
	return func(ev ecs.EventContainer) *systems.InteractionMenu {

		// Create communication channels
		choices := make(chan int)
		prompts := make(chan dialogScriptPrompt)

		// Store the event container as a pointer, for access from the script
		evPtr := &ev

		// Declare a prompt tool that serves as a shorthand to set a new prompt
		promptTool := func(message string, allChoices ...string) int {
			prompts <- dialogScriptPrompt{Prompt: message, Choices: allChoices}
			return <-choices
		}

		// Run the dialog script in parallel
		go func() {
			script(promptTool, &evPtr)
			close(prompts)
		}()

		var promptHandler func(ev ecs.EventContainer) *systems.InteractionMenu
		promptHandler = func(ev ecs.EventContainer) *systems.InteractionMenu {
			evPtr = &ev

			prompt, ok := <-prompts
			if !ok {
				close(choices)
				return nil
			}

			var choiceFuncs []systems.MenuChoice

			for i, choice := range prompt.Choices {
				thisI := i
				choiceFuncs = append(choiceFuncs, systems.MenuChoice{
					Label: choice,
					Action: func(container ecs.EventContainer) *systems.InteractionMenu {
						choices <- thisI
						return promptHandler(container)
					},
				})
			}

			return &systems.InteractionMenu{
				Prompt:  prompt.Prompt,
				Choices: choiceFuncs,
			}
		}

		return promptHandler(ev)
	}
}
