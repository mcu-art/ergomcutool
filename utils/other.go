package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// StructToMap converts any struct into map[string]any
// by marshalling it to json and reverse.
func StructToMap(in any) (map[string]any, error) {
	var inInterface map[string]interface{}
	dataBytes, err := json.Marshal(in)
	if err != nil {
		return inInterface,
			fmt.Errorf("ergomcutool.utils.StructToMap: failed to marshal input into json: %v", err)
	}
	_ = json.Unmarshal(dataBytes, &inInterface)
	return inInterface, nil
}

func GetUserConfirmationViaConsole(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "Y" || text == "y" {
		return true
	}
	return false
}
