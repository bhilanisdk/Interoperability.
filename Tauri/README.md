# Tauri

**BhilaniSDK | Interoperability** by **Kantini, Chanchali**

*Get SDK Sample*

	https://github.com/bhilanisdk

*Run SDK Sample*
    
    npm install
    
    npm run tauri dev

*Usage*

    import wasmInit, { fetch_from_js } from "@rustbyshabari/interoperability-wrapper-wasm";
    
    export interface FilterResponse {
        message: string;
        data: any[]; 
        pagination: {
            current_page: number;
            items_per_page: number;
            total_pages: number;
            total_items: number;
            next_page_url?: string | null;
            prev_page_url?: string | null;
        } | null;
    }
    
    let wasmPromise: Promise<void> | null = null;
    
    async function ensureWasmInitialized() {
        if (!wasmPromise) {
            wasmPromise = wasmInit().then(() => {});
        }
        return wasmPromise;
    }
    
    const params = {
        language: null,
        integration: null,
        crates: null,
        developmentkit: null,
        page: "1",
        ids: null
    };
    
    export async function fetchDataFromWasm(pageNumber: number): Promise<FilterResponse> {
        await ensureWasmInitialized();
        
        params.page = pageNumber.toString();
    
        try {
    		const response: FilterResponse = await fetch_from_js(params);
    		return response;
        } catch (error) {
            console.error("WASM Bridge Error:", error);
            throw error;
        }
    }
    
Screenshot (Page 1)
<img width="1920" height="1017" alt="tauri1" src="https://github.com/user-attachments/assets/b6cdb603-7b6d-4ee0-a61f-d749c5d76205" />

Screenshot (Page 4)
<img width="1920" height="1015" alt="tauri2" src="https://github.com/user-attachments/assets/bf776469-e6e3-4165-93da-27e614548b64" />

**🙏 Mata Shabari 🙏**