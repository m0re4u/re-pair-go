package systems

import (
	"fmt"
	"image/color"
	"math"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"
)

// Spritesheet contains art for units
var Spritesheet *common.Spritesheet

// IdleAnimation for our unit
var IdleAnimation *common.Animation

// Unit unit defition
type Unit struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	common.MouseComponent
	common.AnimationComponent
	position engo.Point
	selected bool
	speed    float32
	shadow   Shadow
	path     *PathPoint
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
	ast        AStar
	p2p        AStarConfig
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

	// Visuals
	Spritesheet = common.NewSpritesheetFromFile("textures/art.png", 8, 8)
	IdleAnimation = &common.Animation{Name: "idle", Frames: []int{7, 8}}

	// Pathing
	us.ast = NewAStar(300, 300)
	us.p2p = NewPointToPoint()

	fmt.Println("UnitSpawner was added to the Scene")

}

// NewUnit create a new unit entity
func (us *UnitSpawner) newUnit(posx float32, posy float32) Unit {
	texture := Spritesheet.Cell(7)
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

	unit.AnimationComponent = common.NewAnimationComponent(Spritesheet.Drawables(), 0.5)
	unit.AnimationComponent.AddDefaultAnimation(IdleAnimation)

	unit.speed = 2

	return unit
}

// moveUnit move the unit a single step in the direction given by transx and transy
func (us *UnitSpawner) moveUnit(unit *Unit, transx float32, transy float32) {
	if transx > 0 {
		unit.SpaceComponent.Position.X += unit.speed
		unit.shadow.SpaceComponent.Position.X += unit.speed
	} else if transx < 0 {
		unit.SpaceComponent.Position.X -= unit.speed
		unit.shadow.SpaceComponent.Position.X -= unit.speed
	} else if transy > 0 {
		unit.SpaceComponent.Position.Y += unit.speed
		unit.shadow.SpaceComponent.Position.Y += unit.speed
	} else if transy < 0 {
		unit.SpaceComponent.Position.Y -= unit.speed
		unit.shadow.SpaceComponent.Position.Y -= unit.speed
	}
	// Else, both translations are 0 and do a noop
}

// SelectUnit select a unit and color shadow
func (us *UnitSpawner) SelectUnit(unit *Unit) {
	unit.selected = true
	unit.shadow.RenderComponent.Color = color.RGBA{0, 255, 0, 255}
}

// DeselectUnit deselect a unit and make shadow black again
func (us *UnitSpawner) DeselectUnit(unit *Unit) {
	unit.selected = false
	unit.shadow.RenderComponent.Color = color.RGBA{0, 0, 0, 255}
}

// MoveUnit move unit to target location
func (us *UnitSpawner) MoveUnit(unit *Unit, target engo.Point) {
	source := []Point{Convert(unit.SpaceComponent.Center())}
	ttarget := []Point{Convert(target)}
	fmt.Println("Finding path from", source, "to", ttarget)
	end := us.ast.FindPath(us.p2p, source, ttarget)
	unit.path = end

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
		case *common.AnimationSystem:
			sys.Add(&unit.BasicEntity, &unit.AnimationComponent, &unit.RenderComponent)
		case *UnitSpawner:
			sys.Add(&unit)
		}
	}

}

// Update is ran every frame, with `dt` being the time
// in seconds since the last frame
func (us *UnitSpawner) Update(dt float32) {
	for _, unit := range us.AliveUnits {
		if unit.path != nil && unit.path.Parent != nil {
			nextTarget := ConvertBack(unit.path.Point)
			fmt.Println("next step:", unit.path.Point, "target", nextTarget)
			transx := float32(nextTarget.X) - unit.SpaceComponent.Center().X
			transy := float32(nextTarget.Y) - unit.SpaceComponent.Center().Y
			us.moveUnit(unit, transx, transy)
			if math.Abs(float64(transx))+math.Abs(float64(transy)) < 4 {
				unit.path = unit.path.Parent
			}
		}
	}
}
