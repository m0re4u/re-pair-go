package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/common"
)

// Player entity
type Player struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

// followEntity mouse follower
type followEntity struct {
	*ecs.BasicEntity
	*common.RenderComponent
	*common.SpaceComponent
}

// MouseFollower things that follow the mouse
type MouseFollower struct {
	entities []followEntity
}

// Add a follower to the mouse
func (s *MouseFollower) Add(basic *ecs.BasicEntity, render *common.RenderComponent, space *common.SpaceComponent) {
	s.entities = append(s.entities, followEntity{basic, render, space})
}

// Remove a follower from the mouse
func (s *MouseFollower) Remove(basic ecs.BasicEntity) {
	delete := -1
	for index, e := range s.entities {
		if e.BasicEntity.ID() == basic.ID() {
			delete = index
			break
		}
	}

	if delete >= 0 {
		s.entities = append(s.entities[:delete], s.entities[delete+1:]...)
	}
}

// Update a mouse follower's position
func (s *MouseFollower) Update(dt float32) {
	for _, e := range s.entities {
		e.SpaceComponent.Position.X = engo.Input.Mouse.X
		e.SpaceComponent.Position.Y = engo.Input.Mouse.Y
	}
}