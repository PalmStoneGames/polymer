package polymer

import (
	"github.com/gopherjs/gopherjs/js"
	"reflect"
)

var typeOfPtrBindProto = reflect.TypeOf(&BindProto{})

type BindInterface interface {
	Notify(path string, val interface{})
	data() *BindProto
}

type BindProto struct {
	this *js.Object
}

func (p *BindProto) data() *BindProto { return p }
func (p *BindProto) Notify(path string, val interface{}) {
	lookupProto(p.data().this).Notify("Model."+path, val)
}

type AutoBindGoTemplate struct {
	*WrappedElement
}

func (el *AutoBindGoTemplate) Bind(model BindInterface) {
	refType := reflect.TypeOf(model)
	if refType.Kind() != reflect.Ptr || refType.Elem().Kind() != reflect.Struct {
		panic("BindInterface should be a pointer to a struct")
	}

	refVal := reflect.ValueOf(model).Elem()

	bindProtoField, found := refType.Elem().FieldByName("BindProto")
	if !found || !bindProtoField.Anonymous || bindProtoField.Type != typeOfPtrBindProto {
		panic("BindInterface should have an anonymous field polymer.BindProto")
	}

	// Set the BindProto
	refVal.FieldByIndex(bindProtoField.Index).Set(reflect.New(typeOfPtrBindProto.Elem()))

	jsObj := unwrap(el.Underlying())
	proto := lookupProto(jsObj).(*autoBindTemplate)

	if proto.Model != nil {
		panic("Model may only be bound once")
	}

	if model.data().this != nil {
		panic("model is already bound to another template")
	}

	// Setup handlers
	for _, handler := range parseHandlers(refType) {
		jsObj.Set(handler.Name, eventHandlerCallback(handler.Func))
	}

	// Setup compute functions
	for _, handler := range parseComputes(refType) {
		jsObj.Set(handler.Name, computeCallback(handler.Func))
	}

	// Setup channel based event handlers
	for _, handler := range parseChanHandlers(refType) {
		// Create channel
		chanVal := refVal.FieldByIndex(handler.Index)
		chanVal.Set(reflect.MakeChan(chanVal.Type(), 0))

		// Set handler function
		jsObj.Set(handler.Name, eventChanCallback(chanVal))
	}

	// Set the needed data on Go side
	proto.Model = model
	model.data().this = jsObj
	proto.bound = true
	proto.render()

	// Notify the JS side
	proto.Notify("Model", model)
}

// autoBind is a port of the polymer auto-bind template to go, so we can bind our own observers to it
// Original source: https://github.com/Polymer/polymer/blob/master/src/lib/template/dom-bind.html
type autoBindTemplate struct {
	*Proto

	importsReady bool
	bound        bool
	readied      bool
	children     *js.Object // js Array of child nodes

	Model interface{} `polymer:"bind"`
}

func (*autoBindTemplate) TagName() string {
	return "dom-bind-go"
}

func (*autoBindTemplate) Extends() string {
	return "template"
}

func (t *autoBindTemplate) Created() {
	js.Global.Get("Polymer").Get("RenderStatus").Call("whenReady", t.markImportsReady)
}

func (t *autoBindTemplate) Attached() {
	if t.importsReady {
		t.render()
	}
}

func (t *autoBindTemplate) Detached() {
	t.removeChildren()
}

func (t *autoBindTemplate) markImportsReady() {
	t.importsReady = true
	t.ensureReady()
}

func (t *autoBindTemplate) ensureReady() {
	if t.bound && !t.readied {
		t.This().Call("_readySelf")
	}
}

func (t *autoBindTemplate) insertChildren() {
	t.ParentElement().InsertBefore(t.Root(), t)
}

func (t *autoBindTemplate) removeChildren() {
	if t.children.Bool() {
		root := t.Root()
		for i := 0; i < t.children.Length(); i++ {
			root.AppendChild(WrapJSElement(t.children.Index(i)))
		}
	}
}

func (t *autoBindTemplate) prepConfigure() {
	config := js.M{}
	propertyEffects := t.This().Get("_propertyEffects")
	if propertyEffects != js.Undefined {
		for i := 0; i < propertyEffects.Length(); i++ {
			propStr := propertyEffects.Index(i).String()
			config[propStr] = propStr
		}
	}

	setupConfigFunc := t.This().Get("_setupConfigure")
	t.This().Set("_setupConfigure", func() { setupConfigFunc.Call("call", t.This(), config) })
}

func (t *autoBindTemplate) render() {
	if t.bound {
		t.ensureReady()
		if t.children == nil {
			this := t.This()
			this.Set("_template", this)
			this.Call("_prepAnnotations")
			this.Call("_prepEffects")
			this.Call("_prepBehaviors")
			t.prepConfigure()
			this.Call("_prepBindings")
			this.Call("_prepPropertyInfo")
			js.Global.Get("Polymer").Get("Base").Get("_initFeatures").Call("call", this)
			t.children = js.Global.Get("Polymer").Get("DomApi").Call("arrayCopyChildNodes", this.Get("root"))
		}

		t.insertChildren()
		t.Fire("dom-change", nil)
	}
}

func init() {
	Register(&autoBindTemplate{},
		CustomRegistrationAttr{"_template", nil},
		CustomRegistrationAttr{"_registerFeatures", js.MakeFunc(func(this *js.Object, args []*js.Object) interface{} {
			this.Call("_prepConstructor")
			return nil
		})},
		CustomRegistrationAttr{"_initFeatures", js.MakeFunc(func(this *js.Object, args []*js.Object) interface{} { return nil })},
		CustomRegistrationAttr{"_scopeElementClass", js.MakeFunc(func(this *js.Object, args []*js.Object) interface{} {
			element := args[0]
			selector := args[1]

			datahost := this.Get("dataHost")
			if datahost.Bool() {
				return datahost.Call("_scopeElementClass", element, selector)
			} else {
				return selector
			}
		})},
	)
}
