'use client';

import { useState, useEffect } from 'react';
import { fetchDataFromWasm, FilterResponse } from '@/lib/wasmLogic';

export default function WasmInterface() {
  const [currentPage, setCurrentPage] = useState(1);
  const [resultState, setResultState] = useState<FilterResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Match Rust's total_pages from PaginationMetadata
  const totalPages = resultState?.pagination?.total_pages ?? 1;

  useEffect(() => {
    const runWasm = async () => {
      setIsLoading(true);
      try {
        const response = await fetchDataFromWasm(currentPage);
        setResultState(response);
      } catch (e) {
        console.error("WASM Fetch Error:", e);
      } finally {
        setIsLoading(false);
      }
    };

    runWasm();
  }, [currentPage]);

  return (
    <main style={{ padding: '40px', fontFamily: 'sans-serif', maxWidth: '900px' }}>
      <header>
        <h2 style={{ margin: 0 }}>React SDK (Next.js)</h2>
        <p style={{ color: '#666' }}>Rust Core via WebAssembly</p>
      </header>

      <hr style={{ margin: '20px 0' }} />

      {/* Pagination Row - Fixed Next Button logic */}
      <div style={{ display: 'flex', gap: '15px', alignItems: 'center', marginBottom: '20px' }}>
        <button
          onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
          disabled={isLoading || currentPage <= 1}
          style={{ padding: '8px 16px' }}
        >
          Previous
        </button>

        <span style={{ fontWeight: 'bold' }}>
          Page {currentPage} of {totalPages}
        </span>

        <button
          onClick={() => setCurrentPage(prev => prev + 1)}
          // Now correctly disables based on Rust's total_pages
          disabled={isLoading || currentPage >= totalPages}
          style={{ padding: '8px 16px' }}
        >
          Next
        </button>

        {isLoading && <span style={{ color: '#0070f3' }}>Loading...</span>}
      </div>

      {/* Results Display */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
        {resultState?.data.map((item) => (
          <div 
            key={item.id} 
            style={{ 
              padding: '15px', 
              border: '1px solid #ddd', 
              borderRadius: '8px',
              background: '#fff' 
            }}
          >
            <h3 style={{ margin: '0 0 5px 0' }}>{item.title}</h3>
            <p style={{ margin: 0, fontSize: '14px', color: '#555' }}>
              Integration: <strong>{item.integration}</strong> | 
              Language: <strong>{item.language}</strong>
            </p>
          </div>
        ))}
      </div>

      {/* Raw Debug View (Optional) */}
      <details style={{ marginTop: '30px' }}>
        <summary style={{ cursor: 'pointer', color: '#888' }}>View Raw JSON</summary>
        <pre style={{ 
          background: '#f4f4f4', 
          padding: '15px', 
          borderRadius: '5px', 
          fontSize: '12px',
          overflowX: 'auto' 
        }}>
          {JSON.stringify(resultState, null, 2)}
        </pre>
      </details>
    </main>
  );
}