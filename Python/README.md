# Python

**BhilaniSDK | Interoperability** by **Kantini, Chanchali**

*Get SDK Sample*

	https://github.com/bhilanisdk

*Run SDK Sample*

    python3 main.py

*Basic Usage*

    import json
    import interoperability_wrapper_pyo3 as sdk
    
    params = {
        "language": None,
        "integration": None,
        "crates": None,
        "developmentkit": None,
        "page": "1",
        "ids": None
    }
    
    def fetch_interoperability(params_dict):
        try:
            params_json = json.dumps(params_dict)
            return sdk.fetch_from_python(params_json)
        except Exception as e:
            return f"Native Interop Failed: {e}"
    
    def main():
        print("Python SDK")
        
        response = fetch_interoperability(params)
        print(response)
    
    if __name__ == "__main__":
        main()
    
*Dynamic Usage*

    import json
    from dataclasses import dataclass
    from typing import List
    
    @dataclass
    class SDKItem:
        title: str
    
    @dataclass
    class Pagination:
        total_pages: int
    
    @dataclass
    class FetchResponse:
        data: List[SDKItem]
        pagination: Pagination
    
    class PySDKit:
        def __init__(self):
            import interoperability_wrapper_pyo3 as sdk
            self.sdk = sdk
    
        def fetch_interoperability(self, params_json: str) -> str:
            return self.sdk.fetch_from_python(params_json)
    
        def fetch_page(self, page: int) -> str:
            params = json.dumps({"page": str(page)})
            return self.fetch_interoperability(params)
    
    def main():
        sdk = PySDKit()
        
        print("--- Bhilani Interop SDK ---")
    
        for page_num in range(1, 6):
            try:
                response_str = sdk.fetch_page(page_num)
                raw_json = json.loads(response_str)
    
                pagination = Pagination(total_pages=raw_json["pagination"]["total_pages"])
                data_items = [SDKItem(title=item["title"]) for item in raw_json["data"]]
                
                if not data_items or page_num > pagination.total_pages:
                    print(f"Page {page_num}: Success (No Data - Server has {pagination.total_pages} pages)")
                else:
                    print(f"Page {page_num}: Success")
                    for item in data_items:
                        print(f"  - Title: {item.title}")
    
            except Exception as e:
                print(f"Page {page_num}: Failed (Error: {e})")
    
    if __name__ == "__main__":
        main()
        
*Concurrent Usage*

    import asyncio
    import json
    import random
    from dataclasses import dataclass
    from typing import List
    
    @dataclass
    class SDKItem:
        title: str
    
    @dataclass
    class Pagination:
        total_pages: int
    
    @dataclass
    class FetchResponse:
        data: List[SDKItem]
        pagination: Pagination
    
    class PySDKit:
        def __init__(self):
            import interoperability_wrapper_pyo3 as sdk
            self.sdk = sdk
    
        async def fetch_pages(self, page_range: range):
            tasks = [self._fetch_single_page(page) for page in page_range]
            return await asyncio.gather(*tasks)
    
        async def _fetch_single_page(self, page: int):
            await asyncio.sleep(random.uniform(0.05, 0.25))
            try:
                params = json.dumps({"page": str(page)})
                res = await asyncio.wait_for(
                    asyncio.to_thread(self.sdk.fetch_from_python, params),
                    timeout=5.0
                )
                return {"status": "success", "data": res}
            except Exception as e:
                return {"status": "error", "message": str(e)}
    
    async def main():
        sdk = PySDKit()
        
        print("--- Bhilani Interop SDK (Python Concurrency) ---")
    
        results = await sdk.fetch_pages(range(1, 6))
    
        for index, result in enumerate(results):
            page_num = index + 1
            
            if result["status"] == "success":
                try:
                    raw = json.loads(result["data"])
                    parsed = FetchResponse(
                        data=[SDKItem(title=i["title"]) for i in raw["data"]],
                        pagination=Pagination(total_pages=raw["pagination"]["total_pages"])
                    )
    
                    if not parsed.data or page_num > parsed.pagination.total_pages:
                        print(f"Page {page_num}: Success (No Data - Server has {parsed.pagination.total_pages} pages)")
                    else:
                        print(f"Page {page_num}: Success")
                        for item in parsed.data:
                            print(f"  - Title: {item.title}")
                except Exception as e:
                    print(f"Page {page_num}: Success (JSON Parsing Failed: {e})")
            else:
                print(f"Page {page_num}: Failed ({result['message']})")
    
    if __name__ == "__main__":
        asyncio.run(main())

First time
<img width="877" height="440" alt="python1" src="https://github.com/bhilanisdk/media/blob/main/python1.png" />
Second time
<img width="930" height="439" alt="python2" src="https://github.com/bhilanisdk/media/blob/main/python2.png" />
Third time
<img width="842" height="439" alt="python3" src="https://github.com/bhilanisdk/media/blob/main/python3.png" />

**🙏 Mata Shabari 🙏**
