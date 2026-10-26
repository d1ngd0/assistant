package ui

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

//go:embed all:tmpl
var Tmpls embed.FS

//go:embed all:static
var Static embed.FS

func loadTemplates() (*template.Template, error) {
	var tmpl = template.New("")

	tmpl.Funcs(template.FuncMap{
		"md2html": func(md string) template.HTML {
			// create markdown parser with extensions
			extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
			prs := parser.NewWithExtensions(extensions)
			doc := prs.Parse([]byte(md))

			// create HTML renderer with extensions
			htmlFlags := html.CommonFlags | html.HrefTargetBlank
			opts := html.RendererOptions{Flags: htmlFlags}
			renderer := html.NewRenderer(opts)

			return template.HTML(string(markdown.Render(doc, renderer)))
		},
	})

	err := fs.WalkDir(Tmpls, "tmpl", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		f, err := Tmpls.Open(path)
		if err != nil {
			return err
		}

		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		name := strings.TrimPrefix(path, "tmpl/")
		_, err = tmpl.New(name).Parse(string(b))
		return err
	})

	return tmpl, err
}
