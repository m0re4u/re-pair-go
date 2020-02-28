package systems

import (
	"fmt"
	"log"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"
)

// Unit unit defition
type Unit struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	common.MouseComponent
}

// UnitSpawner takes care of unit spawning
type UnitSpawner struct {
	world      *ecs.World
	AliveUnits []*Unit // slice of pointers to all units
}

// Remove is called whenever an Entity is removed from the scene, and thus from this system
func (*UnitSpawner) Remove(ecs.BasicEntity) {}

// Add a unit to the system
func (us *UnitSpawner) Add(u *Unit) {
	us.AliveUnits = append(us.AliveUnits, u)
}

// New is the initialisation of the UnitSpawner System
func (us *UnitSpawner) New(w *ecs.World) {
	us.world = w
	fmt.Println("UnitSpawner was added to the Scene")
}

// NewUnit create a new unit entity
func (us *UnitSpawner) NewUnit(posx float32, posy float32) Unit {
	texture, err := common.LoadedSprite("textures/unit.png")
	if err != nil {
		log.Println("Unable to load texture: " + err.Error())
	}
	unit := Unit{BasicEntity: ecs.NewBasic()}
	unit.RenderComponent = common.RenderComponent{
		Drawable: texture,
		Scale:    engo.Point{X: 8, Y: 8},
	}
	unit.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{X: posx, Y: posy},
		Width:    texture.Width() * unit.RenderComponent.Scale.X,
		Height:   texture.Height() * unit.RenderComponent.Scale.Y,
		Rotation: 0,
	}
	return unit

}

// SpawnUnitAtLocation spawn new unit at the given location
func (us *UnitSpawner) SpawnUnitAtLocation(x float32, y float32) {
	unit := us.NewUnit(x, y)
	for _, system := range us.world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&unit.BasicEntity, &unit.RenderComponent, &unit.SpaceComponent)
		case *common.MouseSystem:
			sys.Add(&unit.BasicEntity, &unit.MouseComponent, &unit.SpaceComponent, &unit.RenderComponent)
		case *UnitSpawner:
			sys.Add(&unit)
		}
	}
}

// Update is ran every frame, with `dt` being the time
// in seconds since the last frame
func (us *UnitSpawner) Update(dt float32) {
	for _, u := range us.AliveUnits {
		fmt.Println("")
		if u.MouseComponent.Hovered {
			fmt.Println("HOVER on unit:", u.BasicEntity.ID())
		}
	}
}
