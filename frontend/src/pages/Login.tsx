import { useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { TextInput, PasswordInput, Button, Paper, Title, Stack, Alert } from '@mantine/core';
import { useLocalStorage } from '@mantine/hooks';
import { auth } from '../api';

export default function Login() {
  const [searchParams] = useSearchParams();
  const sessionExpired = searchParams.get('session_expired') === '1';
  const [, setToken] = useLocalStorage<string | null>({ key: 'token', defaultValue: null });
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const { token } = await auth.login(email, password);
      localStorage.setItem('token', token);
      setToken(token);
      setTimeout(() => { window.location.href = '/meetings'; }, 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка входа');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Paper p="xl" maw={400} mx="auto" mt="xl">
      <Title order={2} mb="md">Вход</Title>
      {sessionExpired && (
        <Alert color="yellow" mb="md">Сессия истекла. Войдите снова.</Alert>
      )}
      <form onSubmit={handleSubmit}>
        <Stack>
          <TextInput
            label="Email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          <PasswordInput
            label="Пароль"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          {error && <span style={{ color: 'var(--mantine-color-red-6)' }}>{error}</span>}
          <Button type="submit" loading={loading}>Войти</Button>
          <Button variant="subtle" component={Link} to="/register">Нет аккаунта? Регистрация</Button>
        </Stack>
      </form>
    </Paper>
  );
}
