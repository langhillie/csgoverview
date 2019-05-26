package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	dem "github.com/markus-wa/demoinfocs-golang"
	common "github.com/markus-wa/demoinfocs-golang/common"
	event "github.com/markus-wa/demoinfocs-golang/events"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	winHeight           int32 = 1024
	winWidth            int32 = 1024
	flashEffectLifetime int   = 10
	heEffectLifetime    int   = 10
	nameMapFontSize     int   = 14
)

var (
	mapName             string
	halfStarts          []int
	roundStarts         []int
	grenadeEffects      map[int][]GrenadeEffect
	curFrame            int
	frameRate           float64
	frameRateRounded    int
	smokeEffectLifetime int
	paused              bool
	states              []OverviewState
	font                *ttf.Font
)

type OverviewState struct {
	IngameTick            int
	Players               []OverviewPlayer
	Grenades              []common.GrenadeProjectile
	Infernos              []common.Inferno
	Bomb                  common.Bomb
	TeamCounterTerrorists common.TeamState
	TeamTerrorists        common.TeamState
}

type GrenadeEffect struct {
	event.GrenadeEvent
	Lifetime int
}

// Do not use Weapons(), but do use Weapons instead
type OverviewPlayer struct {
	common.Player
	Weapons []common.Equipment
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./csgoverview [demoname]")
		return
	}

	err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS)
	if err != nil {
		log.Println(err)
		return
	}
	defer sdl.Quit()

	err = ttf.Init()
	if err != nil {
		log.Println(err)
		return
	}
	defer ttf.Quit()

	font, err = ttf.OpenFont("liberationserif-regular.ttf", nameMapFontSize)
	if err != nil {
		log.Println(err)
		return
	}
	defer font.Close()

	window, err := sdl.CreateWindow("csgoverview", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winHeight, winWidth, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Println(err)
		return
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Println(err)
		return
	}
	defer renderer.Destroy()

	demo, err := os.Open(os.Args[1])
	if err != nil {
		log.Println(err)
		return
	}
	defer demo.Close()

	halfStarts = make([]int, 0)
	roundStarts = make([]int, 0)
	roundStarts = append(roundStarts, 0)
	grenadeEffects = make(map[int][]GrenadeEffect)

	parser := dem.NewParser(demo)

	header, err := parser.ParseHeader()
	if err != nil {
		log.Println(err)
		return
	}

	frameRate = header.FrameRate()
	frameRateRounded = int(math.Round(frameRate))
	mapName = header.MapName
	smokeEffectLifetime = int(18 * frameRate)

	registerEventHandlers(parser)

	err = parser.ParseToEnd()
	if err != nil {
		log.Println(err)
		return
	}

	mapSurface, err := img.Load(fmt.Sprintf("%v.jpg", mapName))
	if err != nil {
		log.Println(err)
		return
	}
	defer mapSurface.Free()

	mapTexture, err := renderer.CreateTextureFromSurface(mapSurface)
	if err != nil {
		log.Println(err)
		return
	}
	defer mapTexture.Destroy()

	err = renderer.Clear()
	if err != nil {
		log.Println(err)
		return
	}
	err = renderer.Copy(mapTexture, nil, nil)
	if err != nil {
		log.Println(err)
		return
	}
	renderer.Present()

	// Second pass of the parser
	_, err = demo.Seek(0, 0)
	if err != nil {
		log.Println(err)
		return
	}

	parser = dem.NewParser(demo)
	states = parseGameStates(parser)

	// MAIN GAME LOOP
	for {
		frameStart := time.Now()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch eventT := event.(type) {
			case *sdl.QuitEvent:
				return

			case *sdl.KeyboardEvent:
				handleKeyboardEvents(eventT)
			}
		}

		if paused {
			sdl.Delay(32)
			// draw?
			continue
		}

		renderer.Clear()
		renderer.Copy(mapTexture, nil, nil)

		infernos := states[curFrame].Infernos
		for _, inferno := range infernos {
			DrawInferno(renderer, &inferno, mapName)
		}

		players := states[curFrame].Players
		for _, player := range players {
			DrawPlayer(renderer, &player, mapName)
		}

		effects := grenadeEffects[curFrame]
		for _, effect := range effects {
			DrawGrenadeEffect(renderer, &effect, mapName)
		}

		grenades := states[curFrame].Grenades
		for _, grenade := range grenades {
			DrawGrenade(renderer, &grenade, mapName)
		}

		bomb := states[curFrame].Bomb
		DrawBomb(renderer, &bomb, mapName)

		renderer.Present()

		updateWindowTitle(window)

		var playbackSpeed float64 = 1

		// frameDuration is in ms
		frameDuration := float64(time.Since(frameStart) / 1000000)
		keyboardState := sdl.GetKeyboardState()
		if keyboardState[sdl.GetScancodeFromKey(sdl.K_w)] != 0 {
			playbackSpeed = 5
		}
		if keyboardState[sdl.GetScancodeFromKey(sdl.K_s)] != 0 {
			playbackSpeed = 0.5
		}
		delay := (1/playbackSpeed)*frameRate - frameDuration
		if delay < 0 {
			delay = 0
		}
		sdl.Delay(uint32(delay))
		if curFrame < len(states)-1 {
			curFrame++
		}
	}

}

