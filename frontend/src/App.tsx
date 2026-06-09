import { useStore } from './state/store';
import { Login } from './components/Login';
import { Room } from './components/Room';

export function App() {
  const token = useStore((s) => s.auth.token);
  return token ? <Room /> : <Login />;
}
