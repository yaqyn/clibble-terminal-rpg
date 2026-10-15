package main

import (
	"fmt"
	"github.com/yaqyn/clibble-pokemon-cli/internal/pokeapi"
	"time"
)

type mapPokemon = map[string]pokeapi.Pokemon

const ballCapacity = 12
const ballInterval = 5 * time.Minute
const encounterWindow = 30 * time.Minute

type encounterBudget struct {
	Throws int       `json:"throws"`
	Until  time.Time `json:"until"`
}
type gameProgress struct {
	Balls      int                        `json:"balls"`
	RefillAt   time.Time                  `json:"refill_at"`
	Encounters map[string]encounterBudget `json:"encounters"`
	MotionOff  bool                       `json:"motion_off,omitempty"`
}

func freshProgress(now time.Time) gameProgress {
	return gameProgress{Balls: ballCapacity, RefillAt: now, Encounters: map[string]encounterBudget{}}
}
func (p gameProgress) refreshed(now time.Time) gameProgress {
	if p.Encounters == nil {
		return freshProgress(now)
	}
	next := p
	next.Encounters = make(map[string]encounterBudget, len(p.Encounters))
	for name, e := range p.Encounters {
		if now.Before(e.Until) {
			next.Encounters[name] = e
		}
	}
	if now.After(p.RefillAt) {
		units := int(now.Sub(p.RefillAt) / ballInterval)
		next.Balls = min(ballCapacity, p.Balls+units)
		if next.Balls == ballCapacity {
			next.RefillAt = now
		} else {
			next.RefillAt = p.RefillAt.Add(time.Duration(units) * ballInterval)
		}
	}
	return next
}
func (p gameProgress) allowance(name string, now time.Time) error {
	p = p.refreshed(now)
	if e, ok := p.Encounters[name]; ok && e.Throws >= 3 {
		return fmt.Errorf("%s needs a rest; try again in %s", placeName(name), e.Until.Sub(now).Round(time.Second))
	}
	if p.Balls == 0 {
		return fmt.Errorf("your pouch is empty; your next Pokeball arrives in %s", p.RefillAt.Add(ballInterval).Sub(now).Round(time.Second))
	}
	return nil
}
func (c *config) now() time.Time {
	if c.clock != nil {
		return c.clock()
	}
	return time.Now()
}
func (c *config) persist(owned mapPokemon, p gameProgress) error {
	if c.saveProgress != nil {
		return c.saveProgress(owned, p)
	}
	if c.save != nil {
		return c.save(owned)
	}
	return nil
}
