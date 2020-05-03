package utils

import (
	"github.com/emctague/go-loopy/ecs"
	"github.com/emctague/go-loopy/systems"
)

type dialogScriptPrompt struct {
	Prompt  string
	Choices []string
}

// PromptTool is a function which produces a new dialog prompt with the given message and possible choices.
type PromptTool func(message string, choices ...string)

// MakeDialogScript generates the appropriate dialog handling functions given a function which produces prompts using
// the given PromptTool and then reads choices via the given channel.
// This simplifies the process of writing an interactive menu significantly, to feel more like writing a blocking
// command-line menu.
func MakeDialogScript(handler func(choices chan int, prompt PromptTool, ev **ecs.EventContainer)) func(ev ecs.EventContainer) *systems.InteractionMenu {
	return func(ev ecs.EventContainer) *systems.InteractionMenu {
		choices := make(chan int)
		prompts := make(chan dialogScriptPrompt)

		evPtr := &ev

		var promptHandler func(ev ecs.EventContainer) *systems.InteractionMenu

		promptTool := func(message string, choices ...string) {
			prompts <- dialogScriptPrompt{Prompt: message, Choices: choices}
		}

		go func() {
			handler(choices, promptTool, &evPtr)
			close(prompts)
		}()

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
