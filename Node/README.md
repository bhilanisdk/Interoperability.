# Node

**BhilaniSDK | Interoperability** by **Kantini, Chanchali**

*Get SDK Sample*

	https://github.com/bhilanisdk

*Run SDK Sample*

    Command 1

        npm install

    Command 2

        node index.ts

*Basic Usage*

    import * as wasm from "@rustbyshabari/interoperability-wrapper-wasm";
    
    const params = {
        language: null,
        integration: null,
        crates: null,
        developmentkit: null,
        page: "1",
        ids: null
    };
    
    const fetchInteroperability = async (payload: any) => {
        try {
            const res = await wasm.fetch_from_js(payload);
            return JSON.stringify(res, null, 2);
        } catch (error) {
            return `WASM Fetch failed: ${error}`;
        }
    };
    
    console.log("Node SDK");
    
    const response = await fetchInteroperability(params);
    console.log(response);

*Dynamic Usage*

    import * as wasm from "@rustbyshabari/interoperability-wrapper-wasm";
    
    interface FetchParams {
        page: string;
        [key: string]: any; 
    }
    
    interface Pagination {
        total_pages: number;
    }
    
    interface SDKItem {
        title: string;
    }
    
    interface FetchResponse {
        data: SDKItem[];
        pagination: Pagination;
    }
    
    class NodeSDKit {
        private async fetchInteroperability(params: FetchParams): Promise<FetchResponse> {
            const res = await wasm.fetch_from_js(params);
            return res as FetchResponse;
        }
    
        public async fetchPage(page: number): Promise<FetchResponse> {
            const params: FetchParams = { page: page.toString() };
            return await this.fetchInteroperability(params);
        }
    }
    
    async function main() {
        const sdk = new NodeSDKit();
        
        console.log("--- Bhilani Interop SDK ---");
    
        for (let pageNum = 1; pageNum <= 5; pageNum++) {
            try {
                const parsed = await sdk.fetchPage(pageNum);
                const totalPages = parsed.pagination.total_pages;
    
                if (parsed.data.length === 0 || pageNum > totalPages) {
                    console.log(`Page ${pageNum}: Success (No Data - Server has ${totalPages} pages)`);
                } else {
                    console.log(`Page ${pageNum}: Success`);
                    parsed.data.forEach(item => console.log(`  - Title: ${item.title}`));
                }
            } catch (e: any) {
                console.log(`Page ${pageNum}: Failed (Error: ${e.message})`);
            }
        }
    }
    
    main();
    
*Concurrent Usage*

    import * as wasm from "@rustbyshabari/interoperability-wrapper-wasm";
    
    interface FetchParams {
        page: string;
        [key: string]: any;
    }
    
    interface Pagination {
        total_pages: number;
    }
    
    interface SDKItem {
        title: string;
    }
    
    interface FetchResponse {
        data: SDKItem[];
        pagination: Pagination;
    }
    
    class NodeSDKit {
        private async fetchInteroperability(params: FetchParams): Promise<string> {
            const res = await wasm.fetch_from_js(params);
            return JSON.stringify(res);
        }
    
        public async fetchPages(pageRange: number[]): Promise<PromiseSettledResult<string>[]> {
            const tasks = pageRange.map(async (page) => {
                await new Promise(resolve => setTimeout(resolve, Math.random() * (251 - 50) + 50));
                
                const controller = new AbortController();
                const timeoutId = setTimeout(() => controller.abort(), 5000);
    
                try {
                    const params: FetchParams = { page: page.toString() };
                    const result = await this.fetchInteroperability(params);
                    clearTimeout(timeoutId);
                    return result;
                } catch (error) {
                    clearTimeout(timeoutId);
                    throw error;
                }
            });
    
            return Promise.allSettled(tasks);
        }
    }
    
    async function main() {
        const sdk = new NodeSDKit();
        console.log("--- Bhilani Interop SDK (TS Concurrency) ---");
    
        const pageRange = [1, 2, 3, 4, 5];
        const results = await sdk.fetchPages(pageRange);
    
        results.forEach((result, index) => {
            const pageNum = index + 1;
    
            if (result.status === "fulfilled") {
                try {
                    const parsed: FetchResponse = JSON.parse(result.value);
                    const totalPages = parsed.pagination.total_pages;
    
                    if (parsed.data.length === 0 || pageNum > totalPages) {
                        console.log(`Page ${pageNum}: Success (No Data - Server has ${totalPages} pages)`);
                    } else {
                        console.log(`Page ${pageNum}: Success`);
                        parsed.data.forEach(item => console.log(`  - Title: ${item.title}`));
                    }
                } catch (e: any) {
                    console.log(`Page ${pageNum}: Success (JSON Parsing Failed: ${e.message})`);
                }
            } else {
                console.log(`Page ${pageNum}: Failed (${result.reason?.message || result.reason})`);
            }
        });
    }
    
    main();

First time
<img width="887" height="435" alt="node1" src="https://github.com/user-attachments/assets/8187cc75-d83d-42fe-814c-5e29c23a2be7" />
Second time
<img width="802" height="438" alt="node2" src="https://github.com/user-attachments/assets/52ab3338-2a47-4b02-a5d1-d5c2af71c56e" />
Third time
<img width="905" height="435" alt="node3" src="https://github.com/user-attachments/assets/702f8851-52ae-4eda-adcb-7577e9a51e8a" />

**🙏 Mata Shabari 🙏**