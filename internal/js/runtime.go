package js

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/vyquocvu/litebrowser/internal/dom"
)

// Runtime wraps the Goja JavaScript runtime
type Runtime struct {
	vm         *goja.Runtime
	parser     *dom.Parser
	htmlCache  string
}

// NewRuntime creates a new JavaScript runtime with console.log and document.getElementById
func NewRuntime() *Runtime {
	vm := goja.New()
	parser := dom.NewParser()
	
	runtime := &Runtime{
		vm:     vm,
		parser: parser,
	}

	// Setup console.log
	console := vm.NewObject()
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}
		fmt.Println(args...)
		return goja.Undefined()
	})
	vm.Set("console", console)

	// Setup document.getElementById
	document := vm.NewObject()
	document.Set("getElementById", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Null()
		}
		id := call.Arguments[0].String()
		
		if runtime.htmlCache == "" {
			return goja.Null()
		}
		
		text, err := runtime.parser.GetElementByID(runtime.htmlCache, id)
		if err != nil || text == "" {
			return goja.Null()
		}
		
		// Return a simple object with textContent
		obj := vm.NewObject()
		obj.Set("textContent", text)
		return obj
	})
	vm.Set("document", document)

	return runtime
}

// SetHTMLContent sets the HTML content for document operations
func (r *Runtime) SetHTMLContent(html string) {
	r.htmlCache = html
}

// RunScript executes JavaScript code
func (r *Runtime) RunScript(script string) (goja.Value, error) {
	return r.vm.RunString(script)
}
