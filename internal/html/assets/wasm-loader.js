(async function() {
  const go = new Go();
  try {
    // Decode base64 to get gzip-compressed data
    const compressed = Uint8Array.from(atob(window.WASM_BINARY), c => c.charCodeAt(0));

    // Decompress using DecompressionStream (modern browsers)
    let bytes;
    if (typeof DecompressionStream !== 'undefined') {
      const ds = new DecompressionStream('gzip');
      const writer = ds.writable.getWriter();
      writer.write(compressed);
      writer.close();
      const reader = ds.readable.getReader();
      const chunks = [];
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        chunks.push(value);
      }
      const totalLength = chunks.reduce((acc, chunk) => acc + chunk.length, 0);
      bytes = new Uint8Array(totalLength);
      let offset = 0;
      for (const chunk of chunks) {
        bytes.set(chunk, offset);
        offset += chunk.length;
      }
    } else {
      // Fallback: use pako if available, or show error
      if (typeof pako !== 'undefined') {
        bytes = pako.inflate(compressed);
      } else {
        throw new Error('Browser does not support DecompressionStream');
      }
    }

    const result = await WebAssembly.instantiate(bytes.buffer, go.importObject);
    go.run(result.instance);
  } catch (err) {
    // Show user-friendly error with guidance
    const indicator = document.getElementById('wasm-loading-indicator');
    if (indicator) indicator.classList.add('hidden');
    const errorContainer = document.getElementById('wasm-error-container');
    if (errorContainer) {
      errorContainer.classList.remove('hidden');
      errorContainer.innerHTML = `
        <div class="wasm-fallback">
          <p><strong>Could not load the bundle creator</strong></p>
          <p>This can happen with older browsers or certain privacy settings.</p>
          <div class="loading-error-actions">
            <button class="btn btn-primary" id="reload-page-btn">Reload page</button>
            <a href="{{GITHUB_REPO}}" class="btn btn-secondary" target="_blank">Use CLI tool instead</a>
          </div>
        </div>
      `;
      document.getElementById('reload-page-btn')?.addEventListener('click', () => window.location.reload());
    }
  }
})();
