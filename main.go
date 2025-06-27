package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Serve static files (frontend)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/upload", handleUpload)

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "templates/index.html")
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20) // 10 MB
	file, handler, err := r.FormFile("pdf")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save uploaded PDF temporarily
	tmpPDF := filepath.Join(os.TempDir(), fmt.Sprintf("upload-%d-%s", time.Now().UnixNano(), handler.Filename))
	out, err := os.Create(tmpPDF)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	io.Copy(out, file)

	// Convert PDF to LaTeX (placeholder: extract text and wrap in LaTeX)
	texContent, err := pdfToLatex(tmpPDF)
	if err != nil {
		http.Error(w, "Failed to convert PDF", http.StatusInternalServerError)
		return
	}

	// Send .tex file as download
	w.Header().Set("Content-Disposition", "attachment; filename=output.tex")
	w.Header().Set("Content-Type", "application/x-tex")
	w.Write([]byte(texContent))

	// Clean up
	os.Remove(tmpPDF)
}

// pdfToLatex is a placeholder: extracts text and wraps it in a LaTeX template
func pdfToLatex(pdfPath string) (string, error) {
	// For demo: use pdftotext if available, else just return a stub
	text := ""
	cmd := "pdftotext"
	if _, err := execLookPath(cmd); err == nil {
		txtPath := pdfPath + ".txt"
		_ = os.Remove(txtPath)
		if err := runCmd(cmd, pdfPath, txtPath); err == nil {
			b, err := os.ReadFile(txtPath)
			if err == nil {
				text = string(b)
			}
			os.Remove(txtPath)
		}
	}
	if text == "" {
		text = "PDF to LaTeX conversion is not implemented. Please install 'pdftotext' for basic text extraction."
	}
	latex := `\\documentclass{article}
\\begin{document}
` + latexEscape(text) + `
\\end{document}`
	return latex, nil
}

// latexEscape escapes special LaTeX characters
func latexEscape(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\textbackslash{}",
		"%", "\\%",
		"$", "\\$",
		"#", "\\#",
		"_", "\\_",
		"{", "\\{",
		"}", "\\}",
		"&", "\\&",
		"~", "\\textasciitilde{}",
		"^", "\\textasciicircum{}",
	)
	return replacer.Replace(s)
}

// execLookPath checks if a command exists
func execLookPath(cmd string) (string, error) {
	return exec.LookPath(cmd)
}

// runCmd runs a command with args
func runCmd(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	return cmd.Run()
}
