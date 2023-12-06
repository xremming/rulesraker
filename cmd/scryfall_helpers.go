package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type CardSymbolList struct {
	Data []CardSymbol
}

type CardSymbol struct {
	Symbol string
	SVGURI string `json:"svg_uri"`
}

func getSymbolsReplacer() (*strings.Replacer, error) {
	resp, err := http.DefaultClient.Get("https://api.scryfall.com/symbology")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data CardSymbolList

	dec := json.NewDecoder(resp.Body)

	err = dec.Decode(&data)
	if err != nil {
		return nil, err
	}

	replace := make([]string, 0, 2*len(data.Data))
	for _, cardSymbol := range data.Data {
		replace = append(replace, cardSymbol.Symbol, fmt.Sprintf(
			`<img class="symbol" title="%v" alt="%v" src="%v">`,
			cardSymbol.Symbol, cardSymbol.Symbol, cardSymbol.SVGURI,
		))
	}

	return strings.NewReplacer(replace...), nil
}
