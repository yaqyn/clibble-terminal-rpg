package main

import (
	"fmt"
	"time"

	"github.com/yaqyn/clibble-terminal-rpg/internal/pokeapi"
)

type journalRecords = map[string]pokeapi.Pokemon

const seedCapacity = 12
const seedInterval = 5 * time.Minute
const encounterWindow = 30 * time.Minute

type encounterBudget struct {
	Offers int       `json:"offers"`
	Until  time.Time `json:"until"`
}
type gameProgress struct {
	Seeds      int                        `json:"seeds"`
	RefillAt   time.Time                  `json:"refill_at"`
	Encounters map[string]encounterBudget `json:"encounters"`
	MotionOff  bool                       `json:"motion_off,omitempty"`
}

func freshProgress(now time.Time) gameProgress {
	return gameProgress{Seeds: seedCapacity, RefillAt: now, Encounters: map[string]encounterBudget{}}
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
		units := int(now.Sub(p.RefillAt) / seedInterval)
		next.Seeds = min(seedCapacity, p.Seeds+units)
		if next.Seeds == seedCapacity {
			next.RefillAt = now
		} else {
			next.RefillAt = p.RefillAt.Add(time.Duration(units) * seedInterval)
		}
	}
	return next
}
func (p gameProgress) allowance(name string, now time.Time) error {
	p = p.refreshed(now)
	if e, ok := p.Encounters[name]; ok && e.Offers >= 3 {
		return fmt.Errorf("%s needs a rest; try again in %s", placeName(name), e.Until.Sub(now).Round(time.Second))
	}
	if p.Seeds == 0 {
		return fmt.Errorf("your pouch is empty; your next catch seed arrives in %s", p.RefillAt.Add(seedInterval).Sub(now).Round(time.Second))
	}
	return nil
}
func (c *config) now() time.Time {
	if c.clock != nil {
		return c.clock()
	}
	return time.Now()
}
func (c *config) persist(owned journalRecords, p gameProgress) error {
	if c.saveProgress != nil {
		return c.saveProgress(owned, p)
	}
	if c.save != nil {
		return c.save(owned)
	}
	return nil
}
