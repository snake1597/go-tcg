package entity

import (
	"time"
)

type Card struct {
	Classes      []string     `json:"classes"`
	CostType     string       `json:"cost_type"`
	CostValue    string       `json:"cost_value"`
	CreatedAt    time.Time    `json:"created_at"`
	Durability   *int64       `json:"durability"`
	Effect       string       `json:"effect"`
	EffectRaw    string       `json:"effect_raw"`
	Elements     []string     `json:"elements"`
	Flavor       string       `json:"flavor"`
	LastUpdate   time.Time    `json:"last_update"`
	Legality     *string      `json:"legality"`
	Level        *int64       `json:"level"`
	Life         *int64       `json:"life"`
	Name         string       `json:"name"`
	Power        *int64       `json:"power"`
	ReferencedBy []string     `json:"referenced_by"`
	References   []*Reference `json:"references"`
	Rule         []*Rule      `json:"rule"`
	Slug         string       `json:"slug"`
	Speed        *bool        `json:"speed"`
	Subtypes     []string     `json:"subtypes"`
	Types        []string     `json:"types"`
	UUID         string       `json:"uuid"`
}

type Rule struct {
	DateAdded   string `json:"date_added"`
	Description string `json:"description"`
	Title       string `json:"title"`
}

type Reference struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Direction string `json:"direction"`
}
