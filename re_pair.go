package main

import (
	"image/color"
	"log"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"

	"re-pair-go/systems"
)

type myScene struct{}

// Type uniquely defines your game type
func (*myScene) Type() string { return "myGame" }

// Preload is called before loading any assets from the disk,
// to allow you to register / queue them
func (*myScene) Preload() {
	engo.Files.Load("textures/unit.png")
	engo.Files.Load("textures/cursor.png")
}

// Setup is called before the main loop starts. It allows you to add entities
// and systems to your Scene.
func (*myScene) Setup(u engo.Updater) {
	world, _ := u.(*ecs.World)
	engo.Input.RegisterButton("SpawnUnit", engo.KeySpace)
	common.SetBackground(color.White)

	world.AddSystem(&common.RenderSystem{})
	world.AddSystem(&common.MouseSystem{})

	world.AddSystem(&systems.UnitSpawner{})
	world.AddSystem(&systems.MouseFollower{})

	engo.SetCursor(engo.CursorCrosshair)
	// Create an entity
	guy := systems.Player{BasicEntity: ecs.NewBasic()}

	texture, err := common.LoadedSprite("textures/cursor.png")
	if err != nil {
		log.Println(err)
	}
	// Initialize the components, set scale to 8x
	guy.RenderComponent = common.RenderComponent{
		Drawable: texture,
		Scale:    engo.Point{X: 4, Y: 4},
	}
	guy.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{X: 0, Y: 0},
		Width:    texture.Width() * guy.RenderComponent.Scale.X,
		Height:   texture.Height() * guy.RenderComponent.Scale.Y,
	}

	// Add it to appropriate systems
	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&guy.BasicEntity, &guy.RenderComponent, &guy.SpaceComponent)
		case *systems.MouseFollower:
			sys.Add(&guy.BasicEntity, &guy.RenderComponent, &guy.SpaceComponent)
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
	engo.Run(opts, &myScene{})
}