func grenadeEventHandler(lifetime int, frame int, e event.GrenadeEvent) {
	for i := 0; i < lifetime; i++ {
		effect := GrenadeEffect{
			GrenadeEvent: e,
			Lifetime:     i,
		}
		effects, ok := grenadeEffects[frame+i]
		if ok {
			grenadeEffects[frame+i] = append(effects, effect)
		} else {
			grenadeEffects[frame+i] = []GrenadeEffect{effect}
		}
	}
}

func registerEventHandlers(parser *dem.Parser) {
	h1 := parser.RegisterEventHandler(func(event.RoundStart) {
		roundStarts = append(roundStarts, parser.CurrentFrame())
	})
	h2 := parser.RegisterEventHandler(func(event.MatchStart) {
		halfStarts = append(halfStarts, parser.CurrentFrame())
	})
	h3 := parser.RegisterEventHandler(func(event.GameHalfEnded) {
		halfStarts = append(halfStarts, parser.CurrentFrame())
	})
	h4 := parser.RegisterEventHandler(func(event.TeamSideSwitch) {
		halfStarts = append(halfStarts, parser.CurrentFrame())
	})
	parser.RegisterEventHandler(func(e event.FlashExplode) {
		frame := parser.CurrentFrame()
		grenadeEventHandler(flashEffectLifetime, frame, e.GrenadeEvent)
	})
	parser.RegisterEventHandler(func(e event.HeExplode) {
		frame := parser.CurrentFrame()
		grenadeEventHandler(heEffectLifetime, frame, e.GrenadeEvent)
	})
	parser.RegisterEventHandler(func(e event.SmokeStart) {
		frame := parser.CurrentFrame()
		grenadeEventHandler(smokeEffectLifetime, frame, e.GrenadeEvent)
	})
	parser.RegisterEventHandler(func(event.AnnouncementWinPanelMatch) {
		parser.UnregisterEventHandler(h1)
		parser.UnregisterEventHandler(h2)
		parser.UnregisterEventHandler(h3)
		parser.UnregisterEventHandler(h4)
	})
}

// parse demo and save GameStates in slice
func parseGameStates(parser *dem.Parser) []OverviewState {
	states := make([]OverviewState, 0)

	for ok, err := parser.ParseNextFrame(); ok; ok, err = parser.ParseNextFrame() {
		if err != nil {
			log.Println(err)
			// return here or not?
			continue
		}

		gameState := parser.GameState()

		players := make([]OverviewPlayer, 0)

		for _, p := range gameState.Participants().Playing() {
			// common.RawWeapons is map[int]*common.Equipment, but it is unclear what the key means
			equipment := make([]common.Equipment, 0)
			for _, eq := range p.Weapons() {
				equipment = append(equipment, *eq)
			}
			player := OverviewPlayer{
				Player:  *p,
				Weapons: equipment,
			}
			players = append(players, player)
		}

		grenades := make([]common.GrenadeProjectile, 0)

		for _, grenade := range gameState.GrenadeProjectiles() {
			grenades = append(grenades, *grenade)
		}

		infernos := make([]common.Inferno, 0)

		for _, inferno := range gameState.Infernos() {
			infernos = append(infernos, *inferno)
		}

		bomb := *gameState.Bomb()

		cts := *gameState.TeamCounterTerrorists()
		ts := *gameState.TeamTerrorists()

		state := OverviewState{
			IngameTick:            parser.GameState().IngameTick(),
			Players:               players,
			Grenades:              grenades,
			Infernos:              infernos,
			Bomb:                  bomb,
			TeamCounterTerrorists: cts,
			TeamTerrorists:        ts,
		}

		states = append(states, state)
	}

	return states
}

