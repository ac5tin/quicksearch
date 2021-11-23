package tokeniser

type model struct {
	ModelFileName string `json:"model_file_name"`
	ModelFileHash string `json:"model_file_hash"`
	S3Link        string `json:"s3_link"`
}

type tokeniseReq struct {
	model
	ArrayText []string `json:"array_text"`
}

func (t *Tokeniser) Tokenise(text []string, out *[]string) error {
	return nil
}
