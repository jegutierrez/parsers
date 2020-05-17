package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrorInvalidJSON failed parse json.
var ErrorInvalidJSON = errors.New("invalid json given")

func main() {
	jsonObject, err := ParseJSON(`{"value":1,"bool":true,"array":[1.2,5,7]}`)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%#v\n", jsonObject)
	fmt.Println("--------------------------")
	jsonArray, err := ParseJSON(`[{"value":1.2,"bool":true,"array":[1,2,5,7]},{"value":1.2,"bool":true,"array":[1,2,5,7]}]`)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("%#v\n", jsonArray)
}

// ParseJSON from a string return a map with the keys and values.
func ParseJSON(json string) (interface{}, error) {
	// clean white spaces
	json = strings.ReplaceAll(json, " ", "")
	json = strings.ReplaceAll(json, "\n", "")
	json = strings.ReplaceAll(json, "\t", "")
	json = strings.ReplaceAll(json, "\r", "")

	l, err := lex([]byte(json))
	if err != nil {
		return nil, err
	}
	return parse(l)
}

const (
	quoteChar          = `"`
	openBraceChar      = `{`
	closingBraceChar   = `}`
	openBracketChar    = `[`
	closingBracketChar = `]`
	commaChar          = `,`
	colonChar          = `:`
	trueChars          = `true`
	falseChars         = `true`
	nullChars          = `null`
	dotChar            = `.`
	negativeChar       = `-`
)

// lex return a list of tokens to be parsed.
func lex(json []byte) ([]interface{}, error) {
	var tokens []interface{}

	for i := 0; i < len(json); i++ {
		ch := string(json[i])
		if ch == openBraceChar ||
			ch == closingBraceChar ||
			ch == openBracketChar ||
			ch == closingBracketChar ||
			ch == commaChar ||
			ch == colonChar {
			tokens = append(tokens, string(ch))
			continue
		}

		if ch == quoteChar {
			j := i + 1
			for j < len(json) && string(json[j]) != quoteChar {
				j++
			}
			if j == len(json) {
				return nil, ErrorInvalidJSON
			}
			tokens = append(tokens, string(json[i+1:j]))
			i = j
			continue
		}

		j, err := lexNumber(json, i)
		if err == nil {
			s := string(json[i:j])
			if strings.Contains(s, ".") {
				f, _ := strconv.ParseFloat(s, 64)
				tokens = append(tokens, f)
			} else {
				n, _ := strconv.ParseInt(s, 10, 64)
				tokens = append(tokens, n)
			}
			i = j - 1
			continue
		}

		j, err = lexBoolean(json, i)
		if err == nil {
			s := string(json[i:j])
			if s == trueChars {
				tokens = append(tokens, true)
			} else {
				tokens = append(tokens, false)
			}
			i = j - 1
			continue
		}

		j, err = lexNull(json, i)
		if err == nil {
			tokens = append(tokens, nil)
			i = j - 1
			continue
		}
		return tokens, ErrorInvalidJSON
	}

	return tokens, nil
}

func lexNumber(json []byte, from int) (int, error) {
	i := from
	for i < len(json) && ((json[i] >= '0' && json[i] <= '9') || string(json[i]) == dotChar) {
		i++
	}
	if i == len(json) || (string(json[i]) != commaChar &&
		string(json[i]) != closingBracketChar && string(json[i]) != closingBraceChar) {
		return 0, ErrorInvalidJSON
	}
	return i, nil
}

func lexBoolean(json []byte, from int) (int, error) {
	if from+4 <= len(json) && string(json[from:from+4]) == trueChars {
		return from + 4, nil
	}
	if from+5 <= len(json) && string(json[from:from+5]) == falseChars {
		return from + 4, nil
	}
	return 0, ErrorInvalidJSON
}

func lexNull(json []byte, from int) (int, error) {
	if from+4 <= len(json) && string(json[from:from+4]) == nullChars {
		return from + 4, nil
	}
	return 0, ErrorInvalidJSON
}

// parse transforms the list of tokens into a valid json.
func parse(tokens []interface{}) (interface{}, error) {
	var response interface{}
	tk, ok := tokens[0].(string)
	if !ok {
		return nil, ErrorInvalidJSON
	}
	if openBraceChar == tk {
		_, object, err := parseObject(tokens, 0)
		if err != nil {
			return nil, err
		}
		response = object
	} else {
		_, array, err := parseArray(tokens, 0)
		if err != nil {
			return nil, err
		}
		response = array
	}
	return response, nil
}

func parseObject(tokens []interface{}, from int) (int, map[string]interface{}, error) {
	response := make(map[string]interface{})
	openBrace, ok := tokens[from].(string)
	if !ok || openBrace != openBraceChar {
		return 0, nil, ErrorInvalidJSON
	}
	i := from + 1
	for i < len(tokens) {
		key, ok := tokens[i].(string)
		if !ok {
			return 0, nil, ErrorInvalidJSON
		}
		i++
		colon, ok := tokens[i].(string)
		if !ok || colon != colonChar {
			return 0, nil, ErrorInvalidJSON
		}
		var value interface{}
		var err error
		i, value, err = getValue(tokens, i+1)
		if err != nil {
			return 0, nil, ErrorInvalidJSON
		}
		closingToken, ok := tokens[i].(string)
		if !ok || (closingToken != commaChar && closingToken != closingBraceChar) {
			return 0, nil, ErrorInvalidJSON
		}
		response[key] = value
		i++
		if closingToken == closingBraceChar {
			break
		}
	}
	return i, response, nil
}

func parseArray(tokens []interface{}, from int) (int, []interface{}, error) {
	array := make([]interface{}, 0)
	openBrace, ok := tokens[from].(string)
	if !ok || openBrace != openBracketChar {
		return 0, nil, ErrorInvalidJSON
	}
	i := from + 1
	for i < len(tokens) {
		var value interface{}
		var err error
		i, value, err = getValue(tokens, i)
		if err != nil {
			return 0, nil, ErrorInvalidJSON
		}
		closingToken, ok := tokens[i].(string)
		if !ok || (closingToken != commaChar && closingToken != closingBracketChar) {
			return 0, nil, ErrorInvalidJSON
		}
		array = append(array, value)
		i++
		if closingToken == closingBracketChar {
			break
		}
	}
	return i, array, nil
}

func getValue(tokens []interface{}, from int) (int, interface{}, error) {
	var value interface{}
	i := from
	switch valToken := tokens[i].(type) {
	case bool:
		value = tokens[i]
		return i + 1, value, nil
	case int64:
		value = tokens[i]
		return i + 1, value, nil
	case float64:
		value = tokens[i]
		return i + 1, value, nil
	case string:
		if valToken == openBraceChar {
			endIdx, object, err := parseObject(tokens, i)
			if err != nil {
				return i, nil, err
			}
			return endIdx, object, nil
		} else if valToken == openBracketChar {
			endIdx, array, err := parseArray(tokens, i)
			if err != nil {
				return i, nil, err
			}
			return endIdx, array, nil
		} else {
			value = tokens[i]
			return i + 1, value, nil
		}
	default:
		return i, nil, ErrorInvalidJSON
	}
}
