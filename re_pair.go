package main

import (
	"image/color"
	"log"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"

	"re-pair-go/systems"
)

// DefaultScene the default game scene
type DefaultScene struct{}

// Type uniquely defines your game type
func (*DefaultScene) Type() string { return "Re-Pair" }

// Preload is called before loading any assets from the disk,
// to allow you to register / queue them
func (*DefaultScene) Preload() {
	engo.Files.Load("textures/unit.png")
	engo.Files.Load("textures/cursor.png")
}

// Setup is called before the main loop starts. It allows you to add entities
// and systems to your Scene.
func (*DefaultScene) Setup(u engo.Updater) {
	world, _ := u.(*ecs.World)

	// Input settings
	engo.Input.RegisterButton("SpawnUnit", engo.KeySpace)
	engo.SetCursor(engo.CursorCrosshair)

	common.SetBackground(color.White)

	world.AddSystem(&common.RenderSystem{})
	world.AddSystem(&common.MouseSystem{})

	world.AddSystem(&systems.UnitSpawner{})
	world.AddSystem(&systems.MouseFollower{})

	// Create an player entity
	player := systems.Player{BasicEntity: ecs.NewBasic()}

	texture, err := common.LoadedSprite("textures/cursor.png")
	if err != nil {
		log.Println(err)
	}
	// Initialize the components, set scale to 8x
	player.RenderComponent = common.RenderComponent{
		Drawable: texture,
		Scale:    engo.Point{X: 4, Y: 4},
	}
	player.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{X: 0, Y: 0},
		Width:    texture.Width() * player.RenderComponent.Scale.X,
		Height:   texture.Height() * player.RenderComponent.Scale.Y,
	}

	// Add it to appropriate systems
	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&player.BasicEntity, &player.RenderComponent, &player.SpaceComponent)
		case *systems.MouseFollower:
			sys.Add(&player.BasicEntity, &player.RenderComponent, &player.SpaceComponent)
		}
	}
}

func main() {
	opts := engo.RunOptions{
		Title:          "Re-Pair Game",
		Width:          960,
		Height:         1060,
		StandardInputs: true,
	}
	engo.Run(opts, &DefaultScene{})
}
