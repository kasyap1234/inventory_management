package validation

import (
	"html"
	"reflect"
	"strings"
)

// SanitizeStruct trims and cleans string fields based on `sanitize` struct tags.
// Supported directives: trim (default), lower, html.
func SanitizeStruct(target interface{}) {
	if target == nil {
		return
	}

	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return
	}

	sanitizeValue(v.Elem())
}

func sanitizeValue(v reflect.Value) {
	if !v.IsValid() {
		return
	}

	switch v.Kind() {
	case reflect.Struct:
		typeOf := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if !field.CanSet() {
				continue
			}

			tag := typeOf.Field(i).Tag.Get("sanitize")
			if tag == "-" {
				continue
			}

			sanitizeField(field, tag)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			value := v.MapIndex(key)
			sanitizeValue(value)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.CanSet() {
				sanitizeValue(elem)
			}
		}
	case reflect.Ptr:
		if !v.IsNil() {
			sanitizeValue(v.Elem())
		}
	}
}

func sanitizeField(field reflect.Value, tag string) {
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return
		}
		field = field.Elem()
	}

	switch field.Kind() {
	case reflect.Struct, reflect.Slice, reflect.Array, reflect.Map:
		sanitizeValue(field)
	case reflect.String:
		field.SetString(applyDirectives(field.String(), tag))
	}
}

func applyDirectives(value, tag string) string {
	directives := parseDirectives(tag)
	clean := value

	if directives.trim {
		clean = strings.TrimSpace(clean)
	}

	if directives.lower {
		clean = strings.ToLower(clean)
	}

	if directives.html {
		clean = SanitizeHTMLElement(clean)
	}

	return clean
}

// SanitizeHTMLElement escapes HTML characters to prevent XSS attacks
func SanitizeHTMLElement(input string) string {
	return html.EscapeString(input)
}

type directiveSet struct {
	trim  bool
	lower bool
	html  bool
}

func parseDirectives(tag string) directiveSet {
	if tag == "" {
		return directiveSet{trim: true}
	}

	parts := strings.Split(tag, ",")
	set := directiveSet{}

	for _, part := range parts {
		switch strings.TrimSpace(part) {
		case "trim", "":
			set.trim = true
		case "lower":
			set.lower = true
		case "html":
			set.html = true
		}
	}

	if !set.trim && !set.lower && !set.html {
		set.trim = true
	}

	return set
}
