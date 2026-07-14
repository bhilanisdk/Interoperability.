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

interface TimedResult {
    pageNum: number;
    status: "success" | "failure";
    value?: string;
    error?: string;
    duration: number;
}

class NodeSDKit {
    private static isLibLoaded = false;

    static {
        try {
            if (wasm) this.isLibLoaded = true;
        } catch (e) {
            console.error("Native library failed to load.");
        }
    }

    public static range(start: number, end: number): number[] {
        return Array.from({ length: end - start + 1 }, (_, i) => start + i);
    }

    public isReady(): boolean {
        return NodeSDKit.isLibLoaded;
    }

    private async fetchInteroperability(params: FetchParams): Promise<string> {
        const res = await wasm.fetch_from_js(params);
        return JSON.stringify(res);
    }

    public async fetchPages(pageRange: number[]): Promise<TimedResult[]> {
        const tasks = pageRange.map(async (page): Promise<TimedResult> => {
            if (!NodeSDKit.isLibLoaded) {
                return { pageNum: page, status: "failure", error: "Lib missing", duration: 0 };
            }

            await new Promise(r => setTimeout(r, Math.random() * 200 + 50));
            
            const start = Date.now();
            try {
                const value = await this.fetchInteroperability({ page: page.toString() });
                return { pageNum: page, status: "success", value, duration: Date.now() - start };
            } catch (e: any) {
                return { pageNum: page, status: "failure", error: e.message, duration: Date.now() - start };
            }
        });

        return Promise.all(tasks);
    }
}

async function main() {
    const sdk = new NodeSDKit();
    const totalStart = Date.now();

    console.log("--- Bhilani Interop SDK (Node Concurrency) ---");

    if (!sdk.isReady()) return;

    const results = await sdk.fetchPages(NodeSDKit.range(1, 5));

    results.forEach(({ pageNum, status, value, error, duration }) => {
        if (status === "success" && value) {
            try {
                const parsed: FetchResponse = JSON.parse(value);
                if (parsed.data.length === 0 || pageNum > parsed.pagination.total_pages) {
                    console.log(`Page ${pageNum}: Success (No Data) [${duration}ms]`);
                } else {
                    console.log(`Page ${pageNum}: Success [${duration}ms]`);
                    parsed.data.forEach(item => console.log(`  - Title: ${item.title}`));
                }
            } catch {
                console.log(`Page ${pageNum}: Success (JSON Error) [${duration}ms]`);
            }
        } else {
            console.log(`Page ${pageNum}: Failed (${error}) [${duration}ms]`);
        }
    });

    console.log(`\nTotal session duration: ${Date.now() - totalStart}ms`);
}

main();