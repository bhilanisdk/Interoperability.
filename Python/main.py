import asyncio
import json
import random
import sys
import platform
import time
from dataclasses import dataclass
from typing import List, Optional

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

@dataclass
class TimedResult:
    page_num: int
    status: str
    data: Optional[str] = None
    error: Optional[str] = None
    duration: int = 0

class PySDKit:
    _is_lib_loaded = False
    _sdk = None

    def __init__(self):
        os_name = platform.system().lower()
        arch = platform.machine().lower()
        
        try:
            # Platform support check
            is_supported_os = any(x in os_name for x in ["windows", "darwin", "linux"])
            is_supported_arch = any(x in arch for x in ["64", "amd64", "arm64", "aarch64"])

            if is_supported_os and is_supported_arch:
                import interoperability_wrapper_pyo3 as sdk
                self._sdk = sdk
                self._is_lib_loaded = True
            else:
                print(f"Unsupported platform: {os_name} ({arch}). Native features disabled.", file=sys.stderr)
        except ImportError as e:
            print(f"Native library not found: {e}", file=sys.stderr)

    def is_ready(self):
        return self._is_lib_loaded

    async def fetch_pages(self, page_range: range):
        tasks = [self._fetch_single_page(page) for page in page_range]
        return await asyncio.gather(*tasks)

    async def _fetch_single_page(self, page: int):
        if not self._is_lib_loaded:
            return TimedResult(page, "error", error="Library not loaded")
        
        await asyncio.sleep(random.uniform(0.05, 0.25))
        start_time = time.perf_counter()
        
        try:
            params = json.dumps({"page": str(page)})
            # Execute synchronous bridge call in a thread to prevent blocking
            res = await asyncio.wait_for(
                asyncio.to_thread(self._sdk.fetch_from_python, params),
                timeout=5.0
            )
            duration = int((time.perf_counter() - start_time) * 1000)
            return TimedResult(page, "success", data=res, duration=duration)
        except Exception as e:
            duration = int((time.perf_counter() - start_time) * 1000)
            return TimedResult(page, "error", error=str(e), duration=duration)

async def main():
    sdk = PySDKit()
    total_start = time.perf_counter()
    
    print("--- Bhilani Interop SDK (Python Concurrency) ---")
    
    if not sdk.is_ready():
        print("Abort: Native library not loaded for this platform.")
        return

    results = await sdk.fetch_pages(range(1, 6))

    for res in results:
        if res.status == "success":
            try:
                raw = json.loads(res.data)
                parsed = FetchResponse(
                    data=[SDKItem(title=i["title"]) for i in raw["data"]],
                    pagination=Pagination(total_pages=raw["pagination"]["total_pages"])
                )

                if not parsed.data or res.page_num > parsed.pagination.total_pages:
                    print(f"Page {res.page_num}: Success (No Data) [{res.duration}ms]")
                else:
                    print(f"Page {res.page_num}: Success [{res.duration}ms]")
                    for item in parsed.data:
                        print(f"  - Title: {item.title}")
            except Exception:
                print(f"Page {res.page_num}: Success (JSON Parsing Failed) [{res.duration}ms]")
        else:
            print(f"Page {res.page_num}: Failed ({res.error}) [{res.duration}ms]")

    total_duration = int((time.perf_counter() - total_start) * 1000)
    print(f"\nTotal session duration: {total_duration}ms")

if __name__ == "__main__":
    asyncio.run(main())