package textprocessor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Token struct {
	Token string  `json:"token"`
	Score float32 `json:"score"`
}

type InputText struct {
	Lang string `json:"lang"`
	Text string `json:"text"`
}

func (tp *TextProcessor) Tokenise(input InputText, tokens *[]Token) error {
	inp := new([]InputText)
	*inp = append(*inp, input)

	b, err := json.Marshal(inp)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/tokenise", os.Getenv("TEXTPROCESSOR_ENDPOINT")), bytes.NewBuffer(b))
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

	res := new([][]Token)
	if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("textprocessor failed to tokenise text")
	}

	*tokens = (*res)[0]
	return nil
}

func (tp *TextProcessor) TokeniseMulti(input *[]InputText, tokens *[][]Token) error {
	inp := *input

	b, err := json.Marshal(inp)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/tokenise", os.Getenv("TEXTPROCESSOR_ENDPOINT")), bytes.NewBuffer(b))
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

	if resp.StatusCode != 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		errMsg := fmt.Sprintf("status code not 200, Err: %s", buf.String())
		return fmt.Errorf("textprocessor failed to tokenise text: %s", errMsg)
	}

	res := new([][]Token)
	if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
		return err
	}

	*tokens = (*res)
	return nil

}
