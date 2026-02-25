document.getElementById('setup-form').addEventListener('submit', async function(e) {
  e.preventDefault();
  const pw = document.getElementById('password').value;
  const confirm = document.getElementById('confirm').value;
  const errorEl = document.getElementById('error');
  errorEl.textContent = '';

  if (pw !== confirm) {
    errorEl.textContent = 'Passwords do not match.';
    return;
  }
  if (pw.length < 8) {
    errorEl.textContent = 'Password must be at least 8 characters.';
    return;
  }

  const btn = this.querySelector('button');
  btn.disabled = true;
  btn.textContent = 'Setting up...';

  try {
    const resp = await fetch('/api/setup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: pw }),
    });
    if (!resp.ok) {
      const text = await resp.text();
      throw new Error(text || 'Setup failed.');
    }
    window.location.href = '/';
  } catch (err) {
    errorEl.textContent = err.message;
    btn.disabled = false;
    btn.textContent = 'Set password';
  }
});
