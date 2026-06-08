package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (app *Config) RegisterOrganization(w http.ResponseWriter, r *http.Request) {}
func (app *Config) UploadPolicy(w http.ResponseWriter, r *http.Request)        {}

func chunkText(text string, chunkSize int) []string {
	words := strings.Fields(text)
	var chunks []string
	var current strings.Builder
	for _, word := range words {
		if current.Len()+len(word)+1 > chunkSize && current.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(word)
	}
	if current.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(current.String()))
	}
	return chunks
}

func readBody(r *http.Request) (content []byte, filename string, err error) {
	if err = r.ParseMultipartForm(10 << 20); err != nil {
		return nil, "", err
	}
	file, header, err := r.FormFile("policy")
	if err != nil {
		return nil, "", fmt.Errorf("policy file required")
	}
	defer file.Close()
	content, err = io.ReadAll(file)
	return content, header.Filename, err
}
