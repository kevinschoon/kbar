package main

import (
	"barista.run/colors"
	"barista.run/pango"
)

type StyleFunc func (*pango.Node) *pango.Node

type Comparator struct {
	String  func(string) bool
	Float64 func(float64) bool
	style   StyleFunc
}

func Float(fn func(float64) bool, s StyleFunc) Comparator {
	return Comparator{nil, fn, s}
}

func String(fn func(string) bool, s StyleFunc) Comparator {
	return Comparator{fn, nil, s}
}

type Styleizer struct {
    base StyleFunc
    styles []Comparator
}

func NewStyleizer(base StyleFunc, styles ...Comparator) Styleizer {
    return Styleizer{base, styles}
}

func SetColor(color colors.ColorfulColor) StyleFunc {
    return func(node *pango.Node) *pango.Node {
        return node.Color(color)
    }
}

func (s *Styleizer) Float64(v float64) StyleFunc {
	for _, cmp := range s.styles {
		if cmp.Float64 != nil && cmp.Float64(v) {
			return cmp.style
		}
	}
	return s.base
}

func (s *Styleizer) Int(v int) StyleFunc {
	return s.Float64(float64(v))
}

func (s *Styleizer) String(v string) StyleFunc {
	for _, cmp := range s.styles {
		if cmp.String != nil && cmp.String(v) {
			return cmp.style
		}
	}
	return s.base
}
