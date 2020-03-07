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

// Unit interface which defines what a unit can do
type Unit interface {
	// exported
	Deselect()
	Select()
	Move(AStar, AStarConfig, engo.Point)
	Register(*UnitSpawner)
	// internal
	step(float32, float32)
}

// BasicUnit Common unit fields
type BasicUnit struct {
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

// Fish First specific unit type
type Fish struct {
	*BasicUnit
}

// Blob Second unit type
type Blob struct {
	*BasicUnit
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
	AliveUnits []*BasicUnit // slice of pointers to all units
	ast        AStar
	p2p        AStarConfig
}

// Remove is called whenever an Entity is removed from the scene, and thus from this system
func (*UnitSpawner) Remove(ecs.BasicEntity) {}

// Add a unit to the system
func (us *UnitSpawner) Add(u *BasicUnit) {
	us.AliveUnits = append(us.AliveUnits, u)
}

// New is the initialisation of the UnitSpawner System
func (us *UnitSpawner) New(w *ecs.World) {
	us.world = w

	// Visuals
	Spritesheet = common.NewSpritesheetFromFile("textures/art.png", 8, 8)

	// Pathing
	us.ast = NewAStar(300, 300) // algo
	us.p2p = NewPointToPoint()  // config

	fmt.Println("UnitSpawner was added to the Scene")

}

// setUnitParameters assign the (texture, animation, speed) parameters to the provided unit
func (us *UnitSpawner) setUnitParameters(unit *BasicUnit, texture common.Drawable, anim *common.Animation, speed float32) {
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
	unit.AnimationComponent.AddDefaultAnimation(anim)
	unit.speed = speed
}

// Create unit object based on unit type
func (us *UnitSpawner) giveUnitParameters(unit *BasicUnit, unitID int) Unit {
	var texture common.Drawable
	var idle *common.Animation
	var speed float32
	if unitID == 0 {
		texture = Spritesheet.Cell(7)
		idle = &common.Animation{Name: "idle", Frames: []int{7, 8}}
		speed = 4
		us.setUnitParameters(unit, texture, idle, speed)
		return &Fish{unit}

	} else if unitID == 1 {
		texture = Spritesheet.Cell(5)
		idle = &common.Animation{Name: "idle", Frames: []int{5, 6}}
		speed = 2
		us.setUnitParameters(unit, texture, idle, speed)
		return &Blob{unit}
	} else {
		return nil
	}
}

// NewUnit create a new unit entity
func (us *UnitSpawner) newUnit(posx float32, posy float32, unitID int) Unit {
	// Create empty unit entity
	unit := BasicUnit{BasicEntity: ecs.NewBasic()}
	unit.position = engo.Point{X: posx, Y: posy}
	// Assign the correct unit parameters according to requested ID
	u := us.giveUnitParameters(&unit, unitID)
	return u
}

// stepUnit move the unit a single step in the direction given by transx and transy
func (unit *BasicUnit) step(transx float32, transy float32) {
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

// Select select a unit and color shadow
func (unit *BasicUnit) Select() {
	unit.selected = true
	unit.shadow.RenderComponent.Color = color.RGBA{0, 255, 0, 255}
}

// Deselect deselect a unit and make shadow black again
func (unit *BasicUnit) Deselect() {
	unit.selected = false
	unit.shadow.RenderComponent.Color = color.RGBA{0, 0, 0, 255}
}

// Move move unit to target location
func (unit *BasicUnit) Move(ast AStar, cfg AStarConfig, target engo.Point) {
	source := []Point{EngoToPathing(unit.SpaceComponent.Center())}
	ttarget := []Point{EngoToPathing(target)}
	end := ast.FindPath(cfg, source, ttarget)
	unit.path = end
}

// Register the unit to the spawner
func (unit *BasicUnit) Register(us *UnitSpawner) {
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
			sys.Add(unit)
		}
	}
}

// SpawnUnitAtLocation spawn new unit at the given location
func (us *UnitSpawner) SpawnUnitAtLocation(x float32, y float32, unitID int) {
	unit := us.newUnit(x, y, unitID)
	unit.Register(us)
}

// Update is ran every frame, with `dt` being the time
// in seconds since the last frame
func (us *UnitSpawner) Update(dt float32) {
	for _, unit := range us.AliveUnits {
		if unit.path != nil && unit.path.Parent != nil {
			nextTarget := PathingToEngo(unit.path.Point)
			transx := float32(nextTarget.X) - unit.SpaceComponent.Center().X
			transy := float32(nextTarget.Y) - unit.SpaceComponent.Center().Y
			unit.step(transx, transy)
			if math.Abs(float64(transx))+math.Abs(float64(transy)) < 4 {
				unit.path = unit.path.Parent
			}
		}
	}
}
