package main

import (
	"bufio"
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func getUserSelection() string {
	fmt.Println("Digite o símbolo da criptomoeda (ex.: 'bitcoin', 'ethereum'):")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func fetchHistoricalData(symbol string, days int) ([]float64, error) {
	client := resty.New()
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s/market_chart?vs_currency=usd&days=%d", symbol, days)

	var response struct {
		Prices [][]float64 `json:"prices"`
	}

	_, err := client.R().SetResult(&response).Get(url)
	if err != nil {
		return nil, err
	}

	var prices []float64
	for _, price := range response.Prices {
		prices = append(prices, price[1])
	}

	return prices, nil
}

func calculateSupportResistance(prices []float64) (support float64, resistance float64) {
	support = prices[0]
	resistance = prices[0]

	for _, price := range prices {
		if price < support {
			support = price
		}
		if price > resistance {
			resistance = price
		}
	}
	return
}
func plotSupportResistance(prices []float64, supports []float64, resistances []float64, days int) {
	p := plot.New()
	p.Title.Text = "Análise de Suporte e Resistência"
	p.X.Label.Text = "Dias"
	p.Y.Label.Text = "Preço (USD)"

	// Prepara os pontos do gráfico
	pts := make(plotter.XYs, len(prices))
	for i, price := range prices {
		pts[i].X = float64(i) // X representa o índice do dia
		pts[i].Y = price
	}

	// Adiciona a linha dos preços
	line, err := plotter.NewLine(pts)
	if err != nil {
		panic(err)
	}
	p.Add(line)

	// Adiciona suportes e resistências como linhas horizontais
	for _, support := range supports {
		supportLine := plotter.NewFunction(func(x float64) float64 { return support })
		supportLine.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255} // Azul
		p.Add(supportLine)
	}
	for _, resistance := range resistances {
		resistanceLine := plotter.NewFunction(func(x float64) float64 { return resistance })
		resistanceLine.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255} // Vermelho
		p.Add(resistanceLine)
	}

	// Define os limites do eixo X
	p.X.Min = 0
	p.X.Max = float64(days - 1)

	// Salva o gráfico
	if err := p.Save(8*vg.Inch, 6*vg.Inch, "support_resistance.png"); err != nil {
		panic(err)
	}
}

func calculateMultipleSupportResistance(prices []float64) ([]float64, []float64) {
	// Arrays para armazenar suportes e resistências
	var supports []float64
	var resistances []float64

	// Define uma margem de tolerância para identificar níveis locais
	tolerance := 0.02 // 2% de variação em relação aos preços locais

	for i := 1; i < len(prices)-1; i++ {
		if prices[i] < prices[i-1] && prices[i] < prices[i+1] {
			// Verifica se é um suporte
			if len(supports) == 0 || prices[i] < supports[len(supports)-1]*(1-tolerance) {
				supports = append(supports, prices[i])
			}
		} else if prices[i] > prices[i-1] && prices[i] > prices[i+1] {
			// Verifica se é uma resistência
			if len(resistances) == 0 || prices[i] > resistances[len(resistances)-1]*(1+tolerance) {
				resistances = append(resistances, prices[i])
			}
		}

		// Limita a 3 níveis
		if len(supports) >= 3 && len(resistances) >= 3 {
			break
		}
	}

	return supports, resistances
}

func main() {
	symbol := getUserSelection()
	fmt.Printf("Analisando a criptomoeda: %s\n", symbol)

	days := 90 // Últimos 90 dias
	fmt.Printf("Analisando os últimos %d dias.\n", days)

	prices, err := fetchHistoricalData(symbol, days)
	if err != nil {
		log.Fatalf("Erro ao buscar dados: %v", err)
	}

	supports, resistances := calculateMultipleSupportResistance(prices)
	fmt.Printf("Suportes identificados: %v\n", supports)
	fmt.Printf("Resistências identificadas: %v\n", resistances)

	plotSupportResistance(prices, supports, resistances, days)
	fmt.Println("Análise salva em 'support_resistance.png'")
}
