package main

import (
	"math"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type Option func(*GamePerson)

func read24(b [3]byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
}

func write24(b *[3]byte, v uint32) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
}

func WithName(name string) func(*GamePerson) {
	return func(p *GamePerson) {
		copy(p.name[:], name)
	}
}

func WithCoordinates(x, y, z int) func(*GamePerson) {
	return func(p *GamePerson) {
		p.x = int32(x)
		p.y = int32(y)
		p.z = int32(z)
	}
}

func WithGold(gold int) func(*GamePerson) {
	return func(p *GamePerson) {
		p.gold = uint32(gold)
	}
}

func WithMana(mana int) func(*GamePerson) {
	return func(person *GamePerson) {
		v := read24(person.hi)
		// 1111 1111 1111 1111 1111 1100 0000 0000
		v &^= (0x3FF << 0)
		// 0000 0000 0000 0000 0000 0011 1111 1111
		v |= uint32(mana&0x3FF) << 0
		write24(&person.hi, v)
	}
}

func WithHealth(health int) func(*GamePerson) {
	return func(person *GamePerson) {
		v := read24(person.hi)
		v &^= (0x3FF << 10)
		v |= uint32(health&0x3FF) << 10
		write24(&person.hi, v)
	}
}

func WithRespect(respect int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.respStr &^= 0x0F
		person.respStr |= byte(respect & 0x0F)
	}
}

func WithStrength(strength int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.respStr &^= 0xF0
		person.respStr |= byte((strength & 0x0F) << 4)
	}
}

func WithExperience(experience int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.expLevel &^= 0x0F // 00001111
		person.expLevel |= byte(experience & 0x0F)
	}
}

func WithLevel(level int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.expLevel &^= 0xF0 // 11110000
		person.expLevel |= byte((level & 0x0F) << 4)
	}
}

func WithHouse() func(*GamePerson) {
	return func(person *GamePerson) {
		person.flagsType |= 1 << 0
	}
}

func WithGun() func(*GamePerson) {
	return func(person *GamePerson) {
		person.flagsType |= 1 << 1
	}
}

func WithFamily() func(*GamePerson) {
	return func(person *GamePerson) {
		person.flagsType |= 1 << 2
	}
}

func WithType(personType int) func(*GamePerson) {
	return func(person *GamePerson) {
		person.flagsType &^= 0x18
		person.flagsType |= byte(personType&0x03) << 3
	}
}

const (
	BuilderGamePersonType = iota
	BlacksmithGamePersonType
	WarriorGamePersonType
)

type GamePerson struct {
	name [42]byte // 42B
	hi   [3]byte  // 3B 24 бита: мана (10) + здоровье (10) + 4 резерв
	// 3B 24 бита
	respStr   byte   // уважение(4)+сила(4)
	expLevel  byte   // опыт(4)+уровень(4)
	flagsType byte   // дом(1)+оружие(1)+семья(1)+тип(2)+ 3резерв
	x, y, z   int32  // 12B
	gold      uint32 // 4B
}

func NewGamePerson(options ...Option) GamePerson {
	var p GamePerson
	for _, opt := range options {
		opt(&p)
	}
	return p
}

func (p *GamePerson) Name() string {
	return string(p.name[:])
}

func (p *GamePerson) X() int {
	return int(p.x)
}

func (p *GamePerson) Y() int {
	return int(p.y)
}

func (p *GamePerson) Z() int {
	return int(p.z)
}

func (p *GamePerson) Gold() int {
	return int(p.gold)
}

func (p *GamePerson) Mana() int {
	v := read24(p.hi)
	return int((v >> 0) & 0x3FF)
}

func (p *GamePerson) Health() int {
	v := read24(p.hi)
	return int((v >> 10) & 0x3FF)
}

func (p *GamePerson) Respect() int {
	return int(p.respStr & 0x0F)
}

func (p *GamePerson) Strength() int {
	return int((p.respStr >> 4) & 0x0F)
}

func (p *GamePerson) Experience() int {
	return int(p.expLevel & 0x0F)
}

func (p *GamePerson) Level() int {
	return int((p.expLevel >> 4) & 0x0F)
}

func (p *GamePerson) HasHouse() bool {
	return p.flagsType&(1<<0) != 0
}

func (p *GamePerson) HasGun() bool {
	return p.flagsType&(1<<1) != 0
}

func (p *GamePerson) HasFamilty() bool {
	return p.flagsType&(1<<2) != 0
}

func (p *GamePerson) Type() int {
	return int((p.flagsType >> 3) & 0x03)
}

func TestGamePerson(t *testing.T) {
	assert.LessOrEqual(t, unsafe.Sizeof(GamePerson{}), uintptr(64))

	const x, y, z = math.MinInt32, math.MaxInt32, 0
	const name = "aaaaaaaaaaaaa_bbbbbbbbbbbbb_cccccccccccccc"
	const personType = BuilderGamePersonType
	const gold = math.MaxInt32
	const mana = 1000
	const health = 1000
	const respect = 10
	const strength = 10
	const experience = 10
	const level = 10

	options := []Option{
		WithName(name),
		WithCoordinates(x, y, z),
		WithGold(gold),
		WithMana(mana),
		WithHealth(health),
		WithRespect(respect),
		WithStrength(strength),
		WithExperience(experience),
		WithLevel(level),
		WithHouse(),
		WithFamily(),
		WithType(personType),
	}

	person := NewGamePerson(options...)
	assert.Equal(t, name, person.Name())
	assert.Equal(t, x, person.X())
	assert.Equal(t, y, person.Y())
	assert.Equal(t, z, person.Z())
	assert.Equal(t, gold, person.Gold())
	assert.Equal(t, mana, person.Mana())
	assert.Equal(t, health, person.Health())
	assert.Equal(t, respect, person.Respect())
	assert.Equal(t, strength, person.Strength())
	assert.Equal(t, experience, person.Experience())
	assert.Equal(t, level, person.Level())
	assert.True(t, person.HasHouse())
	assert.True(t, person.HasFamilty())
	assert.False(t, person.HasGun())
	assert.Equal(t, personType, person.Type())
}
