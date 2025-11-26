package platform

/*
=====================================================
🪟 Window Arayüzü - Platform Bağımsız Pencere Sözleşmesi
=====================================================

Bu dosya, farklı işletim sistemleri ve grafik arabirimleri üzerinde
uygulanabilecek soyut bir pencere yapısını (Window Interface) tanımlar.
Amaç, pencere oluşturma süreçlerini platformdan bağımsız hale getirmek
ve UI katmanını daha sürdürülebilir, test edilebilir ve genişletilebilir
bir mimariye oturtmaktır.

Bu arayüz, bir pencerenin yaşam döngüsünü, kullanıcı ile etkileşim kurma
şeklini ve temel görsel özelliklerinin nasıl yönetileceğini belirler.
Burada yapılan şey yalnızca fonksiyon belirlemek değil; pencerenin nasıl
açılacağı, nasıl kapatılacağı, kullanıcı hareketlerini nasıl yakaladığı,
nasıl dinlediği ve tüm bunları geliştiricinin kontrolüne nasıl sunduğu
konusunda standart oluşturmaktır.

Kısacası, bu interface bir "sözleşmedir."
Bir UI motoru bu sözleşmeyi implement ettiği anda pencere yönetimi artık
kütüphaneye değil **kurallara** bağlı olur. Bu, yazılımı ölçeklendirirken
ve farklı platformlara taşırken paha biçilemez bir esneklik sağlar.

@author Ahmet ALTUN
@github github.com/biyonik
@linkedin linkedin.com/in/biyonik
@email ahmet.altun60@gmail.com
*/

// Window, bir grafik uygulamasında gösterilebilen ve kullanıcıyla etkileşim kurabilen temel pencere yapısını tanımlar.
// Amaç platform bağımlılığını ortadan kaldırarak Windows, Linux, macOS veya başka bir UI motorunda bile ortak
// pencere davranışını korumaktır.
type Window interface {

	// Show, pencerenin görünür hale getirilmesini sağlar.
	// Bu yöntem çağrıldığında pencere ekrana yansır ve kullanım için aktif hale gelir.
	Show()

	// Close, pencerenin kapatılmasını sağlar.
	// Uygulama bu yöntem ile pencereyi kontrollü biçimde sonlandırabilir.
	Close()

	// SetTitle, pencere başlığını dinamik olarak değiştirmek için kullanılır.
	// title parametresi ile pencerenin kullanıcıya yansıyan ana metni belirlenir.
	SetTitle(title string)

	// SetSize, pencerenin genişlik ve yükseklik değerlerinin ayarlanmasını sağlar.
	// width ve height parametreleri tamamen piksel bazlıdır ve UI yerleşimi için kritiktir.
	SetSize(width, height int)

	// OnClose, pencere kapandığında tetiklenecek geri çağırma fonksiyonunu tanımlar.
	// Bu özellik pencere yaşam döngüsünü yönetmek için güçlü bir kontrol sağlar.
	OnClose(callback func())

	// OnMouseMove, fare imlecinin pencere üzerinde hareket ettiğinde tetiklenen event’i yakalar.
	// x ve y piksel koordinatları ile uygulama gerçek zamanlı etkileşim elde eder.
	OnMouseMove(callback func(x, y int))

	// OnClick, pencere üzerinde fare tıklaması gerçekleştiğinde tetiklenir.
	// x,y tıklanan konumu, button ise hangi fare tuşunun kullanıldığını belirtir.
	OnClick(callback func(x, y int, button MouseButton))

	// OnKeyDown, klavyeden bir tuşa basıldığında tetiklenen event’i yakalar.
	// keyCode parametresi ile hangi tuşa basıldığı bilgisi sağlanır.
	OnKeyDown(callback func(keyCode int))

	// OnKeyUp, klavyeden basılan tuş bırakıldığında tetiklenen event’i yakalar.
	// keyCode parametresi ile hangi tuşun bırakıldığı bilgisi sağlanır.
	OnKeyUp(callback func(keyCode int))

	// SetPosition, pencerenin ekran üzerindeki konumunu belirler.
	// x ve y parametreleri ile sol üst köşenin koordinatları ayarlanır.
	SetPosition(x, y int)

	// GetPosition , pencerenin mevcut ekran konumunu döndürür.
	// Dönen x ve y değerleri sol üst köşenin koordinatlarını temsil eder.
	GetPosition() (x, y int)

	// Center, pencereyi ekrana ortalar
	Center()

	// Run, pencerenin ana event-loop sürecini başlatır.
	// UI etkileşimi canlı kalır, eventler işlenir, life-cycle sürdürülür.
	Run()
}
