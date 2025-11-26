package platform

/*
=========================================
🖱 MouseButton Türü ve Tıklama Buton Enum
=========================================

Bu dosya, fare tıklama işlemlerinde kullanılan butonların standart bir
şekilde temsil edilmesini sağlayan `MouseButton` türünü ve ona bağlı
sabit değerleri içerir. Amaç kullanıcı etkileşiminde belirsizliği ortadan
kaldırmak, sol/sağ/orta buton ayrımlarını net bir şekilde ortaya koymak ve
pencere/arayüz katmanlarının platform bağımsız geliştirilmesine imkân
tanımaktır.

Bu tip sayesinde uygulama, tıklama işlemlerinde hangi butonun kullanıldığını
kolaylıkla algılayabilir; örneğin sol tuş seçim yapma, sağ tuş bağlam
menüsü açma, orta tuş da özel bir kontrol mekanizması için atanabilir.
Kodun ilerleyen aşamalarında input yönetimi, etkileşimli UI davranışları,
kısa yol tanımları gibi alanlarda geniş yer bulacak temel yapı taşlarından
biridir.

Kısacası, burada yazılan yalnızca birkaç sabit değil; **tüm fare tıklama
ekosisteminin üzerinde yükseldiği çekirdek yapıdır.** Grafik arayüzü olan
her proje, olay yönetimi sırasında mutlaka bu enum tipine dokunur.

@author Ahmet ALTUN
@github github.com/biyonik
@linkedin linkedin.com/in/biyonik
@email ahmet.altun60@gmail.com
*/

// MouseButton, fare düğmesini temsil eden bir tiptir.
// Bu tür int tabanlıdır ve sabitlerle birlikte kullanılarak hangi tuşa basıldığını anlamayı sağlar.
type MouseButton int

const (
	// MouseButtonLeft, farenin sol tuşuna karşılık gelir.
	// Genellikle seçim, tıklama ve sürükleme gibi temel etkileşimlerde kullanılır.
	MouseButtonLeft MouseButton = iota

	// MouseButtonRight, farenin sağ tuşuna işaret eder.
	// Çoğu arayüzde bağlam menüsü açma veya alternatif işlem tetiklemek için kullanılır.
	MouseButtonRight

	// MouseButtonMiddle, farenin orta tuşunu temsil eder (genellikle tekerlek tuşu).
	// Özel kontroller, pan/scroll fonksiyonları veya gelişmiş etkileşimlerde tercih edilir.
	MouseButtonMiddle
)
