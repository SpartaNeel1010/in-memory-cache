package main 

import (
	"fmt"
	"strconv"
	"strings"
)
func parseRESP(input string) ([]string, error) {
	
	lines := strings.Split(input, "\r\n")


	
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	if !strings.HasPrefix(lines[0], "*") {
		return nil, fmt.Errorf("expected array prefix '*', got: %s", lines[0])
	}

	elements := []string{}

	for i := 1; i < len(lines); i++ {
		// fmt.Println(lines[i])
		
		if !strings.HasPrefix(lines[i], "$") {
			continue
		}
		elementLength, err := strconv.Atoi(lines[i][1:])
		if err != nil {
			return nil, fmt.Errorf("failed to parse bulk string length: %v", err)
		}
		fmt.Println(elementLength)
		i++
		if i >= len(lines) {
			return nil, fmt.Errorf("unexpected end of input while reading string content")
		}
		if len(strings.Trim(lines[i], "\r")) != elementLength {
			return nil, fmt.Errorf("string length not matching with the given length i. $4 hello")
		}
		v:=strings.Trim(lines[i], "\r")
		fmt.Println(v)
		elements = append(elements,v )

	}



	return elements, nil
}