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