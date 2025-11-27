package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	gomerrors "github.com/biyonik/gomad/internal/errors"
)

/*
========================================================================================================================
 GOMAD - Bridge Registry Sistemi
========================================================================================================================
Bu dosya, Go ile JavaScript arasında köprü kurmak için oluşturulmuş güçlü bir fonksiyon bağlama (binding) yapısı içerir.
Amaç şudur:

    → Go fonksiyonlarını runtime’da yansıma (reflect) ile çözümlemek
    → Bu fonksiyonları JavaScript tarafından çağrılabilir hâle getirmek
    → Parametreleri JSON üzerinden alıp işlemek
    → Geri dönüş değerlerini güvenli şekilde parse etmek
    → Tüm sistemi thread-safe ve modüler hâlde tutmak

Bu sistem ile tarayıcı / JS köprüsü kurulur; Javascript tarafından "Call" ile gönderilen method adı ve parametreler,
bu kayıt defteri (Registry) içinde çözülür ve gerçek Go fonksiyonuna aktarılır. Böylece iki dünya birbirine bağlanmış olur.

Yani **Go'da yazılmış fonksiyonları JS'ye açan kapı tam olarak burasıdır.**
Laravel’deki "Service Container", Symfony’deki "Event Dispatcher" ruhunu taşır — ama Go’ya özgü, hafif, yalın, hızlı.

------------------------------------------------------------------------------------------------------------------------
@author      Ahmet ALTUN
@email       ahmet.altun60@gmail.com
@github      github.com/biyonik
@linkedin    linkedin.com/in/biyonik
========================================================================================================================
*/

// ======================================================================================================================
//  BoundFunc — JS tarafından çağrılmak üzere bağlanmış (expose edilmiş) Go fonksiyonu temsil eder.
//  Fonksiyon reflect.Value olarak saklanır ve çalıştırma zamanı parametre parse, çağrı ve dönüş işlemleri yapılır.
// ======================================================================================================================

// BoundFunc represents a bound Go function that can be called from JavaScript.
// Bu struct, reflect kullanarak fonksiyonu dinamik olarak çağırır.
type BoundFunc struct {
	// Name is the function name as registered.
	Name string

	// Fn is the actual Go function.
	Fn reflect.Value

	// Type is the function's reflect.Type.
	Type reflect.Type

	// NumIn is the number of input parameters.
	NumIn int

	// NumOut is the number of output values.
	NumOut int

	// HasError indicates if the last return value is an error.
	HasError bool
}

// ======================================================================================================================
//  Registry — Fonksiyon kayıt defteri
//  JavaScript → Go çağrılarının merkezi.
//  * Thread-safe çalışır
//  * Fonksiyon kayıt, silme, listeleme, çağırma mekanizması sağlar
// ======================================================================================================================

// Registry manages bound functions that can be called from JavaScript.
//
// Thread-safe: Tüm metodlar concurrent kullanım için güvenli.
//
// Örnek kullanım:
//
//	r := NewRegistry()
//	r.Register("greet", func(name string) string {
//	    return "Hello, " + name
//	})
//
//	result, err := r.Call("greet", []interface{}{"Ahmet"})
//	// result = "Hello, Ahmet!"
type Registry struct {
	funcs map[string]*BoundFunc
	mu    sync.RWMutex
}

// NewRegistry creates a new function registry.
// Amaç: Fonksiyonların JS tarafından çağrılabilmesi için merkezi bir kayıt oluşturmak.
func NewRegistry() *Registry {
	return &Registry{
		funcs: make(map[string]*BoundFunc),
	}
}

// Register binds a Go function with the given name.
// Bu metod, fonksiyonu Registry içerisine ekler ve JS tarafından çağrılabilir hâle getirir.
//
// Desteklenen fonksiyon tipleri:
//   - func()
//   - func() error
//   - func() T
//   - func() (T, error)
//   - func(args...) (T, error)
//
// T: JSON serileştirilebilir her tür olabilir.
//
// Validasyonlar:
//
//	✔ İsim boş olamaz
//	✔ Fonksiyon nil olamaz
//	✔ Aynı isimle iki defa kayıt yapılamaz
//	✔ En fazla bir adet error dönüşü olabilir
func (r *Registry) Register(name string, fn interface{}) error {
	// Validasyonlar
	if name == "" {
		return gomerrors.NewBindingError(name, "name cannot be empty", nil)
	}

	if fn == nil {
		return gomerrors.NewBindingError(name, "function cannot be nil", nil)
	}

	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	if fnType.Kind() != reflect.Func {
		return gomerrors.NewBindingError(name, "not a function", nil)
	}

	// Duplicate check
	r.mu.RLock()
	_, exists := r.funcs[name]
	r.mu.RUnlock()

	if exists {
		return gomerrors.NewBindingError(name, "already registered", gomerrors.ErrAlreadyExists)
	}

	// Fonksiyon analiz & return değer kontrolü
	numOut := fnType.NumOut()
	hasError := false

	if numOut > 0 {
		lastOut := fnType.Out(numOut - 1)
		hasError = lastOut.Implements(reflect.TypeOf((*error)(nil)).Elem())
	}

	if numOut > 2 {
		return gomerrors.NewBindingError(name, "too many return values (max 2)", nil)
	}

	if numOut == 2 && !hasError {
		return gomerrors.NewBindingError(name, "second return value must be error", nil)
	}

	bound := &BoundFunc{
		Name:     name,
		Fn:       fnVal,
		Type:     fnType,
		NumIn:    fnType.NumIn(),
		NumOut:   numOut,
		HasError: hasError,
	}

	r.mu.Lock()
	r.funcs[name] = bound
	r.mu.Unlock()

	return nil
}

