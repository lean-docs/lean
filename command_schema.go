package lean

import (
	"reflect"
	"strings"
)

func DOCXCommandSchema(id string) map[string]any {
	fields := map[string][]string{
		"text.format.suggest": {"start", "end", "run"},
		"text.suggest":        {"start", "end", "text"}, "text.replace": {"start", "end", "text", "run"}, "text.format": {"start", "end", "run"}, "text.link": {"start", "end", "hyperlink"},
		"change.accept": {"change_id"}, "change.reject": {"change_id"}, "bookmark.set": {"bookmark"}, "bookmark.remove": {"text"}, "footnote.set": {"footnote"}, "footnote.remove": {"at"},
		"style.set": {"style_definition"}, "style.remove": {"style"}, "paragraph.format": {"paragraph"}, "paragraph.style": {"style"}, "paragraph.insert": {"at", "text", "style"}, "paragraph.split": {"at"},
		"list.set": {"numbering", "level", "restart"}, "table.insert.suggest": {"at", "rows", "columns", "style"}, "table.row.insert.suggest": {"at"}, "table.row.delete.suggest": {"at"}, "table.style": {"style"}, "table.insert": {"at", "rows", "columns", "style"}, "table.row.insert": {"at"}, "table.row.delete": {"at"}, "table.column.insert": {"at"}, "table.column.delete": {"at"},
		"image.insert": {"at", "image"}, "shape.insert": {"at", "shape"}, "drawing.layout": {"drawing"}, "page.layout": {"page"}, "page.columns": {"columns_layout"}, "section.insert": {"at", "page"}, "section.delete": {}, "section.break": {"section_break"},
		"watermark.set": {"watermark"}, "header.options": {"even_and_odd_headers", "different_first_page"}, "header.set": {"header_footer"}, "footer.set": {"header_footer"},
	}
	allowed := map[string]bool{"id": true, "section": true, "part": true, "variant": true, "block": true}
	for _, name := range fields[id] {
		allowed[name] = true
	}
	properties := map[string]any{}
	value := reflect.TypeOf(DocumentCommand{})
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if allowed[name] {
			properties[name] = commandTypeSchema(field.Type, map[reflect.Type]bool{})
		}
	}
	schema := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
	properties["id"] = map[string]any{"type": "string", "enum": []string{id}}
	schema["required"] = []string{"id"}
	return schema
}

func commandTypeSchema(value reflect.Type, seen map[reflect.Type]bool) map[string]any {
	for value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Slice, reflect.Array:
		if value.Elem().Kind() == reflect.Uint8 {
			return map[string]any{"type": "string", "contentEncoding": "base64"}
		}
		return map[string]any{"type": "array", "items": commandTypeSchema(value.Elem(), seen)}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": commandTypeSchema(value.Elem(), seen)}
	case reflect.Struct:
		if seen[value] {
			return map[string]any{"type": "object"}
		}
		seen[value] = true
		defer delete(seen, value)
		fields := map[string]any{}
		for index := 0; index < value.NumField(); index++ {
			field := value.Field(index)
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if field.PkgPath != "" || name == "-" {
				continue
			}
			if name == "" {
				name = field.Name
			}
			fields[name] = commandTypeSchema(field.Type, seen)
		}
		return map[string]any{"type": "object", "properties": fields, "additionalProperties": false}
	default:
		return map[string]any{"type": "object"}
	}
}