func handleKeyboardEvents(eventT *sdl.KeyboardEvent) {
	if eventT.Type == sdl.KEYDOWN && eventT.Keysym.Sym == sdl.K_SPACE {
		paused = !paused
	}

	if eventT.Type == sdl.KEYDOWN && eventT.Keysym.Sym == sdl.K_a {
		if eventT.Keysym.Mod == sdl.KMOD_LSHIFT || eventT.Keysym.Mod == sdl.KMOD_RSHIFT {
			if curFrame < frameRateRounded*30 {
				curFrame = 0
			} else {
				curFrame -= frameRateRounded * 30
			}
		} else {
			if curFrame < frameRateRounded*10 {
				curFrame = 0
			} else {
				curFrame -= frameRateRounded * 10
			}
		}
	}

	if eventT.Type == sdl.KEYDOWN && eventT.Keysym.Sym == sdl.K_d {
		if eventT.Keysym.Mod == sdl.KMOD_LSHIFT || eventT.Keysym.Mod == sdl.KMOD_RSHIFT {
			if curFrame+frameRateRounded*30 > len(states)-1 {
				curFrame = len(states) - 1
			} else {
				curFrame += frameRateRounded * 30
			}
		} else {
			if curFrame+frameRateRounded*10 > len(states)-1 {
				curFrame = len(states) - 1
			} else {
				curFrame += frameRateRounded * 10
			}
		}
	}

	if eventT.Type == sdl.KEYDOWN && eventT.Keysym.Sym == sdl.K_q {
		if eventT.Keysym.Mod == sdl.KMOD_LSHIFT || eventT.Keysym.Mod == sdl.KMOD_RSHIFT {
			set := false
			for i, frame := range halfStarts {
				if curFrame < frame {
					if i > 1 && curFrame < halfStarts[i-1]+frameRateRounded/2 {
						curFrame = halfStarts[i-2]
						set = true
						break
					}
					curFrame = halfStarts[i-1]
					set = true
					break
				}
			}
			// not set -> last round of match
			if !set {
				if len(halfStarts) > 1 && curFrame < halfStarts[len(halfStarts)-1]+frameRateRounded/2 {
					curFrame = halfStarts[len(halfStarts)-2]
				} else {
					curFrame = halfStarts[len(halfStarts)-1]
				}
			}
		} else {
			set := false
			for i, frame := range roundStarts {
				if curFrame < frame {
					if i > 1 && curFrame < roundStarts[i-1]+frameRateRounded/2 {
						curFrame = roundStarts[i-2]
						set = true
						break
					}
					curFrame = roundStarts[i-1]
					set = true
					break
				}
			}
			// not set -> last round of match
			if !set {
				if len(roundStarts) > 1 && curFrame < roundStarts[len(roundStarts)-1]+frameRateRounded/2 {
					curFrame = roundStarts[len(roundStarts)-2]
				} else {
					curFrame = roundStarts[len(roundStarts)-1]
				}
			}
		}
	}

	if eventT.Type == sdl.KEYDOWN && eventT.Keysym.Sym == sdl.K_e {
		if eventT.Keysym.Mod == sdl.KMOD_LSHIFT || eventT.Keysym.Mod == sdl.KMOD_RSHIFT {
			for _, frame := range halfStarts {
				if curFrame < frame {
					curFrame = frame
					break
				}
			}
		} else {
			for _, frame := range roundStarts {
				if curFrame < frame {
					curFrame = frame
					break
				}
			}
		}
	}
}

func updateWindowTitle(window *sdl.Window) {
	cts := states[curFrame].TeamCounterTerrorists
	ts := states[curFrame].TeamTerrorists
	clanNameCTs := cts.ClanName
	if clanNameCTs == "" {
		clanNameCTs = "Counter Terrorists"
	}
	clanNameTs := ts.ClanName
	if clanNameTs == "" {
		clanNameTs = "Terrorists"
	}
	windowTitle := fmt.Sprintf("%s  [%d:%d]  %s", clanNameCTs, cts.Score, ts.Score, clanNameTs)
	// expensive?
	window.SetTitle(windowTitle)
}
