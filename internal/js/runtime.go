package js

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/vyquocvu/goosie/internal/dom"
)

// Runtime wraps the Goja JavaScript runtime
type Runtime struct {
	vm         *goja.Runtime
	parser     *dom.Parser
	htmlCache  string
	eventListeners map[string][]goja.Callable
}

// NewRuntime creates a new JavaScript runtime with console.log and document APIs
func NewRuntime() *Runtime {
	vm := goja.New()
	parser := dom.NewParser()
	
	runtime := &Runtime{
		vm:     vm,
		parser: parser,
		eventListeners: make(map[string][]goja.Callable),
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

	// Setup document object with all DOM APIs
	runtime.setupDocumentAPI()

	return runtime
}

// setupDocumentAPI configures all document-related APIs
func (r *Runtime) setupDocumentAPI() {
	document := r.vm.NewObject()
	
	// document.getElementById
	document.Set("getElementById", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Null()
		}
		id := call.Arguments[0].String()
		
		if r.htmlCache == "" {
			return goja.Null()
		}
		
		element, err := r.parser.GetElementByIDFull(r.htmlCache, id)
		if err != nil || element == nil {
			return goja.Null()
		}
		
		return r.createElementObject(element.ID, element.TextContent, element.TagName, element)
	})
	
	// document.getElementsByClassName
	document.Set("getElementsByClassName", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return r.vm.NewArray()
		}
		className := call.Arguments[0].String()
		
		if r.htmlCache == "" {
			return r.vm.NewArray()
		}
		
		elements, err := r.parser.GetElementsByClassName(r.htmlCache, className)
		if err != nil {
			return r.vm.NewArray()
		}
		
		return r.createElementArray(elements)
	})
	
	// document.getElementsByTagName
	document.Set("getElementsByTagName", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return r.vm.NewArray()
		}
		tagName := call.Arguments[0].String()
		
		if r.htmlCache == "" {
			return r.vm.NewArray()
		}
		
		elements, err := r.parser.GetElementsByTagName(r.htmlCache, tagName)
		if err != nil {
			return r.vm.NewArray()
		}
		
		return r.createElementArray(elements)
	})
	
	// document.querySelector
	document.Set("querySelector", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Null()
		}
		selector := call.Arguments[0].String()
		
		if r.htmlCache == "" {
			return goja.Null()
		}
		
		element, err := r.parser.QuerySelector(r.htmlCache, selector)
		if err != nil || element == nil {
			return goja.Null()
		}
		
		return r.createElementObject(element.ID, element.TextContent, element.TagName, element)
	})
	
	// document.querySelectorAll
	document.Set("querySelectorAll", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return r.vm.NewArray()
		}
		selector := call.Arguments[0].String()
		
		if r.htmlCache == "" {
			return r.vm.NewArray()
		}
		
		elements, err := r.parser.QuerySelectorAll(r.htmlCache, selector)
		if err != nil {
			return r.vm.NewArray()
		}
		
		return r.createElementArray(elements)
	})
	
	// document.createElement
	document.Set("createElement", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Null()
		}
		tagName := call.Arguments[0].String()
		
		obj := r.vm.NewObject()
		obj.Set("tagName", tagName)
		obj.Set("textContent", "")
		obj.Set("children", r.vm.NewArray())
		
		// Add manipulation methods
		r.addManipulationMethods(obj)
		
		return obj
	})
	
	r.vm.Set("document", document)
}

// createElementObject creates a JavaScript object representing a DOM element
func (r *Runtime) createElementObject(id, textContent, tagName string, elem *dom.Element) goja.Value {
	obj := r.vm.NewObject()
	obj.Set("textContent", textContent)
	
	if id != "" {
		obj.Set("id", id)
	}
	
	if tagName != "" {
		obj.Set("tagName", tagName)
	} else if elem != nil {
		obj.Set("tagName", elem.TagName)
	}
	
	if elem != nil {
		// Add classes
		if len(elem.Classes) > 0 {
			classesVal := r.vm.ToValue(elem.Classes)
			obj.Set("classList", classesVal)
		}
		
		// Add attributes
		if len(elem.Attributes) > 0 {
			attrs := r.vm.NewObject()
			for key, val := range elem.Attributes {
				attrs.Set(key, val)
			}
			obj.Set("attributes", attrs)
		}
	}
	
	// Add manipulation methods
	r.addManipulationMethods(obj)
	r.addEventMethods(obj)
	
	return obj
}

// createElementArray creates a JavaScript array of DOM elements
func (r *Runtime) createElementArray(elements []*dom.Element) goja.Value {
	elemValues := make([]interface{}, len(elements))
	for i, elem := range elements {
		elemObj := r.createElementObject(elem.ID, elem.TextContent, elem.TagName, elem)
		elemValues[i] = elemObj
	}
	return r.vm.ToValue(elemValues)
}

