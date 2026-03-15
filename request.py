import requests
import time
import urllib3

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

class RequestSender:
    @staticmethod
    def send(formatted_req):
        try:
            start = time.time()
            headers = formatted_req["headers"].copy()
            
            # Host başlığı URL içinde olduğu için headerdan siliyoruz 
            # (Requests kütüphanesi kendisi ekler, çiftleme hatasını önler)
            if 'Host' in headers:
                del headers['Host']
            
            # Content-Length'i requests hesaplamalı
            if 'Content-Length' in headers:
                del headers['Content-Length']

            # Body verisi
            body_bytes = formatted_req["body"].encode('utf-8') if formatted_req["body"] else None

            resp = requests.request(
                method=formatted_req["method"],
                url=formatted_req["url"],
                headers=headers,
                data=body_bytes,
                verify=False, 
                timeout=15,
                allow_redirects=False
            )
            
            return {
                "status": resp.status_code,
                "text": resp.text,
                "headers": dict(resp.headers),
                "length": len(resp.content),
                "time_ms": int((time.time() - start) * 1000),
                "payload": formatted_req.get("payload_used", "Original")
            }
        except Exception as e:
            print(f"DEBUG - Gönderilen URL: {formatted_req.get('url')}")
            print(f"DEBUG - Hata: {str(e)}")
            return {"error": f"İstek Hatası: {str(e)}", "payload": formatted_req.get("payload_used")}