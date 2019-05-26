A 2D demo replay tool for Counter Strike: Global Offensive.

# Tools

* [golang](https://golang.org/)
* [SDL2](https://wiki.libsdl.org/Introduction)
* [go-sdl2](https://github.com/veandco/go-sdl2)
* [demoinfocs-golang](https://github.com/markus-wa/demoinfocs-golang)
* [csgo-overviews](https://github.com/zoidbergwill/csgo-overviews)

# Roadmap

## Milestone v0.1.0

### Demo playback

* [X] Map
* [X] Playerpositions
* [X] Playernames
* [X] Player line of vision
* [ ] ~~Shots~~
* [X] Grenades during flight
* [ ] ~~NadeTails~~
* [X] Effects for flashbangs and hes
* [X] Effects for smokes
* [X] Effects for mollys
* [X] Fade-out effect for smokes
* [X] Timer for smokes
* [X] `x` at places where players died
* [X] Indicator for flash-effect / -duration
* [X] Indicator for defusing player
* [X] C4 on the ground
* [X] C4 on player

### Keybinds

* [X] a -> 10 s backwards
* [X] d -> 10 s forwards
* [X] A -> 30 s backwards
* [X] D -> 30 s forwards
* [X] w -> hold to speed up 5 x
* [X] s -> hold to slow down to 0.5 x
* [X] q -> round backwards
* [X] e -> round forwards
* [X] Q -> to start of previous half
* [X] E -> to start of next half
* [X] space -> toggle pause
* [ ] [num]g -> go to round [num]
* [ ] p -> take screenshot
* [ ] i -> export gif/mp4

### Misc

* [ ] jpg of current frame
* [ ] Gif / mp4 export of a single round
* [X] unicode support for playernames (ttf)
* [ ] Smoke radius scaling (map metadata)

## Milestone v0.2.0

### Additional information about round and players

* [ ] #round / #total
* [X] Score
* [X] Teamnames
* [ ] Warmup-indicator
* [ ] Freezetime timer
* [ ] Time remaining in current round
* [ ] Bombplant Indicator
* [ ] Bomb timer
* [ ] Defuse timer
* [ ] Killfeed
* [ ] Player details (left/right)
    - Name
    - Hp
    - Armor
    - Helmet
    - Primary
    - Secondary
    - Grenades
    - Defkit
    - (Taser)
    - Money
    - Kills in current round
* [ ] Results of previous rounds (survivors?)

## Milestone v0.3.0

### Misc

* [ ] Scaling
* [ ] Multi-platform
 
### Configurable keybinds and scaling

* hardcoded defaults

### Command Line Interface

* take Screenshot
* export Gif

# Non-Features

### GUI / Buttons

* Keybinds

### Analysis

* CSGO Demos Manager
* other 3rd-party tools

### Networking?

* join.me
* screenshare
* stream
* gif export
* cli

### Drawing?

* twiddla
* paint
* live overlay ?
