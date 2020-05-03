# go-loopy

This is a strange, overly-complicated experiment!

`go-loopy` is essentially a game engine and simple game with an ECS architecture.

Here's a rundown:

 1. Entities don't really exist as anything more than a Unique ID shared by associated components.
 
 2. Components are *not* stored in contiguous memory - they are allocated individually and passed around as pointers.
    This is not ideal and was done for convenience.
 
 3. `go-loopy` is *heavily* event based. Systems in this ECS are goroutines that self-register as event handlers with
    the ECS, and then constantly listen for events.
    
    1. Systems can post new events, which will be processed either:
       1. After all currently posted events the current frame (`PublishThisFrame`)
       2. On the next Frame (`PublishNextFrame`)
       4. Immediately after the current event is finished being processed. (`ev.Next <- ...`)
       
    2. All systems process the same event in parallel, and will wait for other systems to finish before all systems move
       onto the next event. This is not ideal, but again, it is done for convenience and to ensure that events can be
       separated across "updates"/"frames"/"ticks".
       
       The `Update` events are responsible for most major processing, however, and therefore the bulk of heavy work
       occurs in parallel.
 
    3. Updates, entity creation, entity deletion, etc. are all events that systems can handle as they please. Most
       systems will keep track of one or more maps which map unique entity IDs to collections of relevant, associated
       components.
 
    4. Behavior which significantly changes component values that might also be used or changed in parallel by other
       systems should usually be relegated to its own event - for example, manual changes to position and velocity
       should occur in their own events in order to avoid conflicting with the transform and physics system update
       events.

 4. Systems may possess singleton / dedicated entities to serve some particular purpose. For example, the interactives
    system (which provides a simple player-to-NPC scripted dialog system) owns two entities which store text to be drawn
    by the renderer system.

 5. This program abuses reflection quite a bit for convenience without much consideration for performance.
 
 6. The main method is very, very ugly.