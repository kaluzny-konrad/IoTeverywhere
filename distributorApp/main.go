package main

import (
	"log"
	"os"
	"path/filepath"
	"encoding/csv"
	"time"
	"strconv"
	"io"
	"fmt"

	"github.com/gocql/gocql"
	"github.com/streadway/amqp"
)

type TemperatureRecord struct {
	Timestamp   time.Time
	Temperature float64
}

func main() {
	dataFolder := "../devices/termometer/data" // Ścieżka do folderu z danymi

	// Połącz się z bazą danych Cassandra
	cassandraHosts := os.Getenv("CASSANDRA_HOSTS")
	cassandraKeyspace := os.Getenv("CASSANDRA_KEYSPACE")
	cluster := gocql.NewCluster(cassandraHosts)
	cluster.Keyspace = cassandraKeyspace
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal("Błąd podczas łączenia z bazą danych Cassandra:", err)
	}
	defer session.Close()

	// Połącz się z RabbitMQ
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	queue := os.Getenv("QUEUE")
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatal("Błąd podczas łączenia z RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Błąd podczas tworzenia kanału RabbitMQ:", err)
	}
	defer ch.Close()

	for {
		// Pobierz listę plików CSV z folderu
		files, err := filepath.Glob(filepath.Join(dataFolder, "temperature_*.csv"))
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

			for _, record := range records {
				// Zapisz rekord do bazy danych Cassandra
				if err := session.Query("INSERT INTO temperature (timestamp, temperature) VALUES (?, ?)", record.Timestamp, record.Temperature).Exec(); err != nil {
					log.Printf("Błąd podczas zapisywania do Cassandra: %s", err)
					continue
				}

				// Wysyłanie rekordu na RabbitMQ
				err := ch.Publish(
					"",     // Exchange
					queue,  // Routing Key
					false,  // Mandatory
					false,  // Immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(fmt.Sprintf("Nowy rekord: %+v", record)),
					})
				if err != nil {
					log.Printf("Błąd podczas wysyłania na RabbitMQ: %s", err)
				}
			}
		}

		// Oczekuj na kolejne sprawdzenie plików
		time.Sleep(1 * time.Second)
	}
}
