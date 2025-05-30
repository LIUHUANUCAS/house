package main

import (
	"sync"
)

var beijing *Beijing
var shanghai *Shanghai
var fortune *Fortune

type DataAccessor interface {
	GetDB() *sync.Map
}
type Beijing struct {
	DB *sync.Map
}

func (b *Beijing) GetDB() *sync.Map {
	return b.DB
}

type Shanghai struct {
	DB *sync.Map
}

func (s *Shanghai) GetDB() *sync.Map {
	return s.DB
}

type Fortune struct {
	DB *sync.Map
}

func (f *Fortune) GetDB() *sync.Map {
	return f.DB
}

// InitInMemoryDB initializes the in-memory databases for Beijing, Shanghai, and Fortune data.
func InitInMemoryDB() {
	// Initialize in-memory databases
	beijing = &Beijing{DB: &sync.Map{}}
	shanghai = &Shanghai{DB: &sync.Map{}}
	fortune = &Fortune{DB: &sync.Map{}}
}

// GetInMemDataAccessor retrieves the in-memory data accessor for the specified factory.
func GetInMemDataAccessor(d DataAccessor) *sync.Map {
	return d.GetDB()
}

func getDB(factory string) *sync.Map {

	switch factory {
	case "daily":
		return GetInMemDataAccessor(beijing)
	case "sh":
		return GetInMemDataAccessor(shanghai)
	case "fortune":
		return GetInMemDataAccessor(fortune)
	}

	return &sync.Map{}
}
