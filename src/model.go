package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
)

// A Model is an input-output model in the data folder.
type Model struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Folder      string        `json:"-"`
	Sectors     []*Sector     `json:"-"`
	Indicators  []*Indicator  `json:"-"`
	DemandInfos []*DemandInfo `json:"-"`

	sectorMap map[string]*Sector
	numCache  map[string]*Matrix
	dqiCache  map[string][][]string
}

// NewModel creates a new model from the given folder (which should be a
// sub-folder of the data directory).
func NewModel(id, folder string) (*Model, error) {
	sectors, err := ReadSectors(folder)
	if err != nil {
		return nil, err
	}
	sectorMap := make(map[string]*Sector)
	for i := range sectors {
		s := sectors[i]
		sectorMap[s.ID] = s
	}
	indicators, err := ReadIndicators(folder)
	if err != nil {
		return nil, err
	}
	demands, err := ReadDemandInfos(folder)
	if err != nil {
		return nil, err
	}
	m := Model{
		ID:          id,
		Name:        id,
		Folder:      folder,
		Sectors:     sectors,
		Indicators:  indicators,
		DemandInfos: demands,
		sectorMap:   sectorMap,
		numCache:    make(map[string]*Matrix),
		dqiCache:    make(map[string][][]string)}
	return &m, nil
}

// Sector returns the sector with the given ID.
func (m *Model) Sector(id string) *Sector {
	return m.sectorMap[id]
}

// SectorIDs returns the IDs of the sectors in the index order.
func (m *Model) SectorIDs() []string {
	ids := make([]string, len(m.Sectors))
	for i := range m.Sectors {
		ids[i] = m.Sectors[i].ID
	}
	return ids
}

// IndicatorIDs returns the IDs of the indicators in the index order.
func (m *Model) IndicatorIDs() []string {
	ids := make([]string, len(m.Indicators))
	for i := range m.Indicators {
		ids[i] = m.Indicators[i].ID
	}
	return ids
}

// Matrix returns the numeric matrix with the given name (e.g. `A`) from the
// model.
func (m *Model) Matrix(name string) (*Matrix, error) {
	matrix := m.numCache[name]
	if matrix != nil {
		return matrix, nil
	}
	file := filepath.Join(m.Folder, name+".bin")
	var err error
	matrix, err = Load(file)
	if err != nil {
		return nil, err
	}
	m.numCache[name] = matrix
	return matrix, nil
}

// DqiMatrix returns the DQI matrix with the given name (e.g. `B_dqi`) from the
// model.
func (m *Model) DqiMatrix(name string) ([][]string, error) {
	matrix := m.dqiCache[name]
	if matrix != nil {
		return matrix, nil
	}
	file := filepath.Join(m.Folder, name+".csv")
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	m.dqiCache[name] = records
	return records, nil
}
