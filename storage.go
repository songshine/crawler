package crawler

import (
	"encoding/csv"
	"log"
	"os"
)

const flushWindow = 10

type DataStorager interface {
	Persist(DataCollector)
}

func NewCSVDataStorage(path string) DataStorager {
	return &csvDataStorage{path}
}

type csvDataStorage struct {
	path string
}

func (s *csvDataStorage) Persist(c DataCollector) {
	file, err := os.Create(s.path)
	if err != nil {
		log.Fatal("Persist: create csv file failed", err)
	}
	writer := csv.NewWriter(file)
	defer writer.Flush()
	var (
		total   = 0
		names   = c.Names()
		results = c.Collect()
	)

	writer.Write(names)
	for r := range results {
		writer.Write(r)
		total++
		log.Printf(">>> Result count received in storage: %d \n", total)
		if total%flushWindow == 0 {
			writer.Flush()
		}

	}
}
