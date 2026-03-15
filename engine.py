import re

class RequestParser:
    @staticmethod
    def parse_to_dict(raw_text, payload_value=None, use_https=False):
        try:
            # Satır sonlarını ve boşlukları temizle
            raw_text = raw_text.strip()
            if '\n\n' in raw_text:
                header_part, body = raw_text.split('\n\n', 1)
            else:
                header_part, body = raw_text, ""
            
            lines = [line.strip() for line in header_part.splitlines() if line.strip()]
            if not lines:
                return {"error": "İstek metni boş!"}

            # İlk satır: METHOD PATH PROTOCOL
            first_line = lines[0].split()
            if len(first_line) < 2:
                return {"error": "Geçersiz istek satırı!"}
            
            method = first_line[0].upper()
            path = first_line[1]

            # Headerları sözlüğe al
            headers = {}
            for line in lines[1:]:
                if ":" in line:
                    k, v = line.split(':', 1)
                    headers[k.strip()] = v.strip()

            # Payload Yerleştirme (§...§)
            if payload_value is not None:
                path = re.sub(r"§.*?§", str(payload_value), path)
                body = re.sub(r"§.*?§", str(payload_value), body)
                for key in headers:
                    headers[key] = re.sub(r"§.*?§", str(payload_value), headers[key])

            # Protokol ve Host Belirleme
            protocol = "https" if use_https else "http"
            host = headers.get('Host', '').split()[0] # Boşluktan sonrasını at (Örn: User-Agent yapışmasını engeller)
            
            if not host:
                return {"error": "Host başlığı bulunamadı!"}
            
            # URL'yi oluştur ve temizle
            if not path.startswith('/'):
                path = '/' + path
            
            # URL içinde oluşabilecek hatalı karakterleri/satırları temizle
            full_url = f"{protocol}://{host}{path}".replace('\r', '').replace('\n', '')
            
            return {
                "method": method,
                "url": full_url,
                "headers": headers,
                "body": body,
                "payload_used": payload_value
            }
        except Exception as e:
            return {"error": f"Parser Hatası: {str(e)}"}