// addManipulationMethods adds DOM manipulation methods to an element object
func (r *Runtime) addManipulationMethods(obj *goja.Object) {
	// appendChild
	obj.Set("appendChild", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}
		child := call.Arguments[0]
		
		// Use JavaScript to manipulate the children array
		_, err := r.vm.RunString(`
			(function(parent, child) {
				if (!parent.children) {
					parent.children = [];
				}
				parent.children.push(child);
			})
		`)
		if err != nil {
			return goja.Undefined()
		}
		
		fn, _ := goja.AssertFunction(r.vm.Get("_"))
		if fn != nil {
			fn(goja.Undefined(), obj, child)
		}
		
		// Direct manipulation approach
		childrenVal := obj.Get("children")
		var childrenArr []goja.Value
		
		if childrenVal != nil && childrenVal != goja.Undefined() {
			if childrenObj := childrenVal.ToObject(r.vm); childrenObj != nil {
				length := childrenObj.Get("length").ToInteger()
				for i := int64(0); i < length; i++ {
					childrenArr = append(childrenArr, childrenObj.Get(fmt.Sprintf("%d", i)))
				}
			}
		}
		
		childrenArr = append(childrenArr, child)
		obj.Set("children", r.vm.ToValue(childrenArr))
		
		return child
	})
	
	// removeChild
	obj.Set("removeChild", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}
		childToRemove := call.Arguments[0]
		childToRemoveObj := childToRemove.ToObject(r.vm)
		
		childrenVal := obj.Get("children")
		if childrenVal == nil || childrenVal == goja.Undefined() {
			return goja.Undefined()
		}
		
		var childrenArr []goja.Value
		if childrenObj := childrenVal.ToObject(r.vm); childrenObj != nil {
			length := childrenObj.Get("length").ToInteger()
			for i := int64(0); i < length; i++ {
				child := childrenObj.Get(fmt.Sprintf("%d", i))
				childObj := child.ToObject(r.vm)
				// Compare by checking if they reference the same object
				if childObj != childToRemoveObj {
					childrenArr = append(childrenArr, child)
				}
			}
		}
		
		obj.Set("children", r.vm.ToValue(childrenArr))
		return childToRemove
	})
	
	// replaceChild
	obj.Set("replaceChild", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		newChild := call.Arguments[0]
		oldChild := call.Arguments[1]
		oldChildObj := oldChild.ToObject(r.vm)
		
		childrenVal := obj.Get("children")
		if childrenVal == nil || childrenVal == goja.Undefined() {
			return goja.Undefined()
		}
		
		var childrenArr []goja.Value
		if childrenObj := childrenVal.ToObject(r.vm); childrenObj != nil {
			length := childrenObj.Get("length").ToInteger()
			for i := int64(0); i < length; i++ {
				child := childrenObj.Get(fmt.Sprintf("%d", i))
				childObj := child.ToObject(r.vm)
				if childObj == oldChildObj {
					childrenArr = append(childrenArr, newChild)
				} else {
					childrenArr = append(childrenArr, child)
				}
			}
		}
		
		obj.Set("children", r.vm.ToValue(childrenArr))
		return oldChild
	})
	
	// insertBefore
	obj.Set("insertBefore", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		newChild := call.Arguments[0]
		refChild := call.Arguments[1]
		refChildObj := refChild.ToObject(r.vm)
		
		childrenVal := obj.Get("children")
		var childrenArr []goja.Value
		
		if childrenVal != nil && childrenVal != goja.Undefined() {
			if childrenObj := childrenVal.ToObject(r.vm); childrenObj != nil {
				length := childrenObj.Get("length").ToInteger()
				for i := int64(0); i < length; i++ {
					childrenArr = append(childrenArr, childrenObj.Get(fmt.Sprintf("%d", i)))
				}
			}
		}
		
		newChildrenArr := make([]goja.Value, 0)
		inserted := false
		
		for _, child := range childrenArr {
			childObj := child.ToObject(r.vm)
			if childObj == refChildObj && !inserted {
				newChildrenArr = append(newChildrenArr, newChild)
				inserted = true
			}
			newChildrenArr = append(newChildrenArr, child)
		}
		
		if !inserted {
			newChildrenArr = append(newChildrenArr, newChild)
		}
		
		obj.Set("children", r.vm.ToValue(newChildrenArr))
		return newChild
	})
}

// addEventMethods adds event listener methods to an element object
func (r *Runtime) addEventMethods(obj *goja.Object) {
	// addEventListener
	obj.Set("addEventListener", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		
		eventType := call.Arguments[0].String()
		callback, ok := goja.AssertFunction(call.Arguments[1])
		if !ok {
			return goja.Undefined()
		}
		
		// Store the listener
		key := fmt.Sprintf("%p:%s", obj, eventType)
		r.eventListeners[key] = append(r.eventListeners[key], callback)
		
		return goja.Undefined()
	})
	
	// removeEventListener
	obj.Set("removeEventListener", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		
		eventType := call.Arguments[0].String()
		
		// Clear all listeners for this event type (simplified implementation)
		// In a full implementation, we'd need to track function identity
		key := fmt.Sprintf("%p:%s", obj, eventType)
		r.eventListeners[key] = make([]goja.Callable, 0)
		
		return goja.Undefined()
	})
}

// SetHTMLContent sets the HTML content for document operations
func (r *Runtime) SetHTMLContent(html string) {
	r.htmlCache = html
}

// RunScript executes JavaScript code
func (r *Runtime) RunScript(script string) (goja.Value, error) {
	return r.vm.RunString(script)
}
