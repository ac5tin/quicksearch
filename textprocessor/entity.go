package textprocessor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func (tp *TextProcessor) EntityRecognition(inp InputText, entities *[]string) error {
	input := new([]InputText)
	*input = append(*input, inp)

	b, err := json.Marshal(input)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/entity", os.Getenv("TEXTPROCESSOR_ENDPOINT")), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	res := new([][]string)
	if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("textprocessor failed to recognise entity")
	}

	*entities = (*res)[0]
	return nil
}
