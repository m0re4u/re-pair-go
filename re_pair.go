package main

import (
	"image/color"

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

	us := &systems.UnitSpawner{}
	world.AddSystem(us)
	world.AddSystem(&systems.MouseFollower{})

	us.SpawnUnitAtLocation(200, 200)
	us.SpawnUnitAtLocation(300, 200)
	us.SpawnUnitAtLocation(400, 200)
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
