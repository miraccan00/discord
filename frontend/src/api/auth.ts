// login authenticates against the backend and returns a JWT.
export async function login(
  username: string,
  password: string,
): Promise<{ token: string; username: string }> {
  const res = await fetch('/api/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
  if (!res.ok) {
    const body = (await res.json().catch(() => ({}))) as { error?: string };
    throw new Error(body.error ?? `login failed (${res.status})`);
  }
  return (await res.json()) as { token: string; username: string };
}
