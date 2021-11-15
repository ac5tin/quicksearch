package textprocessor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func (tp *TextProcessor) LangDetect(text string, lang *string) error {
	r := []string{text}
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/langdet", os.Getenv("TEXTPROCESSOR_ENDPOINT")), bytes.NewBuffer(b))
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

	res := new([]string)
	if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("textprocessor failed to detect language")
	}

	*lang = (*res)[0]
	return nil
}
