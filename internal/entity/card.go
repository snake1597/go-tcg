package entity

import (
	"database/sql"
	"time"
)

type Card struct {
	Classes      []string       `json:"classes"`
	CostType     string         `json:"cost_type"`
	CostValue    string         `json:"cost_value"`
	CreatedAt    time.Time      `json:"created_at"`
	Durability   sql.NullInt64  `json:"durability"`
	Effect       string         `json:"effect"`
	EffectRaw    string         `json:"effect_raw"`
	Elements     []string       `json:"elements"`
	Flavor       string         `json:"flavor"`
	LastUpdate   time.Time      `json:"last_update"`
	Legality     sql.NullString `json:"legality"`
	Level        sql.NullInt64  `json:"level"`
	Life         sql.NullInt64  `json:"life"`
	Name         string         `json:"name"`
	Power        sql.NullInt64  `json:"power"`
	ReferencedBy []string       `json:"referenced_by"`
	References   []*Reference   `json:"references"`
	Rule         []string       `json:"rule"`
	Slug         string         `json:"slug"`
	Speed        sql.NullInt64  `json:"speed"`
	Subtypes     []string       `json:"subtypes"`
	Types        []string       `json:"types"`
	UUID         string         `json:"uuid"`
}

type Reference struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Direction string `json:"direction"`
}
