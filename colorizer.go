package main

import (
	"barista.run/colors"
)

type Comparator struct {
	String  func(string) bool
	Float64 func(float64) bool
	Color   colors.ColorfulColor
}

func Float(fn func(float64) bool, c colors.ColorfulColor) Comparator {
	return Comparator{nil, fn, c}
}

func String(fn func(string) bool, c colors.ColorfulColor) Comparator {
	return Comparator{fn, nil, c}
}

type Colorizer struct {
	base colors.ColorfulColor
	cmp  []Comparator
}

func NewColorizer(base colors.ColorfulColor, cmprs ...Comparator) *Colorizer {
	return &Colorizer{base, cmprs}
}

func (c *Colorizer) Float64(v float64) colors.ColorfulColor {
	for _, cmp := range c.cmp {
		if cmp.Float64 != nil && cmp.Float64(v) {
			return cmp.Color
		}
	}
	return c.base
}

func (c *Colorizer) Int(v int) colors.ColorfulColor {
	return c.Float64(float64(v))
}

func (c *Colorizer) String(v string) colors.ColorfulColor {
	for _, cmp := range c.cmp {
		if cmp.String != nil && cmp.String(v) {
			return cmp.Color
		}
	}
	return c.base
}
