declare module 'react-dom/client' {
  import * as React from 'react';
  
  export function createRoot(
    container: Element | DocumentFragment | null,
    options?: any
  ): {
    render(children: React.ReactNode): void;
    unmount(): void;
  };
} 