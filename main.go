package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	inputFile  = flag.String("in", "TEST.md", "Input Markdown file")
	outputFile = flag.String("out", "output.html", "Output HTML file")
	title      = flag.String("title", "Document", "HTML document title")
	serve      = flag.Bool("serve", false, "Serve output HTML at http://localhost:8080")
)

func wrapParagraphs(md string) string {
	lines := strings.Split(md, "\n")
	var result []string
	for _, line := range lines {
		lineTrim := strings.TrimSpace(line)
		if lineTrim == "" {
			result = append(result, "")
			continue
		}
		if strings.HasPrefix(lineTrim, "<h") ||
			strings.HasPrefix(lineTrim, "<ul") ||
			strings.HasPrefix(lineTrim, "<li") ||
			strings.HasPrefix(lineTrim, "<pre") ||
			strings.HasPrefix(lineTrim, "<img") ||
			strings.HasPrefix(lineTrim, "<p") ||
			strings.HasPrefix(lineTrim, "<code") ||
			strings.HasPrefix(lineTrim, "<strong") ||
			strings.HasPrefix(lineTrim, "<em") {
			result = append(result, line)
		} else {
			result = append(result, "<p>"+line+"</p>")
		}
	}
	return strings.Join(result, "\n")
}

func parseMarkdown(md string) string {
	// Escape HTML
	md = template.HTMLEscapeString(md)

	// Convert code blocks
	md = regexp.MustCompile(`(?s)\x60\x60\x60(.*?)\x60\x60\x60`).ReplaceAllString(md, "<pre><code>$1</code></pre>")
	// Inline code
	md = regexp.MustCompile("`([^`]*)`").ReplaceAllString(md, "<code>$1</code>")
	// Headers
	md = regexp.MustCompile(`(?m)^# (.*)$`).ReplaceAllString(md, "<h1>$1</h1>")
	md = regexp.MustCompile(`(?m)^## (.*)$`).ReplaceAllString(md, "<h2>$1</h2>")
	md = regexp.MustCompile(`(?m)^### (.*)$`).ReplaceAllString(md, "<h3>$1</h3>")
	// Bold and Italic
	md = regexp.MustCompile(`\*\*(.*?)\*\*`).ReplaceAllString(md, "<strong>$1</strong>")
	md = regexp.MustCompile(`\*(.*?)\*`).ReplaceAllString(md, "<em>$1</em>")
	// Links and images
	md = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`).ReplaceAllString(md, `<a href="$2">$1</a>`)
	md = regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`).ReplaceAllString(md, `<img src="$2" alt="$1" />`)
	// Lists
	md = regexp.MustCompile(`(?m)^- (.*)$`).ReplaceAllString(md, "<li>$1</li>")
	md = regexp.MustCompile(`(?s)((<li>.*?</li>\n*)+)`).ReplaceAllString(md, "<ul>$1</ul>")

	// Paragraphs (manual line wrap)
	md = wrapParagraphs(md)

	return md
}
func startServer(outputFile string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, outputFile)
	})
	fmt.Println("Serving at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func main() {
	flag.Parse()

	markdownContent, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	htmlBody := parseMarkdown(string(markdownContent))

	// Load template
	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		log.Fatalf("Error loading HTML template: %v", err)
	}

	// Output file
	outFile, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer outFile.Close()

	// Render HTML
	err = tmpl.Execute(outFile, map[string]interface{}{
		"Title": *title,
		"Body":  template.HTML(htmlBody),
	})
	if err != nil {
		log.Fatalf("Error rendering HTML: %v", err)
	}

	fmt.Printf("✅ Converted %s to %s\n", *inputFile, *outputFile)

	if *serve {
		startServer(*outputFile)
	}
}
