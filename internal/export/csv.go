package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/stpnv0/WarehouseControl/internal/domain"
)

var auditCSVHeader = []string{
	"ID",
	"Item ID",
	"Action",
	"Changed By (ID)",
	"Changed By (Username)",
	"Changed At",
	"Changes",
}

func WriteAuditCSV(w io.Writer, entries []*domain.AuditEntryWithUser) error {
	cw := csv.NewWriter(w)

	if err := cw.Write(auditCSVHeader); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for _, e := range entries {
		row := []string{
			fmt.Sprintf("%d", e.ID),
			e.ItemID.String(),
			string(e.Action),
			e.ChangedBy.String(),
			e.Username,
			e.ChangedAt.Format(time.RFC3339),
			formatDiff(e.Diff),
		}

		if err := cw.Write(row); err != nil {
			return fmt.Errorf("write row %d: %w", e.ID, err)
		}
	}

	cw.Flush()
	return cw.Error()
}

func formatDiff(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	var fields map[string]struct {
		Old interface{} `json:"old"`
		New interface{} `json:"new"`
	}

	if err := json.Unmarshal(raw, &fields); err != nil {
		return string(raw)
	}

	var result string
	for field, v := range fields {
		if result != "" {
			result += "; "
		}
		result += fmt.Sprintf("%s: %v -> %v", field, v.Old, v.New)
	}

	return result
}
