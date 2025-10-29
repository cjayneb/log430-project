package core_test

import (
	"brokerx/market-data-service/core"
	"brokerx/market-data-service/models"
	"testing"
)

const resourcesPath = "../resources/"

func TestMarketDataServiceImpl_GetCurrentStockPriceBySymbol(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		symbol  string
		want    float64
		wantErr bool
	}{
		{
			name:    "returns correct price when given symbol",
			symbol:  "AAPL",
			want:    176.35,
			wantErr: false,
		},
		{
			name:    "returns error when symbol not found",
			symbol:  "AAPLLL",
			want:    0.0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := core.NewMarketDataServiceImpl(resourcesPath)

			got, gotErr := m.GetCurrentStockPriceBySymbol(tt.symbol)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetCurrentStockPriceBySymbol() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetCurrentStockPriceBySymbol() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("GetCurrentStockPriceBySymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarketDataServiceImpl_GetInstrumentBySymbol(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		symbol  string
		want    *models.Instrument
		wantErr bool
	}{
		{
			name:   "returns correct instrument when symbol is found",
			symbol: "AAPL",
			want: &models.Instrument{
				Symbol:           "AAPL",
				Name:             "Apple Inc.",
				TickSize:         0.01,
				PriceBandPercent: 10,
				Status:           "Active",
			},
			wantErr: false,
		},
		{
			name:    "returns error when symbol not found",
			symbol:  "AAPawdL",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := core.NewMarketDataServiceImpl(resourcesPath)

			got, gotErr := m.GetInstrumentBySymbol(tt.symbol)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetInstrumentBySymbol() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetInstrumentBySymbol() succeeded unexpectedly")
			}
			if *got != *tt.want {
				t.Errorf("GetInstrumentBySymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}
