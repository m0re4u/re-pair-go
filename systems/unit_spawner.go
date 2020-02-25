package systems

import (
	"fmt"
	"log"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"
)

const unitScaling = 5

// MouseTracker info about mouse
type MouseTracker struct {
	ecs.BasicEntity
	common.MouseComponent
}

// Unit unit defition
type Unit struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

// UnitSpawner takes care of unit spawning
type UnitSpawner struct {
	world *ecs.World

	mouseTracker MouseTracker
}

// Remove is called whenever an Entity is removed from the scene, and thus from this system
func (*UnitSpawner) Remove(ecs.BasicEntity) {}

// New is the initialisation of the System
func (us *UnitSpawner) New(w *ecs.World) {
	us.world = w
	fmt.Println("UnitSpawner was added to the Scene")

	us.mouseTracker.BasicEntity = ecs.NewBasic()
	us.mouseTracker.MouseComponent = common.MouseComponent{Track: true}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.MouseSystem:
			sys.Add(&us.mouseTracker.BasicEntity, &us.mouseTracker.MouseComponent, nil, nil)
		}
	}
}

// Update is ran every frame, with `dt` being the time
// in seconds since the last frame
func (us *UnitSpawner) Update(dt float32) {
	if engo.Input.Button("SpawnUnit").JustPressed() {
		fmt.Println("Player spawned unit")

		unit := Unit{BasicEntity: ecs.NewBasic()}

		unit.SpaceComponent = common.SpaceComponent{
			Position: engo.Point{X: us.mouseTracker.MouseComponent.MouseX - 20, Y: us.mouseTracker.MouseComponent.MouseY - 20},
			Width:    8,
			Height:   8,
		}

		texture, err := common.LoadedSprite("textures/unit.png")
		if err != nil {
			log.Println("Unable to load texture: " + err.Error())
		}

		unit.RenderComponent = common.RenderComponent{
			Drawable: texture,
			Scale:    engo.Point{X: unitScaling, Y: unitScaling},
		}

		for _, system := range us.world.Systems() {
			switch sys := system.(type) {
			case *common.RenderSystem:
				sys.Add(&unit.BasicEntity, &unit.RenderComponent, &unit.SpaceComponent)
			}
		}
	}
}