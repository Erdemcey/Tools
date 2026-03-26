# 🚀 Go-FastScanner: Concurrent Subdomain & Directory Fuzzer

Bu araç, **Go (Golang)** dilinin yüksek performanslı eşzamanlılık (concurrency) özelliklerini kullanarak web hedefleri üzerinde subdomain ve dizin taraması yapmanızı sağlayan hafif ve etkili bir güvenlik aracıdır.

---

## 🔥 Öne Çıkan Özellikler

* **Goroutine Performansı:** `sync.WaitGroup` ve `channels` yapısı ile yüzlerce isteği saniyeler içinde işler.
* **Esnek Şablon Sistemi:** `[[sub]]` ve `[[dir]]` yer tutucuları sayesinde özel tarama senaryoları (Subdomain + Dizin) oluşturun.
* **Akıllı Filtreleme:** Gereksiz gürültüyü önlemek için `404 Not Found` ve `400 Bad Request` yanıtlarını otomatik olarak filtreler.
* **Özelleştirilebilir Thread:** Sistem kaynaklarınıza göre iş parçacığı (thread) sayısını dinamik olarak belirleyin.

---

## 🛠️ Kurulum

Sisteminizde **Go 1.16+** yüklü olduğundan emin olun:

```bash
# Proje dizinine gidin
cd DizinTarama

# Aracı derleyin
go build -o motor motor.go

```
