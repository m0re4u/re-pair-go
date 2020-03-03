package systems

import (
	"fmt"
	"image/color"
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
	position engo.Point
	selected bool
	shadow   Shadow
}

// Shadow render unit shadow
type Shadow struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
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
func (us *UnitSpawner) newUnit(posx float32, posy float32) Unit {
	texture, err := common.LoadedSprite("textures/unit.png")
	if err != nil {
		log.Println("Unable to load texture: " + err.Error())
	}
	unit := Unit{BasicEntity: ecs.NewBasic()}
	unit.position = engo.Point{X: posx, Y: posy}
	unit.RenderComponent = common.RenderComponent{
		Drawable: texture,
		Scale:    engo.Point{X: 8, Y: 8},
	}
	unit.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{X: unit.position.X, Y: unit.position.Y},
		Width:    texture.Width() * unit.RenderComponent.Scale.X,
		Height:   texture.Height() * unit.RenderComponent.Scale.Y,
	}

	unit.shadow = Shadow{BasicEntity: ecs.NewBasic()}
	unit.shadow.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{X: unit.position.X, Y: unit.position.Y},
		Width:    texture.Width() * unit.RenderComponent.Scale.X,
		Height:   texture.Height() * unit.RenderComponent.Scale.Y,
	}
	unit.shadow.RenderComponent = common.RenderComponent{Drawable: common.Circle{}, Color: color.RGBA{0, 0, 0, 255}}

	return unit
}

func (us *UnitSpawner) moveUnit(unit *Unit, transx float32, transy float32) {
	unit.SpaceComponent.Position.X += transx
	unit.SpaceComponent.Position.Y += transy
	unit.shadow.SpaceComponent.Position.X += transx
	unit.shadow.SpaceComponent.Position.Y += transy
}

// SpawnUnitAtLocation spawn new unit at the given location
func (us *UnitSpawner) SpawnUnitAtLocation(x float32, y float32) {
	unit := us.newUnit(x, y)
	for _, system := range us.world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&unit.BasicEntity, &unit.RenderComponent, &unit.SpaceComponent)
			sys.Add(&unit.shadow.BasicEntity, &unit.shadow.RenderComponent, &unit.shadow.SpaceComponent)
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
		if u.MouseComponent.Clicked {
			us.moveUnit(u, 10, 0)
		}
	}
}
