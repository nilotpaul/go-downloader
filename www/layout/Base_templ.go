// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.707
package layout

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import "github.com/nilotpaul/go-downloader/www/component"

func Base() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, templ_7745c5c3_W io.Writer) (templ_7745c5c3_Err error) {
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templ_7745c5c3_W.(*bytes.Buffer)
		if !templ_7745c5c3_IsBuffer {
			templ_7745c5c3_Buffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templ_7745c5c3_Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<!doctype html><html lang=\"en\" data-theme=\"\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><link href=\"./public/styles.css\" rel=\"stylesheet\"><title>Go Downloader</title><script>\n\t\t\t\t(function() {\n  \t\t\t\t\t  const storedTheme = localStorage.getItem(\"theme\");\n  \t\t\t\t\t  const prefersDarkScheme = window.matchMedia(\"(prefers-color-scheme: dark)\").matches;\n   \t\t\t\t\t  const theme = storedTheme || (prefersDarkScheme ? \"dark\" : \"garden\");\n \t\t\t\t\t   if (!storedTheme) {\n   \t\t\t\t\t\t   localStorage.setItem(\"theme\", theme);\n \t\t\t\t\t   }\n  \t\t\t\t\t  document.documentElement.setAttribute(\"data-theme\", theme);\n \t\t\t\t })();\n\t\t\t</script></head><body class=\"max-w-6xl mx-auto antialiased\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = component.Theme().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
