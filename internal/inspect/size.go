package inspect

import (
	"encoding/json"

	libghostty "go.mitchellh.com/libghostty"
)

type SizeResult struct {
	Cols     uint16 `json:"cols"`
	Rows     uint16 `json:"rows"`
	WidthPx  uint32 `json:"width_px"`
	HeightPx uint32 `json:"height_px"`
}

func GetSize(term *libghostty.Terminal) *SizeResult {
	cols, _ := term.Cols()
	rows, _ := term.Rows()
	widthPx, _ := term.WidthPx()
	heightPx, _ := term.HeightPx()

	return &SizeResult{
		Cols:     cols,
		Rows:     rows,
		WidthPx:  widthPx,
		HeightPx: heightPx,
	}
}

func (s *SizeResult) ToJSON() string {
	data, _ := json.Marshal(s)
	return string(data)
}
