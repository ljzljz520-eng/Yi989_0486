package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"couponbatch/internal/model"
)

func CSV(batches []model.Batch) (string, error) {
	var builder strings.Builder
	if err := WriteCSV(&builder, batches); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func WriteCSV(w io.Writer, batches []model.Batch) error {
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"id", "name", "description", "amount_cents", "currency", "total_stock", "remaining", "issued", "start_at", "end_at", "status"}); err != nil {
		return err
	}
	for _, batch := range batches {
		row := []string{batch.ID, batch.Name, batch.Description, strconv.FormatInt(batch.AmountCents, 10), batch.Currency, strconv.Itoa(batch.TotalStock), strconv.Itoa(batch.Remaining), strconv.Itoa(batch.Issued), batch.StartAt.Format("2006-01-02T15:04:05Z07:00"), batch.EndAt.Format("2006-01-02T15:04:05Z07:00"), string(batch.Status)}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func Summary(batches []model.Batch) string {
	active, stock := 0, 0
	for _, batch := range batches {
		if batch.Status == model.StatusActive {
			active++
		}
		stock += batch.Remaining
	}
	return fmt.Sprintf("batches=%d active=%d remaining=%d", len(batches), active, stock)
}
