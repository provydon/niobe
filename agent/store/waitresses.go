package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type MenuItem struct {
	ID          int64
	WaitressID  int64
	Name        string
	Category    string
	UnitPrice   float64
	Position    int
}

type Waitress struct {
	ID               int64
	Name             string
	Slug             string
	Context          string
	MenuCurrency     string
	TablesCount      int // number of tables (1..N); 0 means not set
	MenuItems        []MenuItem
	ExtractedContext []ExtractedContextItem
	Tools            []Tool
}

type Tool struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Target string `json:"target"`
}

type ExtractedContextItem struct {
	Filename string         `json:"filename"`
	Data     map[string]any `json:"data"`
	Error    string         `json:"error"`
}

type WaitressRepository struct {
	db *sql.DB
}

func NewWaitressRepository(db *sql.DB) *WaitressRepository {
	return &WaitressRepository{db: db}
}

func (r *WaitressRepository) FindBySlug(ctx context.Context, slug string) (*Waitress, error) {
	row := r.db.QueryRowContext(
		ctx,
		`select id, name, slug, context, menu_currency, coalesce(tables_count, 0), extracted_context, tools from waitresses where slug = $1 limit 1`,
		slug,
	)

	var waitress Waitress
	var menuCurrency sql.NullString
	var tablesCount sql.NullInt64
	var extractedRaw []byte
	var toolsRaw []byte

	if err := row.Scan(
		&waitress.ID,
		&waitress.Name,
		&waitress.Slug,
		&waitress.Context,
		&menuCurrency,
		&tablesCount,
		&extractedRaw,
		&toolsRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("waitress not found: %w", err)
		}

		return nil, err
	}

	if menuCurrency.Valid {
		waitress.MenuCurrency = menuCurrency.String
	}
	if tablesCount.Valid && tablesCount.Int64 > 0 {
		waitress.TablesCount = int(tablesCount.Int64)
	}

	if len(extractedRaw) > 0 {
		if err := json.Unmarshal(extractedRaw, &waitress.ExtractedContext); err != nil {
			return nil, fmt.Errorf("decode extracted context: %w", err)
		}
	}

	if len(toolsRaw) > 0 {
		if err := json.Unmarshal(toolsRaw, &waitress.Tools); err != nil {
			return nil, fmt.Errorf("decode tools: %w", err)
		}
	}

	rows, err := r.db.QueryContext(
		ctx,
		`select id, waitress_id, name, category, unit_price, position from menu_items where waitress_id = $1 order by position, id`,
		waitress.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("load menu items: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item MenuItem
		if err := rows.Scan(&item.ID, &item.WaitressID, &item.Name, &item.Category, &item.UnitPrice, &item.Position); err != nil {
			return nil, fmt.Errorf("scan menu item: %w", err)
		}
		waitress.MenuItems = append(waitress.MenuItems, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("menu items rows: %w", err)
	}

	return &waitress, nil
}

func (n *Waitress) FullContext() string {
	parts := []string{}

	if len(n.MenuItems) > 0 {
		lines := []string{"Menu (items, category, unit_price):"}
		if n.MenuCurrency != "" {
			lines = append(lines, "Currency: "+n.MenuCurrency+". When stating a price to the customer, always say it with this currency (e.g. 10 dollars, 35 euros).")
		}
		for _, item := range n.MenuItems {
			lines = append(lines, fmt.Sprintf("- %s | %s | %v", item.Name, item.Category, item.UnitPrice))
		}
		parts = append(parts, strings.Join(lines, "\n"))
	}

	if trimmed := strings.TrimSpace(n.Context); trimmed != "" {
		parts = append(parts, trimmed)
	}

	if len(n.ExtractedContext) > 0 {
		parts = append(parts, "\n--- Structured data from uploaded documents ---")

		for _, item := range n.ExtractedContext {
			filename := item.Filename
			if filename == "" {
				filename = "document"
			}

			if item.Error != "" {
				parts = append(parts, fmt.Sprintf("(%s: %s)", filename, item.Error))
				continue
			}

			if len(item.Data) == 0 {
				continue
			}

			data, err := json.MarshalIndent(item.Data, "", "  ")
			if err != nil {
				continue
			}

			parts = append(parts, fmt.Sprintf("From %s: %s", filename, string(data)))
		}
	}

	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func (n *Waitress) ToolsPrompt() string {
	if len(n.Tools) == 0 {
		return "No actions are configured."
	}

	lines := make([]string, 0, len(n.Tools))

	for _, tool := range n.Tools {
		line := fmt.Sprintf("- %s", toolPromptSummary(tool))
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func toolPromptSummary(tool Tool) string {
	name := strings.ToLower(strings.TrimSpace(tool.Name))
	if strings.Contains(name, "order") {
		return "Can place an order after confirming the details"
	}
	if strings.Contains(name, "issue") || strings.Contains(name, "support") || strings.Contains(name, "repair") || strings.Contains(name, "fix") {
		return "Can send the request to the appropriate team after confirming the details"
	}

	switch tool.Type {
	case "send_email", "send_webhook_event":
		return "Can send the request to the appropriate team after confirming the details"
	default:
		if trimmed := strings.TrimSpace(tool.Name); trimmed != "" {
			return trimmed
		}
		return tool.Type
	}
}
