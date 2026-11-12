package catalog

import "embed"

//go:embed icons/*.svg documents/*
var assets embed.FS

func Icon(extension string) []byte {
	data, err := assets.ReadFile("icons/" + extension + ".svg")
	if err != nil {
		data, _ = assets.ReadFile("icons/txt.svg")
	}
	return data
}

func Document(extension string) ([]byte, error) {
	return assets.ReadFile("documents/blank." + extension)
}

func IsDocument(extension string) bool {
	return extension == "docx" || extension == "xlsx" || extension == "pptx" || extension == "rtf"
}
