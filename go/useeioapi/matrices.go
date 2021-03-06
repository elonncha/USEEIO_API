package main

import (
	"encoding/csv"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
)

// HandleGetMatrix returns the handler for GET /api/{model}/matrix/{matrix}
func HandleGetMatrix(dataDir string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		model := mux.Vars(r)["model"]
		name := mux.Vars(r)["matrix"]

		// get possible row or column
		col, err := indexParam("col", r.URL, w)
		if err != nil {
			return
		}
		row, err := indexParam("row", r.URL, w)
		if err != nil {
			return
		}

		switch name {
		case "A", "B", "C", "D", "L", "U":
			file := filepath.Join(dataDir, model, name+".bin")
			matrix, err := LoadMatrix(file)
			if err != nil {
				http.Error(w, "Failed to load matrix "+name,
					http.StatusInternalServerError)
				return
			}
			serveMatrix(matrix, row, col, w)
		case "B_dqi", "D_dqi", "U_dqi":
			file := filepath.Join(dataDir, model, name+".csv")
			dqis, err := readDqiMatrix(file)
			if err != nil {
				http.Error(w, "Failed to load matrix", http.StatusInternalServerError)
				return
			}
			serveDqiMatrix(dqis, row, col, w)
		default:
			http.Error(w, "Unknown matrix: "+name, http.StatusNotFound)
		}
	}
}

func serveMatrix(matrix *Matrix, row int, col int, w http.ResponseWriter) {
	// return a single column
	if col > -1 {
		if col >= matrix.Cols {
			http.Error(w, "Column out of bounds", http.StatusBadRequest)
			return
		}
		ServeJSON(matrix.Col(col), w)
		return
	}

	// return a single row
	if row > -1 {
		if row >= matrix.Rows {
			http.Error(w, "Row out of bound", http.StatusBadRequest)
			return
		}
		ServeJSON(matrix.Row(row), w)
		return
	}

	// return the full matrix
	ServeJSON(matrix.Slice2d(), w)
}

func readDqiMatrix(file string) ([][]string, error) {
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
	return records, nil
}

func serveDqiMatrix(dqis [][]string, row int, col int, w http.ResponseWriter) {
	// return a single column
	if col > -1 {
		vals := make([]string, len(dqis))
		for row, rowVals := range dqis {
			if rowVals == nil || len(rowVals) <= col {
				http.Error(w, "Column out of bounds", http.StatusBadRequest)
				return
			}
			vals[row] = rowVals[col]
		}
		ServeJSON(vals, w)
		return
	}

	// return a single row
	if row > -1 {
		if row >= len(dqis) {
			http.Error(w, "Row out of bounds", http.StatusBadRequest)
			return
		}
		ServeJSON(dqis[row], w)
		return
	}

	// return the full matrix
	ServeJSON(dqis, w)
}

func indexParam(name string, reqURL *url.URL, w http.ResponseWriter) (int, error) {
	str := reqURL.Query().Get(name)
	if str == "" {
		return -1, nil
	}
	idx, err := strconv.Atoi(str)
	if err != nil || idx < 0 {
		http.Error(w, "Invalid index: "+name+"="+str, http.StatusBadRequest)
		return -1, err
	}
	return idx, nil
}
