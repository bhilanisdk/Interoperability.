'use client';
import dynamic from 'next/dynamic';

// This is the magic line that tells Next.js: "Don't touch this on the server!"
const WasmInterface = dynamic(() => import('../components/WasmInterface'), { 
  ssr: false 
});

export default function Home() {
  return <WasmInterface />;
}