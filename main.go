package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func Calc(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")

	getPrecedence := func(operator rune) int {
		if operator == '+' || operator == '-' {
			return 1
		}
		if operator == '*' || operator == '/' {
			return 2
		}
		return 0
	}

	performOperation := func(operand1, operand2 float64, operator rune) (float64, error) {
		switch operator {
		case '+':
			return operand1 + operand2, nil
		case '-':
			return operand1 - operand2, nil
		case '*':
			return operand1 * operand2, nil
		case '/':
			if operand2 == 0 {
				return 0, errors.New("Деление на ноль!")
			}
			return operand1 / operand2, nil
		default:
			return 0, errors.New("Неизвестная операция!")
		}
	}

	executeOperations := func(operands []float64, operators []rune) ([]float64, []rune, error) {
		if len(operands) < 2 || len(operators) == 0 {
			return operands, operators, errors.New("Неверное выражение!")
		}

		operand2 := operands[len(operands)-1]
		operand1 := operands[len(operands)-2]
		operator := operators[len(operators)-1]
		operands = operands[:len(operands)-2]
		operators = operators[:len(operators)-1]

		result, err := performOperation(operand1, operand2, operator)
		if err != nil {
			return operands, operators, err
		}
		operands = append(operands, result)
		return operands, operators, nil
	}

	var operands []float64
	var operators []rune

	i := 0

	for i < len(expression) {
		char := expression[i]
		if (char >= '0' && char <= '9') || char == '.' {
			j := i
			for j < len(expression) && ((expression[j] >= '0' && expression[j] <= '9') || expression[j] == '.') {
				j++
			}
			num, err := strconv.ParseFloat(expression[i:j], 64)
			if err != nil {
				return 0, err
			}
			operands = append(operands, num)
			i = j
		} else if char == '(' {
			operators = append(operators, rune(char))
			i++
		} else if char == ')' {
			for len(operators) > 0 && operators[len(operators)-1] != '(' {
				var err error
				operands, operators, err = executeOperations(operands, operators)
				if err != nil {
					return 0, err
				}
			}
			if len(operators) == 0 {
				return 0, errors.New("Ошибка со скобками!")
			}
			operators = operators[:len(operators)-1]
			i++
		} else {
			for len(operators) > 0 && getPrecedence(operators[len(operators)-1]) >= getPrecedence(rune(char)) {
				var err error
				operands, operators, err = executeOperations(operands, operators)
				if err != nil {
					return 0, err
				}
			}
			operators = append(operators, rune(char))
			i++
		}
	}

	for len(operators) > 0 {
		var err error
		operands, operators, err = executeOperations(operands, operators)
		if err != nil {
			return 0, err
		}
	}

	if len(operands) != 1 {
		return 0, errors.New("Выражение неверно!")
	}
	return operands[0], nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Метод запрещен!", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Expression string `json:"expression"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Expression == "" {
		http.Error(w, "Выражение неверно!", http.StatusUnprocessableEntity)
		return
	}

	result, err := Calc(request.Expression)
	if err != nil {
		if err.Error() == "Неверное выражение!" || err.Error() == "Ошибка со скобками!" {
			http.Error(w, "Выражение неверно!", http.StatusUnprocessableEntity)
		} else {
			http.Error(w, "Ошибка в работе сервера!", http.StatusInternalServerError)
		}
		return
	}

	response := map[string]string{"result": fmt.Sprintf("%f", result)}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/api/v1/calculate", handler)
	fmt.Println("Сервер запущен на http://localhost:8080/api/calculate")
	http.ListenAndServe(":8080", nil)
}
