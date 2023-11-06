package main

import (
	"encoding/csv"
	"github.com/gocolly/colly"
	"log"
	"os"
)

func main() {
	// Create a file to write the extracted data
	file, err := os.Create("qb_data.csv")
	if err != nil {
		log.Fatal("Failed to create file:", err)
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Define the column headers
	headers := []string{"Year", "Player", "Tm", "Age", "Pos", "G", "GS", "QBrec",
		"Cmp", "Att", "Cmp%", "Yds", "TD", "TD%", "Int", "Int%",
		"1D", "Lng", "Y/A", "AY/A", "Y/C", "Y/G", "Rate", "QBR",
		"Sk", "Yds", "NY/A", "ANY/A", "Sk%", "4QC", "GWD", "Playoffs"}

	// Write the headers to the CSV file
	err = writer.Write(headers)
	if err != nil {
		log.Fatal("Failed to write headers to file:", err)
	}

	// Scrape QB data for the given year
	scrapeQBData("2022", writer, 34)
	scrapeQBData("2021", writer, 34)
	scrapeQBData("2020", writer, 34)
}

// scrapeQBData scrapes QB data for the given year and writes it to the CSV writer
func scrapeQBData(year string, writer *csv.Writer, dataLimit int) {
	// new collector
	c := colly.NewCollector()
	// Go channel to signal the termination
	stopChan := make(chan struct{})
	// Helper function to determine playoffs value based on year and team
	getPlayoffsValue := func(year, team string) string {
		switch year {
		case "2022":
			switch team {
			case "KAN", "BUF", "CIN", "JAX", "LAC", "BAL", "MIA", "PHI", "SFO", "MIN", "TAM", "DAL", "NYG", "SEA":
				return "Y"
			default:
				return "N"
			}
		case "2021":
			switch team {
			case "KAN", "BUF", "CIN", "NWE", "TEN", "LVR", "PIT", "GNB", "SFO", "PHI", "TAM", "DAL", "LAR", "ARI":
				return "Y"
			default:
				return "N"
			}
		case "2020":
			switch team {
			case "KAN", "BUF", "IND", "TEN", "CLE", "BAL", "GNB", "NOR", "LAR", "WAS", "TAM", "CHI", "SEA":
				return "Y"
			default:
				return "N"
			}
		default:
			return "N" // Default to "N" if year doesn't match any conditions
		}
	}
	// Extract the QB data
	rowCount := 0
	c.OnHTML("#div_passing tbody tr", func(row_element *colly.HTMLElement) {
		// Skip the row with rowCount greater than the dataLimit
		rowCount++
		// Skip this weird header row
		if rowCount == 30 {
			return
		}
		// Stop collecting data
		if rowCount > dataLimit {
			return
		}
		var row []string
		row = append(row, year) // Add the year as the first column value
		row_element.ForEach("td", func(_ int, cell_element *colly.HTMLElement) {
			if cell_element.Attr("data-stat") == "qb_rec" {
				// makes the qb_rec to a str "11-4-0"
				row = append(row, `"`+cell_element.Text+`"`)
			} else {
				row = append(row, cell_element.Text)
			}
		})
		// Add the playoffs value at the end of the row
		team := row[2] // Tm column
		playoffs := getPlayoffsValue(year, team)
		row = append(row, playoffs)
		// Write the row to the CSV file
		err := writer.Write(row)
		if err != nil {
			log.Fatal("Failed to write row to file:", err)
		}
		// Stop
		if rowCount == dataLimit {
			close(stopChan)
		}
	})

	// Start the scraping process in a separate goroutine
	go func() {
		err := c.Visit("https://www.pro-football-reference.com/years/" + year + "/passing.htm")
		if err != nil {
			log.Fatal("Failed to visit website:", err)
		}
	}()
	// Wait for the stop signal
	<-stopChan
}
