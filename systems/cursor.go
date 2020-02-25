package systems

import (
	"log"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"
)

// MouseCursor entity for drawing the cursor
type MouseCursor struct {
	base   ecs.BasicEntity
	render common.RenderComponent
	space  common.SpaceComponent
}

// MouseFollower system that controls the cursor
type MouseFollower struct {
	world *ecs.World

	cursor MouseCursor
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
	}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&s.cursor.base, &s.cursor.render, &s.cursor.space)
		}
	}

}

// Remove follower system from the mouse
func (*MouseFollower) Remove(basic ecs.BasicEntity) {}

// Update mouse follower's position
func (s *MouseFollower) Update(dt float32) {
	s.cursor.space.Position.X = engo.Input.Mouse.X
	s.cursor.space.Position.Y = engo.Input.Mouse.Y
}