// Unregister removes a bound function.
// Amaç: Daha önce JS'ye açılmış bir metodu sistemden kaldırmak.
func (r *Registry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.funcs[name]
	if exists {
		delete(r.funcs, name)
	}
	return exists
}

// Has checks if a function is registered.
// Kayıtlı mı değil mi anlamak için basit kontrol.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.funcs[name]
	return exists
}

// List returns all registered function names.
// Debug, inspection veya UI tarafında görüntüleme için kullanılabilir.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.funcs))
	for name := range r.funcs {
		names = append(names, name)
	}
	return names
}

// Call invokes a registered function with the given arguments.
// Message içindeki JSON argümanları çözer, fonksiyona uygun tipe çevirir ve çalıştırır.
//
// Başarılı dönüş → result, nil
// Hatalı dönüş   → nil, error
func (r *Registry) Call(name string, argsJSON json.RawMessage) (interface{}, error) {
	r.mu.RLock()
	bound, exists := r.funcs[name]
	r.mu.RUnlock()

	if !exists {
		return nil, gomerrors.NewBindingError(name, "not found", gomerrors.ErrNotFound)
	}

	// Argüman çözme
	var rawArgs []json.RawMessage
	if argsJSON != nil && len(argsJSON) > 0 {
		if err := json.Unmarshal(argsJSON, &rawArgs); err != nil {
			return nil, gomerrors.NewBindingError(name, "failed to parse arguments", err)
		}
	}

	if len(rawArgs) != bound.NumIn {
		return nil, gomerrors.NewBindingError(name,
			fmt.Sprintf("expected %d arguments, got %d", bound.NumIn, len(rawArgs)),
			gomerrors.ErrInvalidArgument)
	}

	args := make([]reflect.Value, bound.NumIn)
	for i := 0; i < bound.NumIn; i++ {
		argType := bound.Type.In(i)
		argPtr := reflect.New(argType)

		if err := json.Unmarshal(rawArgs[i], argPtr.Interface()); err != nil {
			return nil, gomerrors.NewBindingError(name,
				fmt.Sprintf("failed to convert argument %d to %s", i, argType.String()),
				err)
		}

		args[i] = argPtr.Elem()
	}

	results := bound.Fn.Call(args)

	return processResults(bound, results)
}

// processResults converts reflect.Value results to interface{} and error.
// Fonksiyon dönüş tiplerini çözerek JS'ye uygun hâle getirir.
func processResults(bound *BoundFunc, results []reflect.Value) (interface{}, error) {
	switch bound.NumOut {
	case 0:
		return nil, nil

	case 1:
		if bound.HasError {
			if !results[0].IsNil() {
				return nil, results[0].Interface().(error)
			}
			return nil, nil
		}
		return results[0].Interface(), nil

	case 2:
		var err error
		if !results[1].IsNil() {
			err = results[1].Interface().(error)
		}
		if err != nil {
			return nil, err
		}
		return results[0].Interface(), nil

	default:
		return nil, fmt.Errorf("unexpected number of return values: %d", bound.NumOut)
	}
}

// CallWithMessage is a convenience method that handles a full Message.
// Call gibi çalışır fakat parametreyi Message alır ve Message döner.
// Yani JS <-> Go mesaj protokolünün tam döngü wrapper'ıdır.
func (r *Registry) CallWithMessage(msg *Message) *Message {
	if msg.Type != MessageTypeCall {
		return NewErrorMessage(msg.ID, ErrCodeUnknown, "expected call message", "")
	}

	result, err := r.Call(msg.Method, msg.Args)
	if err != nil {
		code := ErrCodeExecution
		if errors.Is(err, gomerrors.ErrNotFound) {
			code = ErrCodeMethodNotFound
		} else if errors.Is(err, gomerrors.ErrInvalidArgument) {
			code = ErrCodeInvalidArgs
		}
		return NewErrorMessage(msg.ID, code, err.Error(), "")
	}

	resultMsg, err := NewResultMessage(msg.ID, result)
	if err != nil {
		return NewErrorMessage(msg.ID, ErrCodeExecution, "failed to serialize result", err.Error())
	}

	return resultMsg
}
