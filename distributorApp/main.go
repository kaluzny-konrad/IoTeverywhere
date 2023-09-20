package main

import (
    "log"
    "os"
    "path/filepath"
    "encoding/csv"
    "time"
    "strconv"
    "io"
)

type TemperatureRecord struct {
    Timestamp  time.Time
    Temperature float64
}

func main() {
    dataFolder := "../devices/termometer/data"  // Ścieżka do folderu z danymi

    for {
        // Pobierz listę plików CSV z folderu
        files, err := filepath.Glob(filepath.Join(dataFolder, "*.csv"))
        if err != nil {
            log.Printf("Błąd podczas pobierania listy plików CSV: %s", err)
            time.Sleep(1 * time.Second)
            continue
        }

        for _, filename := range files {
            // Pobierz datę modyfikacji pliku
            fileInfo, err := os.Stat(filename)
            if err != nil {
                log.Printf("Błąd podczas pobierania informacji o pliku: %s", err)
                continue
            }

            // Sprawdź, czy plik jest starszy niż 1 sekunda
            if time.Since(fileInfo.ModTime()) < 1*time.Second {
                continue
            }

            // Otwórz plik CSV do odczytu
            csvFile, err := os.Open(filename)
            if err != nil {
                log.Printf("Błąd podczas otwierania pliku CSV: %s", err)
                continue
            }

			// Parsuj plik CSV
			var records []TemperatureRecord
			csvFile.Seek(0, 0) // Przesuń wskaźnik pliku na początek
			csvReader := csv.NewReader(csvFile)
			firstRow := true
			for {
				record, err := csvReader.Read()
				if err == io.EOF {
					break
				} else if err != nil {
					log.Printf("Błąd podczas odczytywania rekordu CSV: %s", err)
					csvFile.Close()
					continue
				}

				// Pomijaj pierwszy wiersz (nagłówek)
				if firstRow {
					firstRow = false
					continue
				}

				// Parsuj rekord i dodaj go do listy records
				// (zakładam, że record[0] to timestamp, record[1] to temperatura)
				timestamp, err := time.Parse("2006-01-02 15:04:05", record[0])
				if err != nil {
					log.Printf("Błąd podczas parsowania znacznika czasu: %s", err)
					continue
				}

				temperature, err := strconv.ParseFloat(record[1], 64)
				if err != nil {
					log.Printf("Błąd podczas parsowania temperatury: %s", err)
					continue
				}

				recordToInsert := TemperatureRecord{Timestamp: timestamp, Temperature: temperature}
				records = append(records, recordToInsert)
			}

            // Po przeczytaniu usuń plik
            csvFile.Close()
            if err := os.Remove(filename); err != nil {
                log.Printf("Błąd podczas usuwania pliku: %s", err)
            }

            // Zapisz dane do pliku saved_data.csv
            savedDataFile, err := os.OpenFile("saved_data.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
            if err != nil {
                log.Printf("Błąd podczas otwierania pliku saved_data.csv: %s", err)
                continue
            }
            savedDataWriter := csv.NewWriter(savedDataFile)
            for _, record := range records {
                savedDataWriter.Write([]string{record.Timestamp.Format("2006-01-02 15:04:05"), strconv.FormatFloat(record.Temperature, 'f', -1, 64)})
            }
            savedDataWriter.Flush()
            savedDataFile.Close()
        }

        // Oczekuj na kolejne sprawdzenie plików
        time.Sleep(1 * time.Second)
    }
}
