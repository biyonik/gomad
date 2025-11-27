// Package bridge, Go ile JavaScript arasında köprü kurarak GOMAD uygulamalarında tip güvenli iletişim sağlar.
// Bu dosya, Go fonksiyonlarını analiz edip otomatik TypeScript tanımları (.d.ts) üretir.
//
// TSGenerator struct'ları, döngüsel referansları ve tanımlanan structları yönetir.
// Böylece Angular veya başka frontend kodları, Go tarafındaki fonksiyonları tip güvenli olarak çağırabilir.
//
// Her fonksiyon ve struct için açıklamalar eklenmiştir.
//
// @author Ahmet ALTUN
// @github github.com/biyonik
// @linkedin linkedin.com/in/biyonik
// @email ahmet.altun60@gmail.com
package bridge

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// TSGenerator, TypeScript tanımları üretmek için state (durum) tutar.
// - definitions: Hangi struct -> Hangi interface adı
// - pending: Kuyrukta bekleyen struct tipleri
// - output: Üretilen TypeScript kodu
type TSGenerator struct {
	definitions map[reflect.Type]string
	pending     []reflect.Type
	output      *strings.Builder
}

// GenerateTypeDefinitions, Bridge içindeki Registry'den TypeScript tanımlarını üretir.
// - Fonksiyon parametrelerini ve dönüş tiplerini TS tipine çevirir.
// - Struct'ları interface olarak kaydeder ve kuyruğu işler.
// - Window.gomad global API genişletmesini ekler.
func (r *Registry) GenerateTypeDefinitions() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	gen := &TSGenerator{
		definitions: make(map[reflect.Type]string),
		pending:     make([]reflect.Type, 0),
		output:      &strings.Builder{},
	}

	// Header ekle
	gen.output.WriteString("// GOMAD Auto-Generated Definitions\n")
	gen.output.WriteString(fmt.Sprintf("// Generated at: %s\n\n", time.Now().Format(time.RFC3339)))

	// API interface oluştur
	var apiBuffer strings.Builder
	apiBuffer.WriteString("export interface GomadAPI {\n")
	apiBuffer.WriteString("    call<T = any>(method: string, ...args: any[]): Promise<T>;\n") // fallback

	for name, bound := range r.funcs {
		apiBuffer.WriteString(fmt.Sprintf("\n    // Binding: %s\n", name))
		apiBuffer.WriteString(fmt.Sprintf("    call(method: '%s'", name))

		for i := 0; i < bound.NumIn; i++ {
			argType := bound.Type.In(i)
			tsType := gen.getTSType(argType)
			apiBuffer.WriteString(fmt.Sprintf(", arg%d: %s", i, tsType))
		}

		apiBuffer.WriteString("): Promise<")
		if bound.NumOut == 0 {
			apiBuffer.WriteString("void")
		} else {
			outType := bound.Type.Out(0)
			tsType := gen.getTSType(outType)
			apiBuffer.WriteString(tsType)
		}
		apiBuffer.WriteString(">;\n")
	}
	apiBuffer.WriteString("}\n\n")

	// Kuyruktaki structları interface olarak oluştur
	gen.processPendingStructs()

	// API ve global Window ekle
	gen.output.WriteString(apiBuffer.String())
	gen.output.WriteString(`
declare global {
    interface Window {
        gomad: GomadAPI & {
            on(event: string, callback: (data: any) => void): () => void;
            off(event: string, callback?: (data: any) => void): void;
        };
    }
}`)

	return gen.output.String()
}

// processPendingStructs, pending kuyruğundaki struct tiplerini interface'e çevirir.
// - Export edilmeyen alanları veya JSON "-" tag'lı alanları atlar.
func (g *TSGenerator) processPendingStructs() {
	for len(g.pending) > 0 {
		t := g.pending[0]
		g.pending = g.pending[1:]

		if _, exists := g.definitions[t]; !exists {
			continue
		}

		interfaceName := g.definitions[t]
		g.output.WriteString(fmt.Sprintf("export interface %s {\n", interfaceName))

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" { // private alanları atla
				continue
			}

			fieldName := field.Name
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] == "-" {
					continue
				}
				if parts[0] != "" {
					fieldName = parts[0]
				}
			}

			tsType := g.getTSType(field.Type)
			g.output.WriteString(fmt.Sprintf("    %s: %s;\n", fieldName, tsType))
		}
		g.output.WriteString("}\n\n")
	}
}

// getTSType, Go tipini TypeScript tipine çevirir.
// - Struct tiplerini registerStruct ile kaydeder ve pending kuyruğuna ekler.
// - time.Time -> string olarak gider.
func (g *TSGenerator) getTSType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.PkgPath() == "time" && t.Name() == "Time" {
		return "string"
	}

	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return "string"
		}
		return g.getTSType(t.Elem()) + "[]"
	case reflect.Map:
		return fmt.Sprintf("Record<string, %s>", g.getTSType(t.Elem()))
	case reflect.Struct:
		return g.registerStruct(t)
	default:
		return "any"
	}
}

// registerStruct, struct tipini definitions map'ine ekler ve TS interface adını döner.
// - Anonymous struct için 'any' döner.
// - Yeni structları pending kuyruğuna ekler.
func (g *TSGenerator) registerStruct(t reflect.Type) string {
	if t.Name() == "" {
		return "any" // Anonymous struct desteği yok
	}

	// Zaten varsa ismini dön (Cache Hit)
	if name, ok := g.definitions[t]; ok {
		return name
	}

	// --- NAMING STRATEGY ---
	pkgPath := t.PkgPath()
	parts := strings.Split(pkgPath, "/")
	pkgName := parts[len(parts)-1] // örn: "auth" veya "main"

	var uniqueName string

	// Eğer "main" paketiyse direkt struct ismini kullan (User)
	if pkgName == "main" || pkgName == "" {
		uniqueName = t.Name()
	} else {
		// Başka paketse Prefix ekle (AuthUser, ChatMessage)
		// İlk harfi büyüt (basic title case)
		prefix := pkgName
		if len(prefix) > 0 {
			prefix = strings.ToUpper(prefix[:1]) + prefix[1:]
		}
		uniqueName = prefix + t.Name()
	}
	// -----------------------

	// Kaydet ve kuyruğa ekle
	g.definitions[t] = uniqueName
	g.pending = append(g.pending, t)

	return uniqueName
}
