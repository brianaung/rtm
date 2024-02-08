// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.560
package view

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import "github.com/brianaung/rtm/internal/auth"
import "github.com/gofrs/uuid/v5"

// RoomData is used to pass room data into the html templates
type RoomDisplayData struct {
	RoomID   uuid.UUID
	RoomName string
}

// MsgData is used to pass the current message log with its metadata to the html templates
type MsgDisplayData struct {
	RoomID   uuid.UUID
	Username string
	Msg      string
	Time     string
	Mine     bool
}

func layout(user *auth.UserContext) templ.Component {
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<!doctype html><html lang=\"en\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width\"><title>rtm</title><script src=\"https://unpkg.com/htmx.org@1.9.9\" integrity=\"sha384-QFjmbokDn2DjBjq+fM+8LUIVrAgqcNW2s0PjAxHETgRn9l4fvX31ZxDxvwQnyMOX\" crossorigin=\"anonymous\"></script><script src=\"https://unpkg.com/htmx.org/dist/ext/ws.js\"></script><link href=\"/dist/output.css\" rel=\"stylesheet\"></head><body><header class=\"mx-auto container flex justify-between items-center p-4\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if user != nil {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<a class=\"font-bold font-2xl hover:underline\" href=\"/dashboard\">HOME</a> <button class=\"rounded border border-black p-1 bg-red-400\" hx-get=\"/logout\" hx-trigger=\"click\" hx-swap=\"none\">Logout</button>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</header><main class=\"mx-auto container flex-col items-center p-4\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</main></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if !templ_7745c5c3_IsBuffer {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteTo(templ_7745c5c3_W)
		}
		return templ_7745c5c3_Err
	})
}
