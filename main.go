package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
)

// Request structure
type CalculationRequest struct {
	Num1     float64 `json:"num1"`
	Num2     float64 `json:"num2"`
	Operator string  `json:"operator"`
}

// Response structure
type CalculationResponse struct {
	Result float64 `json:"result"`
	Error  string  `json:"error,omitempty"` // `omitempty` ensures that if there is no error, it does not appear in the JSON.
}

// Health Resposne
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// cors config
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// allow any origin, not suitable for production or real projects
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// allowed methods
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		// Content-Type
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func (r *CalculationRequest) Validate() error {
	// define allowed operatos
	validOperators := map[string]bool{
		"+": true, "-": true, "*": true, "/": true,
		"^": true, "sqrt": true, "%": true,
	}

	if r.Operator == "" {
		return fmt.Errorf("please enter a valid operator")
	}

	if !validOperators[r.Operator] {
		return fmt.Errorf("operator not supported: %s", r.Operator)
	}

	// Divide by zero
	if r.Operator == "/" && r.Num2 == 0 {
		return fmt.Errorf("cannot divide by zero")
	}

	// square root of negative numbers
	if r.Operator == "sqrt" && r.Num1 < 0 {
		return fmt.Errorf("cannot calculate square root of negative numbers")
	}

	return nil
}

func performCalculation(req CalculationRequest) float64 {
	switch req.Operator {
	case "+":
		return req.Num1 + req.Num2
	case "-":
		return req.Num1 - req.Num2
	case "*":
		return req.Num1 * req.Num2
	case "/":
		return req.Num1 / req.Num2
	case "^":
		return math.Pow(req.Num1, req.Num2)
	case "sqrt":
		return math.Sqrt(req.Num1)
	case "%":
		return (req.Num1 / 100) * req.Num2
	default:
		return 0
	}
}

// Handler
func calculateHandler(w http.ResponseWriter, r *http.Request) {
	var req CalculationRequest

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(CalculationResponse{Error: "Method not allowed"})
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CalculationResponse{Error: "invalid JSON"})
		return
	}

	// Validate business logic and valid operators
	if err := req.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CalculationResponse{Error: err.Error()})
		return
	}

	result := performCalculation(req)

	// Verify if the result is infinite or not a number
	if math.IsInf(result, 0) || math.IsNaN(result) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CalculationResponse{
			Error: "Result to big to be processed (Overflow)",
		})
		return
	}

	json.NewEncoder(w).Encode(CalculationResponse{Result: result})
}

func main() {
	// Endpoint de prueba
	http.HandleFunc("/health", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthResponse{Status: "OK", Message: "Calculadora Activa"})
	}))

	http.HandleFunc("/calculate", enableCORS(calculateHandler))

	log.Println("Servidor escuchando en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
