package systems

import (
	"image/color"
	"log"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"
)

var firstDragged bool = true

// MouseCursor entity for drawing the cursor
type MouseCursor struct {
	base      ecs.BasicEntity
	render    common.RenderComponent
	space     common.SpaceComponent
	mouse     common.MouseComponent
	selection Box
}

// MouseFollower system that controls the cursor
type MouseFollower struct {
	world *ecs.World

	cursor MouseCursor
}

// Box for selection
type Box struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

// New create system that follows the mouse
func (s *MouseFollower) New(w *ecs.World) {
	s.world = w

	texture, err := common.LoadedSprite("textures/cursor.png")
	if err != nil {
		log.Println(err)
	}
	s.cursor = MouseCursor{
		base: ecs.NewBasic(),
		render: common.RenderComponent{
			Drawable: texture,
			Scale:    engo.Point{X: 4, Y: 4},
		},
		space: common.SpaceComponent{
			Position: engo.Point{X: 0, Y: 0},
			Width:    texture.Width() * s.cursor.render.Scale.X,
			Height:   texture.Height() * s.cursor.render.Scale.Y,
		},
		mouse: common.MouseComponent{Track: true},
	}

	s.cursor.selection = Box{BasicEntity: ecs.NewBasic()}
	s.cursor.selection.SpaceComponent = common.SpaceComponent{
		Width:    0,
		Height:   0,
		Position: engo.Point{X: 0, Y: 0},
	}
	s.cursor.selection.RenderComponent = common.RenderComponent{Drawable: common.Rectangle{}, Color: color.RGBA{0, 0, 100, 50}}
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&s.cursor.base, &s.cursor.render, &s.cursor.space)
			sys.Add(&s.cursor.selection.BasicEntity, &s.cursor.selection.RenderComponent, &s.cursor.selection.SpaceComponent)
		case *common.MouseSystem:
			sys.Add(&s.cursor.base, &s.cursor.mouse, &s.cursor.space, &s.cursor.render)

		}
	}

}

// Remove follower system from the mouse
func (*MouseFollower) Remove(basic ecs.BasicEntity) {}

// inBox check if a Point is in a Box
func (s *MouseFollower) inBox(box *Box, point engo.Point) bool {
	var left, right, top, bottom float32
	if box.SpaceComponent.Width >= 0 {
		left = box.SpaceComponent.Position.X
		right = box.SpaceComponent.Position.X + box.SpaceComponent.Width
	} else {
		left = box.SpaceComponent.Position.X + box.SpaceComponent.Width
		right = box.SpaceComponent.Position.X
	}

	if box.SpaceComponent.Height >= 0 {
		top = box.SpaceComponent.Position.Y
		bottom = box.SpaceComponent.Position.Y + box.SpaceComponent.Height
	} else {
		top = box.SpaceComponent.Position.Y + box.SpaceComponent.Height
		bottom = box.SpaceComponent.Position.Y
	}
	// drag to the right
	if point.X >= left && point.X < right && point.Y > top && point.Y < bottom {
		return true
	}
	return false
}

func (s *MouseFollower) boxSelect(box *Box) {
	for _, system := range s.world.Systems() {
		switch sys := system.(type) {
		case *UnitSpawner:
			for _, unit := range sys.AliveUnits {
				// Check if unit center in box
				if s.inBox(box, unit.SpaceComponent.Center()) {
					unit.Select()
				} else {
					unit.Deselect()

				}
			}
		}
	}

}

// Update mouse follower's position
func (s *MouseFollower) Update(dt float32) {
	// Place cursor sprite at current mouse position
	s.cursor.space.Position.X = engo.Input.Mouse.X
	s.cursor.space.Position.Y = engo.Input.Mouse.Y

	// Handle mouse clicks and drags
	if s.cursor.mouse.Clicked {
		// On left click, if there is no entity, clear selection
		for _, system := range s.world.Systems() {
			switch sys := system.(type) {
			case *UnitSpawner:
				for _, unit := range sys.AliveUnits {
					if unit.MouseComponent.Hovered {
						unit.Select()
					} else {
						unit.Deselect()
					}
				}
			}
		}
	} else if s.cursor.mouse.RightClicked {
		for _, system := range s.world.Systems() {
			switch sys := system.(type) {
			case *UnitSpawner:
				for _, unit := range sys.AliveUnits {
					if unit.selected {
						unit.Move(sys.ast, sys.p2p, s.cursor.space.Position)
					}
				}
			}
		}

	} else if s.cursor.mouse.Dragged {
		// On drag, select all under the box area
		if firstDragged {
			// Initial drag point, origin
			origin := engo.Point{X: engo.Input.Mouse.X, Y: engo.Input.Mouse.Y}
			s.cursor.selection.SpaceComponent = common.SpaceComponent{
				Width:    0,
				Height:   0,
				Position: origin,
			}
			firstDragged = false
		} else {
			// Keep dragging -> increment selection box
			s.cursor.selection.SpaceComponent.Width = engo.Input.Mouse.X - s.cursor.selection.Position.X
			s.cursor.selection.SpaceComponent.Height = engo.Input.Mouse.Y - s.cursor.selection.Position.Y
			s.boxSelect(&s.cursor.selection)
		}

	} else {
		// Reset box and variables
		s.cursor.selection.SpaceComponent.Width = 0
		s.cursor.selection.SpaceComponent.Height = 0
		firstDragged = true
	}

}
